// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

(() => {
	// Map web browser API names to their Go names.
	const webAPIs = {
		// Functions
		"crypto.getRandomValues": "crypto/rand.Read",
		"Date": "time.now",
		"performance.now": "time.now",
		"setTimeout": "time.Sleep",
		"AbortController": "context.WithCancel",

		// Interface methods
		"slice": "syscall/js.value.slice",
		"call": "syscall/js.value.call",
		"get": "syscall/js.value.get",
		"set": "syscall/js.value.set",
		"delete": "syscall/js.value.delete",
		"new": "syscall/js.value.new",
		"instanceOf": "syscall/js.value.instanceOf",
		"typeof": "syscall/js.value.type",
		"addEventListener": "syscall/js.value.addEventListener",
		"removeEventListener": "syscall/js.value.removeEventListener",
	};

	if (typeof Go === "undefined") {
		globalThis.Go = class {
			constructor() {
				this._callbackTimeouts = new Map();
				this._nextCallbackTimeoutID = 1;

				const mem = () => {
					// The buffer may change when requesting more memory.
					return new DataView(this._inst.exports.mem.buffer);
				}

				const setInt64 = (addr, v) => {
					mem().setBigInt64(addr, v, true);
				}

				const getInt64 = (addr) => {
					return mem().getBigInt64(addr, true);
				}

				const loadValue = (addr) => {
					const f = mem().getFloat64(addr, true);
					if (f === 0) {
						return undefined;
					}
					if (!isNaN(f)) {
						return f;
					}

					const id = mem().getUint32(addr, true);
					return this._values[id];
				}

				const storeValue = (addr, v) => {
					const nanHead = 0x7FF80000;

					if (typeof v === "number" && v !== 0) {
						if (isNaN(v)) {
							mem().setUint32(addr + 4, nanHead, true);
							mem().setUint32(addr, 0, true);
							return;
						}
						mem().setFloat64(addr, v, true);
						return;
					}

					if (v === undefined) {
						mem().setFloat64(addr, 0, true);
						return;
					}

					let id = this._ids.get(v);
					if (id === undefined) {
						id = this._idPool.pop();
						if (id === undefined) {
							id = this._values.length;
						}
						this._values[id] = v;
						this._goRefCounts[id] = 0;
						this._ids.set(v, id);
					}
					this._goRefCounts[id]++;

					let typeFlag = 0;
					switch (typeof v) {
						case "object":
							if (v !== null) {
								typeFlag = 1;
							}
							break;
						case "string":
							typeFlag = 2;
							break;
						case "symbol":
							typeFlag = 3;
							break;
						case "function":
							typeFlag = 4;
							break;
					}
					mem().setUint32(addr + 4, nanHead | typeFlag, true);
					mem().setUint32(addr, id, true);
				}

				const loadSlice = (addr) => {
					const array = getInt64(addr + 0);
					const len = getInt64(addr + 8);
					return new Uint8Array(this._inst.exports.mem.buffer, array, len);
				}

				const loadSliceOfValues = (addr) => {
					const array = getInt64(addr + 0);
					const len = getInt64(addr + 8);
					const a = new Array(len);
					for (let i = 0; i < len; i++) {
						a[i] = loadValue(array + i * 8);
					}
					return a;
				}

				const loadString = (addr) => {
					const saddr = getInt64(addr + 0);
					const len = getInt64(addr + 8);
					return new TextDecoder("utf-8").decode(new DataView(this._inst.exports.mem.buffer, saddr, len));
				}

				const timeOrigin = Date.now() - performance.now();
				this.importObject = {
					go: {
						// Go's SP does not change as long as no Go code is running. Some operations (e.g. calls, getters and setters)
						// may synchronously trigger a Go event handler. This makes Go code get executed in the middle of the imported
						// function. A goroutine can switch to a new stack, so we cannot assume that SP is fixed without checking
						// if we're in an event handler. We track this stack depth count by incrementing it at the beginning of every
						// event handler and decrementing it at the end.
						"runtime.wasmExit": (sp) => {
							const code = mem().getInt32(sp + 8, true);
							this.exited = true;
							delete this._inst;
							delete this._values;
							delete this._goRefCounts;
							delete this._ids;
							delete this._idPool;
							this.exit(code);
						},

						"runtime.wasmWrite": (sp) => {
							const fd = getInt64(sp + 8);
							const p = getInt64(sp + 16);
							const n = mem().getInt32(sp + 24, true);
							fs.writeSync(fd, new Uint8Array(this._inst.exports.mem.buffer, p, n));
						},

						"runtime.resetMemoryDataView": (sp) => {
							mem = () => new DataView(this._inst.exports.mem.buffer);
						},

						"runtime.nanotime1": (sp) => {
							setInt64(sp + 8, (timeOrigin + performance.now()) * 1000000);
						},

						"runtime.walltime": (sp) => {
							const sec = (timeOrigin + performance.now()) / 1000;
							setInt64(sp + 8, sec);
							mem().setInt32(sp + 16, (sec % 1) * 1000000000, true);
						},

						"runtime.scheduleTimeoutEvent": (sp) => {
							const id = this._nextCallbackTimeoutID;
							this._nextCallbackTimeoutID++;
							this._callbackTimeouts.set(id, setTimeout(
								() => {
									this._resolveCallbackPromise();
									while (this._callbackTimeouts.has(id)) {
										// setTimeout has been called again from Go, but not yet created the promise yet.
										// Let's wait for promise to be created.
										this._resolveCallbackPromise();
									}
								},
								getInt64(sp + 8) + 1, // setTimeout has been seen to fire up to 1 millisecond early
							));
							mem().setInt32(sp + 16, id, true);
						},

						"runtime.clearTimeoutEvent": (sp) => {
							const id = mem().getInt32(sp + 8, true);
							clearTimeout(this._callbackTimeouts.get(id));
							this._callbackTimeouts.delete(id);
						},

						"runtime.getRandomData": (sp) => {
							crypto.getRandomValues(loadSlice(sp + 8));
						},

						"syscall/js.stringVal": (sp) => {
							storeValue(sp + 24, loadString(sp + 8));
						},

						"syscall/js.valueGet": (sp) => {
							const result = Reflect.get(loadValue(sp + 8), loadString(sp + 16));
							sp = this._inst.exports.getsp(); // see comment above
							storeValue(sp + 32, result);
						},

						"syscall/js.valueSet": (sp) => {
							Reflect.set(loadValue(sp + 8), loadString(sp + 16), loadValue(sp + 32));
						},

						"syscall/js.valueDelete": (sp) => {
							Reflect.deleteProperty(loadValue(sp + 8), loadString(sp + 16));
						},

						"syscall/js.valueIndex": (sp) => {
							storeValue(sp + 24, Reflect.get(loadValue(sp + 8), getInt64(sp + 16)));
						},

						"syscall/js.valueSetIndex": (sp) => {
							Reflect.set(loadValue(sp + 8), getInt64(sp + 16), loadValue(sp + 24));
						},

						"syscall/js.valueCall": (sp) => {
							try {
								const v = loadValue(sp + 8);
								const m = Reflect.get(v, loadString(sp + 16));
								const args = loadSliceOfValues(sp + 32);
								const result = Reflect.apply(m, v, args);
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 56, result);
								mem().setUint8(sp + 64, 1);
							} catch (err) {
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 56, err);
								mem().setUint8(sp + 64, 0);
							}
						},

						"syscall/js.valueInvoke": (sp) => {
							try {
								const v = loadValue(sp + 8);
								const args = loadSliceOfValues(sp + 16);
								const result = Reflect.apply(v, undefined, args);
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 40, result);
								mem().setUint8(sp + 48, 1);
							} catch (err) {
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 40, err);
								mem().setUint8(sp + 48, 0);
							}
						},

						"syscall/js.valueNew": (sp) => {
							try {
								const v = loadValue(sp + 8);
								const args = loadSliceOfValues(sp + 16);
								const result = Reflect.construct(v, args);
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 40, result);
								mem().setUint8(sp + 48, 1);
							} catch (err) {
								sp = this._inst.exports.getsp(); // see comment above
								storeValue(sp + 40, err);
								mem().setUint8(sp + 48, 0);
							}
						},

						"syscall/js.valueLength": (sp) => {
							setInt64(sp + 16, parseInt(loadValue(sp + 8).length));
						},

						"syscall/js.valuePrepareString": (sp) => {
							const str = String(loadValue(sp + 8));
							const strPreEncoded = new TextEncoder("utf-8").encode(str);
							storeValue(sp + 16, strPreEncoded);
							setInt64(sp + 24, strPreEncoded.length);
						},

						"syscall/js.valueLoadString": (sp) => {
							const str = loadValue(sp + 8);
							loadSlice(sp + 16).set(str);
						},

						"syscall/js.valueInstanceOf": (sp) => {
							mem().setUint8(sp + 24, loadValue(sp + 8) instanceof loadValue(sp + 16));
						},

						"syscall/js.copyBytesToGo": (sp) => {
							const dst = loadSlice(sp + 8);
							const src = loadValue(sp + 32);
							if (!(src instanceof Uint8Array)) {
								mem().setUint8(sp + 48, 0);
								return;
							}
							const toCopy = src.subarray(0, dst.length);
							dst.set(toCopy);
							setInt64(sp + 40, toCopy.length);
							mem().setUint8(sp + 48, 1);
						},

						"syscall/js.copyBytesToJS": (sp) => {
							const dst = loadValue(sp + 8);
							const src = loadSlice(sp + 16);
							if (!(dst instanceof Uint8Array)) {
								mem().setUint8(sp + 48, 0);
								return;
							}
							const toCopy = src.subarray(0, dst.length);
							dst.set(toCopy);
							setInt64(sp + 40, toCopy.length);
							mem().setUint8(sp + 48, 1);
						},

						"debug": (value) => {
							console.log(value);
						},
					}
				};
			}

			async run(instance) {
				this._inst = instance;
				this._values = [ // JS values that Go currently has references to, indexed by reference id
					NaN,
					0,
					null,
					true,
					false,
					globalThis,
					this,
				];
				this._goRefCounts = []; // number of references that Go has to a JS value, indexed by reference id
				this._ids = new Map();  // mapping from JS values to reference ids
				this._idPool = [];      // unused ids that have been garbage collected
				this.exited = false;    // whether the Go program has exited

				// Pass command line arguments and environment variables to WebAssembly by writing them to the linear memory.
				let offset = 4096;

				const strPtr = (str) => {
					const ptr = offset;
					const bytes = new TextEncoder("utf-8").encode(str + "\0");
					new Uint8Array(this._inst.exports.mem.buffer, offset, bytes.length).set(bytes);
					offset += bytes.length;
					if (offset % 8 !== 0) {
						offset += 8 - (offset % 8);
					}
					return ptr;
				};

				const argc = this.argv.length;

				const argvPtrs = [];
				this.argv.forEach((arg) => {
					argvPtrs.push(strPtr(arg));
				});
				argvPtrs.push(0);

				const keys = Object.keys(this.env).sort();
				keys.forEach((key) => {
					argvPtrs.push(strPtr(`${key}=${this.env[key]}`));
				});
				argvPtrs.push(0);

				const argv = offset;
				argvPtrs.forEach((ptr) => {
					this._inst.exports.mem.setUint32(offset, ptr, true);
					this._inst.exports.mem.setUint32(offset + 4, 0, true);
					offset += 8;
				});

				this._inst.exports.run(argc, argv);
				if (this.exited) {
					this._resolveExitPromise();
				}
				await this._exitPromise;
			}

			_resume() {
				if (this.exited) {
					throw new Error("Go program has already exited");
				}
				this._inst.exports.resume();
				if (this.exited) {
					this._resolveExitPromise();
				}
			}

			_makeFuncWrapper(id) {
				const go = this;
				return function () {
					const event = { id: id, this: this, args: arguments };
					go._pendingEvent = event;
					go._resume();
					return event.result;
				};
			}
		}
	}
})();