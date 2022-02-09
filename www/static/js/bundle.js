/* oak build --web */
// module system
const __Oak_Modules = {};
let __Oak_Import_Aliases;
function __oak_modularize(name, fn) {
	__Oak_Modules[name] = fn;
}
function __oak_module_import(name) {
	if (typeof __Oak_Modules[name] === 'object') return __Oak_Modules[name];
	const module = __Oak_Modules[name] || __Oak_Modules[__Oak_Import_Aliases[name]];
	if (module) {
		__Oak_Modules[name] = {}; // break circular imports
		return __Oak_Modules[name] = module();
	} else {
		throw new Error(`Could not import Oak module "${name}" at runtime`);
	}
}

// language primitives
let __oak_empty_assgn_tgt;
function __oak_eq(a, b) {
	a = __as_oak_string(a);
	b = __as_oak_string(b);
	if (a === __Oak_Empty || b === __Oak_Empty) return true;

	// match either null or undefined to compare correctly against undefined ?s
	// appearing in places like optional arguments
	if (a == null && b == null) return true;
	if (a === null || b === null) return false;

	if (typeof a !== typeof b) return false;
	if (__is_oak_string(a) && __is_oak_string(b)) {
		return a.valueOf() === b.valueOf();
	}
	if (typeof a === 'number' || typeof a === 'boolean' ||
		typeof a === 'function' || typeof a === 'symbol') {
		return a === b;
	}

	// deep equality check for composite values
	if (!Array.isArray(a) && typeof a !== 'object') return false;
	if (len(a) !== len(b)) return false;
	for (const key of keys(a)) {
		if (!__oak_eq(a[key], b[key])) return false;
	}
	return true;
}
function __oak_acc(tgt, prop) {
	return __is_oak_string(tgt)
		? __as_oak_string(tgt.valueOf()[prop]) || null
		: tgt[prop] !== undefined
		? tgt[prop]
		: null;
}
function __oak_obj_key(x) {
	return typeof x === 'symbol' ? Symbol.keyFor(x) : x;
}
function __oak_push(a, b) {
	a = __as_oak_string(a);
	a.push(b);
	return a;
}
function __oak_and(a, b) {
	if (typeof a === 'boolean' && typeof b === 'boolean') {
		return a && b;
	}
	if (__is_oak_string(a) && __is_oak_string(b)) {
		const max = Math.max(a.length, b.length);
		const get = (s, i) => s.valueOf().charCodeAt(i) || 0;

		let res = '';
		for (let i = 0; i < max; i ++) {
			res += String.fromCharCode(get(a, i) & get(b, i));
		}
		return res;
	}
	return a & b;
}
function __oak_or(a, b) {
	if (typeof a === 'boolean' && typeof b === 'boolean') {
		return a || b;
	}
	if (__is_oak_string(a) && __is_oak_string(b)) {
		const max = Math.max(a.length, b.length);
		const get = (s, i) => s.valueOf().charCodeAt(i) || 0;

		let res = '';
		for (let i = 0; i < max; i ++) {
			res += String.fromCharCode(get(a, i) | get(b, i));
		}
		return res;
	}
	return a | b;
}
function __oak_xor(a, b) {
	if (typeof a === 'boolean' && typeof b === 'boolean') {
		return (a && !b) || (!a && b);
	}
	if (__is_oak_string(a) && __is_oak_string(b)) {
		const max = Math.max(a.length, b.length);
		const get = (s, i) => s.valueOf().charCodeAt(i) || 0;

		let res = '';
		for (let i = 0; i < max; i ++) {
			res += String.fromCharCode(get(a, i) ^ get(b, i));
		}
		return res;
	}
	return a ^ b;
}
const __Oak_Empty = Symbol('__Oak_Empty');

// mutable string type
function __is_oak_string(x) {
	if (x == null) return false;
	return x.__mark_oak_string;
}
function __as_oak_string(x) {
	if (typeof x === 'string') return __Oak_String(x);
	return x;
}
const __Oak_String = s => {
	if (__is_oak_string(s)) return s;
	return {
		__mark_oak_string: true,
		assign(i, slice) {
			if (i === s.length) return s += slice;
			return s = s.substr(0, i) + slice + s.substr(i + slice.length);
		},
		push(slice) {
			s += slice;
		},
		toString() {
			return s;
		},
		valueOf() {
			return s;
		},
		get length() {
			return s.length;
		},
	}
}

// tail recursion trampoline helpers
function __oak_resolve_trampoline(fn, ...args) {
	let rv = fn(...args);
	while (rv && rv.__is_oak_trampoline) {
		rv = rv.fn(...rv.args);
	}
	return rv;
}
function __oak_trampoline(fn, ...args) {
	return {
		__is_oak_trampoline: true,
		fn: fn,
		args: args,
	}
}

// env (builtin) functions

// reflection and types
const __Is_Oak_Node = typeof process === 'object';
const __Oak_Int_RE = /^[+-]?\d+$/;
function int(x) {
	x = __as_oak_string(x);
	if (typeof x === 'number') {
		// JS rounds towards higher magnitude, Oak rounds towards higher value
		const rounded = Math.floor(x);
		const diff = x - rounded;
		if (x < 0 && diff === 0.5) return rounded + 1;
		return rounded;
	}
	if (__is_oak_string(x) && __Oak_Int_RE.test(x.valueOf())) {
		const i = Number(x.valueOf());
		if (isNaN(i)) return null;
		return i;
	}
	return null;
}
function float(x) {
	x = __as_oak_string(x);
	if (typeof x === 'number') return x;
	if (__is_oak_string(x)) {
		const f = parseFloat(x.valueOf());
		if (isNaN(f)) return null;
		return f;
	}
	return null;
}
function atom(x) {
	x = __as_oak_string(x);
	if (typeof x === 'symbol' && x !== __Oak_Empty) return x;
	if (__is_oak_string(x)) return Symbol.for(x.valueOf());
	return Symbol.for(string(x));
}
function string(x) {
	x = __as_oak_string(x);
	function display(x) {
		x = __as_oak_string(x);
		if (__is_oak_string(x)) {
			return '\'' + x.valueOf().replace('\\', '\\\\').replace('\'', '\\\'') + '\'';
		} else if (typeof x === 'symbol') {
			if (x === __Oak_Empty) return '_';
			return ':' + Symbol.keyFor(x);
		}
		return string(x);
	}
	if (x === null) {
		return '?';
	} else if (typeof x === 'number') {
		return x.toString();
	} else if (__is_oak_string(x)) {
		return x;
	} else if (typeof x === 'boolean') {
		return x.toString();
	} else if (typeof x === 'function') {
		return x.toString();
	} else if (typeof x === 'symbol') {
		if (x === __Oak_Empty) return '_';
		return Symbol.keyFor(x);
	} else if (Array.isArray(x)) {
		return '[' + x.map(display).join(', ') + ']';
	} else if (typeof x === 'object') {
		const entries = [];
		for (const key of keys(x).sort()) {
			entries.push(`${key}: ${display(x[key])}`);
		}
		return '{' + entries.join(', ') + '}';
	}
	throw new Error('string() called on unknown type ' + x.toString());
}
function codepoint(c) {
	c = __as_oak_string(c);
	return c.valueOf().charCodeAt(0);
}
function char(n) {
	return String.fromCharCode(n);
}
function type(x) {
	x = __as_oak_string(x);
	if (x === null) {
		return Symbol.for('null');
	} else if (typeof x === 'number') {
		// Many discrete APIs check for :int, so we consider all integer
		// numbers :int and fall back to :float. This is not an airtight
		// solution, but works well enough and the alternative (tagged number
		// values/types) have poor perf tradeoffs.
		if (Number.isInteger(x)) return Symbol.for('int');
		return Symbol.for('float');
	} else if (__is_oak_string(x)) {
		return Symbol.for('string');
	} else if (typeof x === 'boolean') {
		return Symbol.for('bool');
	} else if (typeof x === 'symbol') {
		if (x === __Oak_Empty) return Symbol.for('empty');
		return Symbol.for('atom');
	} else if (typeof x === 'function') {
		return Symbol.for('function');
	} else if (Array.isArray(x)) {
		return Symbol.for('list');
	} else if (typeof x === 'object') {
		return Symbol.for('object');
	}
	throw new Error('type() called on unknown type ' + x.toString());
}
function len(x) {
	x = __as_oak_string(x);
	if (__is_oak_string(x)) {
		return x.length;
	} else if (Array.isArray(x)) {
		return x.length;
	} else if (typeof x === 'object' && x !== null) {
		return Object.getOwnPropertyNames(x).length;
	}
	throw new Error('len() takes a string or composite value, but got ' + string(x));
}
function keys(x) {
	if (Array.isArray(x)) {
		const k = [];
		for (let i = 0; i < x.length; i ++) k.push(i);
		return k;
	} else if (typeof x === 'object' && x !== null) {
		return Object.getOwnPropertyNames(x).map(__as_oak_string);
	}
	throw new Error('keys() takes a composite value, but got ' + string(x).valueOf());
}

// OS interfaces
function args() {
	if (__Is_Oak_Node) return process.argv.map(__as_oak_string);
	return [window.location.href];
}
function env() {
	if (__Is_Oak_Node) {
		const e = Object.assign({}, process.env);
		for (const key in e) {
			e[key] = __as_oak_string(e[key]);
		}
		return e;
	}
	return {};
}
function time() {
	return Date.now() / 1000;
}
function nanotime() {
	return int(Date.now() * 1000000);
}
function rand() {
	return Math.random();
}
let randomBytes;
function srand(length) {
	if (__Is_Oak_Node) {
		// lazily import dependency
		if (!randomBytes) randomBytes = require('crypto').randomBytes;
		return randomBytes(length).toString('latin1');
	}

	const bytes = crypto.getRandomValues(new Uint8Array(length));
	return __as_oak_string(Array.from(bytes).map(b => String.fromCharCode(b)).join(''));
}
function wait(duration, cb) {
	setTimeout(cb, duration * 1000);
	return null;
}
function exit(code) {
	if (__Is_Oak_Node) process.exit(code);
	return null;
}
function exec() {
	throw new Error('exec() not implemented');
}

// I/O
function input() {
	throw new Error('input() not implemented');
}
function print(s) {
	s = __as_oak_string(s);
	if (__Is_Oak_Node) {
		process.stdout.write(string(s).toString());
	} else {
		console.log(string(s).toString());
	}
	return s.length;
}
function ls() {
	throw new Error('ls() not implemented');
}
function rm() {
	throw new Error('rm() not implemented');
}
function mkdir() {
	throw new Error('mkdir() not implemented');
}
function stat() {
	throw new Error('stat() not implemented');
}
function open() {
	throw new Error('open() not implemented');
}
function close() {
	throw new Error('close() not implemented');
}
function read() {
	throw new Error('read() not implemented');
}
function write() {
	throw new Error('write() not implemented');
}
function listen() {
	throw new Error('listen() not implemented');
}
function req() {
	throw new Error('req() not implemented');
}

// math
function sin(n) {
	return Math.sin(n);
}
function cos(n) {
	return Math.cos(n);
}
function tan(n) {
	return Math.tan(n);
}
function asin(n) {
	return Math.asin(n);
}
function acos(n) {
	return Math.acos(n);
}
function atan(n) {
	return Math.atan(n);
}
function pow(b, n) {
	return Math.pow(b, n);
}
function log(b, n) {
	return Math.log(n) / Math.log(b);
}

// runtime
function ___runtime_lib() {
	throw new Error('___runtime_lib() not implemented');
}
function ___runtime_lib__oak_qm() {
	throw new Error('___runtime_lib?() not implemented');
}

// JavaScript interop
function bind(target, fn) {
	const fnName = Symbol.keyFor(fn);
	return target[fnName].bind(target);
}
function __oak_js_new(Constructor, ...args) {
	return new Constructor(...args);
}
(__oak_modularize(__Oak_String(``),function _(){return ((UA,contains__oak_qm,each,println,select,selectAll)=>(({println,each}=__oak_module_import(__Oak_String(`std`))),({contains__oak_qm}=__oak_module_import(__Oak_String(`str`))),(UA=((__oak_acc_tgt)=>__oak_acc_tgt.userAgent!==undefined?__oak_acc_tgt.userAgent:null)(navigator)),select=function select(selector=null){return (bind(document,Symbol.for('querySelector'))(selector))},selectAll=function selectAll(selector=null){return (bind(Array,Symbol.for('from'))(bind(document,Symbol.for('querySelectorAll'))(selector)))},each(selectAll(__Oak_String(`.try-details`)),function _(el=null){return ((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(el),false)}),((__oak_cond)=>__oak_eq(__oak_cond,contains__oak_qm(UA,__Oak_String(`Macintosh`)))?((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(select(__Oak_String(`.try-details.os-macos`))),true):__oak_eq(__oak_cond,contains__oak_qm(UA,__Oak_String(`iPhone`)))?((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(select(__Oak_String(`.try-details.os-macos`))),true):__oak_eq(__oak_cond,contains__oak_qm(UA,__Oak_String(`iPad`)))?((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(select(__Oak_String(`.try-details.os-macos`))),true):__oak_eq(__oak_cond,contains__oak_qm(UA,__Oak_String(`Linux`)))?((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(select(__Oak_String(`.try-details.os-linux`))),true):((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign(open,__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt.open):(__oak_assgn_tgt.open)=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(select(__Oak_String(`.try-details.os-other`))),true))(true),({UA,contains__oak_qm,each,println,select,selectAll})))()}),__oak_modularize(__Oak_String(`std`),function _(){return ((_asPredicate,_baseIterator,_hToN,_nToH,append,clamp,clone,compact,contains__oak_qm,debounce,__oak_js_default,each,entries,every,filter,find,first,flatten,fromHex,identity,indexOf,join,last,loop,map,merge,once,partition,println,range,reduce,reverse,slice,some,toHex,uniq,values,zip)=>(identity=function identity(x=null){return x},_baseIterator=function _baseIterator(v=null){return ((__oak_cond)=>__oak_eq(__oak_cond,Symbol.for('string'))?__Oak_String(``):__oak_eq(__oak_cond,Symbol.for('list'))?[]:__oak_eq(__oak_cond,Symbol.for('object'))?({}):null)(type(v))},_asPredicate=function _asPredicate(pred=null){return ((__oak_cond)=>__oak_eq(__oak_cond,Symbol.for('atom'))?((prop)=>((prop=string(pred)),function _(x=null){return __oak_acc(x,__oak_obj_key((prop)))}))():__oak_eq(__oak_cond,Symbol.for('string'))?function _(x=null){return __oak_acc(x,__oak_obj_key((pred)))}:__oak_eq(__oak_cond,Symbol.for('int'))?function _(x=null){return __oak_acc(x,__oak_obj_key((pred)))}:pred)(type(pred))},__oak_js_default=function __oak_js_default(x=null,base=null){return ((__oak_cond)=>__oak_eq(__oak_cond,null)?base:x)(x)},(_nToH=__Oak_String(`0123456789abcdef`)),toHex=function toHex(n=null){return ((sub)=>(sub=function sub(p=null,acc=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(p=null,acc=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__as_oak_string(__oak_acc(_nToH,__oak_obj_key((p)))+acc):__oak_trampoline(__oak_trampolined_sub,int((p/16)),__as_oak_string(__oak_acc(_nToH,__oak_obj_key(((p%16))))+acc)))((p<16))}),__oak_resolve_trampoline(__oak_trampolined_sub,p,acc)))()},sub(int(n),__Oak_String(``))))()},(_hToN=({0:0,1:1,2:2,3:3,4:4,5:5,6:6,7:7,8:8,9:9,a:10,A:10,b:11,B:11,c:12,C:12,d:13,D:13,e:14,E:14,f:15,F:15})),fromHex=function fromHex(s=null){return ((sub)=>(sub=function sub(i=null,acc=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null,acc=null){let next;return ((__oak_cond)=>__oak_eq(__oak_cond,__oak_eq(i,len(s)))?acc:__oak_eq(__oak_cond,!__oak_eq(null,(next=__oak_acc(_hToN,__oak_obj_key((__oak_acc(s,__oak_obj_key((i)))))))))?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1),__as_oak_string((acc*16)+next)):null)(true)}),__oak_resolve_trampoline(__oak_trampolined_sub,i,acc)))()},sub(0,0)))()},clamp=function clamp(min=null,max=null,n=null,m=null){return ((n=((__oak_cond)=>__oak_eq(__oak_cond,true)?min:n)((n<min))),(m=((__oak_cond)=>__oak_eq(__oak_cond,true)?min:m)((m<min))),(m=((__oak_cond)=>__oak_eq(__oak_cond,true)?max:m)((m>max))),(n=((__oak_cond)=>__oak_eq(__oak_cond,true)?m:n)((n>m))),[n,m])},slice=function slice(xs=null,min=null,max=null){return ((sub)=>((min=__oak_js_default(min,0)),(max=__oak_js_default(max,len(xs))),([min,max=null]=clamp(0,len(xs),min,max)),sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,max)?acc:__oak_trampoline(__oak_trampolined_sub,__oak_push(acc,__oak_acc(xs,__oak_obj_key((i)))),__as_oak_string(i+1)))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(_baseIterator(xs),min)))()},clone=function clone(x=null){return ((__oak_cond)=>__oak_eq(__oak_cond,Symbol.for('string'))?__as_oak_string(__Oak_String(``)+x):__oak_eq(__oak_cond,Symbol.for('list'))?slice(x):__oak_eq(__oak_cond,Symbol.for('object'))?reduce(keys(x),({}),function _(acc=null,key=null){return ((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign((key),__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt[__oak_obj_key((key))]):(__oak_assgn_tgt[__oak_obj_key((key))])=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(acc),__oak_acc(x,__oak_obj_key((key))))}):x)(type(x))},append=function append(xs=null,ys=null){return ((xlen)=>((xlen=len(xs)),each(ys,function _(y=null,i=null){return ((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign((__as_oak_string(xlen+i)),__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt[__oak_obj_key((__as_oak_string(xlen+i)))]):(__oak_assgn_tgt[__oak_obj_key((__as_oak_string(xlen+i)))])=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(xs),y)}),xs))()},join=function join(xs=null,ys=null){return append(clone(xs),ys)},range=function range(start=null,end=null,step=null){return ((step=__oak_js_default(step,1)),((__oak_cond)=>__oak_eq(__oak_cond,true)?([start,end=null]=[0,start]):null)(__oak_eq(end,null)),((__oak_cond)=>__oak_eq(__oak_cond,true)?[]:((list,sub)=>((list=[]),((__oak_cond)=>__oak_eq(__oak_cond,true)?sub=function sub(n=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(n=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?(__oak_push(list,n),__oak_trampoline(__oak_trampolined_sub,__as_oak_string(n+step))):list)((n<end))}),__oak_resolve_trampoline(__oak_trampolined_sub,n)))()}:sub=function sub(n=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(n=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?(__oak_push(list,n),__oak_trampoline(__oak_trampolined_sub,__as_oak_string(n+step))):list)((n>end))}),__oak_resolve_trampoline(__oak_trampolined_sub,n)))()})((step>0)),sub(start)))())(__oak_eq(step,0)))},reverse=function reverse(xs=null){return ((sub)=>(sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?acc:__oak_trampoline(__oak_trampolined_sub,__oak_push(acc,__oak_acc(xs,__oak_obj_key((i)))),(i-1)))((i<0))}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(_baseIterator(xs),(len(xs)-1))))()},map=function map(xs=null,f=null){return ((sub)=>((f=_asPredicate(f)),sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?acc:__oak_trampoline(__oak_trampolined_sub,__oak_push(acc,f(__oak_acc(xs,__oak_obj_key((i))),i)),__as_oak_string(i+1)))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(_baseIterator(xs),0)))()},each=function each(xs=null,f=null){return ((sub)=>(sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?null:(f(__oak_acc(xs,__oak_obj_key((i))),i),__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1))))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))()},filter=function filter(xs=null,f=null){return ((sub)=>((f=_asPredicate(f)),sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?acc:((x)=>(((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_push(acc,x):null)(f((x=__oak_acc(xs,__oak_obj_key((i)))),i)),__oak_trampoline(__oak_trampolined_sub,acc,__as_oak_string(i+1))))())(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(_baseIterator(xs),0)))()},reduce=function reduce(xs=null,seed=null,f=null){return ((sub)=>(sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?acc:__oak_trampoline(__oak_trampolined_sub,f(acc,__oak_acc(xs,__oak_obj_key((i))),i),__as_oak_string(i+1)))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(seed,0)))()},flatten=function flatten(xs=null){return reduce(xs,[],append)},compact=function compact(xs=null){return filter(xs,function _(x=null){return !__oak_eq(x,null)})},some=function some(xs=null,pred=null){return ((pred=__oak_js_default(pred,identity)),reduce(xs,false,function _(acc=null,x=null,i=null){return __oak_or(acc,pred(x,i))}))},every=function every(xs=null,pred=null){return ((pred=__oak_js_default(pred,identity)),reduce(xs,true,function _(acc=null,x=null,i=null){return __oak_and(acc,pred(x,i))}))},zip=function zip(xs=null,ys=null,zipper=null){return ((max,sub)=>((zipper=__oak_js_default(zipper,function _(x=null,y=null){return [x,y]})),(max=((__oak_cond)=>__oak_eq(__oak_cond,true)?len(xs):len(ys))((len(xs)<len(ys)))),sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,max)?acc:__oak_trampoline(__oak_trampolined_sub,__oak_push(acc,zipper(__oak_acc(xs,__oak_obj_key((i))),__oak_acc(ys,__oak_obj_key((i))),i)),__as_oak_string(i+1)))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub([],0)))()},partition=function partition(xs=null,by=null){return ((__oak_cond)=>__oak_eq(__oak_cond,Symbol.for('int'))?reduce(xs,[],function _(acc=null,x=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,0)?__oak_push(acc,[x]):(__oak_push(__oak_acc(acc,__oak_obj_key(((len(acc)-1)))),x),acc))((i%by))}):__oak_eq(__oak_cond,Symbol.for('function'))?((last)=>((last=function _(){return null}),reduce(xs,[],function _(acc=null,x=null){return ((__oak_js_this)=>(((__oak_cond)=>__oak_eq(__oak_cond,last)?__oak_push(__oak_acc(acc,__oak_obj_key(((len(acc)-1)))),x):__oak_push(acc,[x]))((__oak_js_this=by(x))),(last=__oak_js_this),acc))()})))():null)(type(by))},uniq=function uniq(xs=null,pred=null){return ((last,sub,ys)=>((pred=__oak_js_default(pred,identity)),(ys=_baseIterator(xs)),(last=function _(){return null}),sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){let p;let x;return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?ys:((__oak_cond)=>__oak_eq(__oak_cond,last)?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1)):(__oak_push(ys,x),(last=p),__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1))))((p=pred((x=__oak_acc(xs,__oak_obj_key((i))))))))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))()},first=function first(xs=null){return __oak_acc(xs,0)},last=function last(xs=null){return __oak_acc(xs,__oak_obj_key(((len(xs)-1))))},find=function find(xs=null,pred=null){return ((sub)=>(sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?-1:((__oak_cond)=>__oak_eq(__oak_cond,true)?i:__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1)))(pred(__oak_acc(xs,__oak_obj_key((i))))))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))()},indexOf=function indexOf(xs=null,x=null){return ((sub)=>(sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,len(xs))?-1:((__oak_cond)=>__oak_eq(__oak_cond,x)?i:__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1)))(__oak_acc(xs,__oak_obj_key((i)))))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))()},contains__oak_qm=function contains__oak_qm(xs=null,x=null){return (indexOf(xs,x)>-1)},values=function values(obj=null){return map(keys(obj),function _(key=null){return __oak_acc(obj,__oak_obj_key((key)))})},entries=function entries(obj=null){return map(keys(obj),function _(key=null){return [key,__oak_acc(obj,__oak_obj_key((key)))]})},merge=function merge(...os){return ((__oak_cond)=>__oak_eq(__oak_cond,0)?null:reduce(os,__oak_acc(os,0),function _(acc=null,o=null){return (reduce(keys(o),acc,function _(root=null,k=null){return ((__oak_assgn_tgt,__oak_assgn_val)=>(__is_oak_string(__oak_assgn_tgt)?__oak_assgn_tgt.assign((k),__oak_assgn_val):__oak_assgn_val===__Oak_Empty?delete (__oak_assgn_tgt[__oak_obj_key((k))]):(__oak_assgn_tgt[__oak_obj_key((k))])=__oak_assgn_val,__oak_assgn_tgt))(__as_oak_string(root),__oak_acc(o,__oak_obj_key((k))))}))}))(len(os))},once=function once(f=null){return ((called__oak_qm)=>((called__oak_qm=false),function _(...args){return ((__oak_cond)=>__oak_eq(__oak_cond,called__oak_qm)?null:((called__oak_qm=true),f(...args)))(true)}))()},loop=function loop(max=null,f=null){return ((breaker,broken,sub)=>(((__oak_cond)=>__oak_eq(__oak_cond,true)?((f=max),(max=-1)):null)(__oak_eq(f,null)),(max=__oak_js_default(max,-1)),(broken=false),breaker=function breaker(){return (broken=true)},sub=function sub(count=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(count=null){return ((__oak_cond)=>__oak_eq(__oak_cond,max)?null:((__oak_cond)=>__oak_eq(__oak_cond,true)?null:(f(count,breaker),__oak_trampoline(__oak_trampolined_sub,__as_oak_string(count+1))))(broken))(count)}),__oak_resolve_trampoline(__oak_trampolined_sub,count)))()},sub(0)))()},debounce=function debounce(duration=null,firstCall=null,f=null){return ((dargs,debounced,target,waiting__oak_qm)=>(((__oak_cond)=>__oak_eq(__oak_cond,true)?([firstCall,f=null]=[Symbol.for('trailing'),firstCall]):null)(__oak_eq(f,null)),(dargs=null),(waiting__oak_qm=false),(target=(time()-duration)),debounced=function debounced(...args){return ((tcall)=>((tcall=time()),(dargs=args),((__oak_cond)=>__oak_eq(__oak_cond,true)?((__oak_cond)=>__oak_eq(__oak_cond,true)?((target=__as_oak_string(tcall+duration)),((__oak_cond)=>__oak_eq(__oak_cond,Symbol.for('leading'))?f(...dargs):__oak_eq(__oak_cond,Symbol.for('trailing'))?((waiting__oak_qm=true),wait((target-time()),function _(){return ((waiting__oak_qm=false),f(...dargs))})):null)(firstCall)):((timeout)=>((waiting__oak_qm=true),(timeout=(target-tcall)),(target=__as_oak_string(target+duration)),wait(timeout,function _(){return ((waiting__oak_qm=false),f(...dargs))})))())((target<=tcall)):null)(!waiting__oak_qm)))()}))()},println=function println(...xs){return ((__oak_cond)=>__oak_eq(__oak_cond,0)?print(__Oak_String(`
`)):((out)=>((out=reduce(slice(xs,1),string(__oak_acc(xs,0)),function _(acc=null,x=null){return (__as_oak_string(__as_oak_string(acc+__Oak_String(` `))+string(x)))})),print(__as_oak_string(out+__Oak_String(`
`)))))())(len(xs))},({_asPredicate,_baseIterator,_hToN,_nToH,append,clamp,clone,compact,contains__oak_qm,debounce,__oak_js_default,each,entries,every,filter,find,first,flatten,fromHex,identity,indexOf,join,last,loop,map,merge,once,partition,println,range,reduce,reverse,slice,some,toHex,uniq,values,zip})))()}),__oak_modularize(__Oak_String(`str`),function _(){return ((_extend,_matchesAt__oak_qm,_replaceNonEmpty,_splitNonEmpty,_trimEndNonEmpty,_trimSpace,_trimStartNonEmpty,checkRange,contains__oak_qm,cut,__oak_js_default,digit__oak_qm,endsWith__oak_qm,indexOf,join,letter__oak_qm,lower,lower__oak_qm,padEnd,padStart,reduce,replace,slice,space__oak_qm,split,startsWith__oak_qm,trim,trimEnd,trimStart,upper,upper__oak_qm,word__oak_qm)=>(({__oak_js_default,slice,reduce}=__oak_module_import(__Oak_String(`std`))),checkRange=function checkRange(lo=null,hi=null){let checker;return checker=function checker(c=null){return ((p)=>((p=codepoint(c)),__oak_and((lo<=p),(p<=hi))))()}},upper__oak_qm=function upper__oak_qm(c=null){return __oak_and((c>=__Oak_String(`A`)),(c<=__Oak_String(`Z`)))},lower__oak_qm=function lower__oak_qm(c=null){return __oak_and((c>=__Oak_String(`a`)),(c<=__Oak_String(`z`)))},digit__oak_qm=function digit__oak_qm(c=null){return __oak_and((c>=__Oak_String(`0`)),(c<=__Oak_String(`9`)))},space__oak_qm=function space__oak_qm(c=null){return ((__oak_cond)=>__oak_eq(__oak_cond,__Oak_String(` `))?true:__oak_eq(__oak_cond,__Oak_String(`	`))?true:__oak_eq(__oak_cond,__Oak_String(`
`))?true:__oak_eq(__oak_cond,__Oak_String(``))?true:__oak_eq(__oak_cond,__Oak_String(``))?true:false)(c)},letter__oak_qm=function letter__oak_qm(c=null){return __oak_or(upper__oak_qm(c),lower__oak_qm(c))},word__oak_qm=function word__oak_qm(c=null){return __oak_or(letter__oak_qm(c),digit__oak_qm(c))},join=function join(strings=null,joiner=null){return ((joiner=__oak_js_default(joiner,__Oak_String(``))),((__oak_cond)=>__oak_eq(__oak_cond,0)?__Oak_String(``):reduce(slice(strings,1),__oak_acc(strings,0),function _(a=null,b=null){return __as_oak_string(__as_oak_string(a+joiner)+b)}))(len(strings)))},startsWith__oak_qm=function startsWith__oak_qm(s=null,prefix=null){return __oak_eq(slice(s,0,len(prefix)),prefix)},endsWith__oak_qm=function endsWith__oak_qm(s=null,suffix=null){return __oak_eq(slice(s,(len(s)-len(suffix))),suffix)},_matchesAt__oak_qm=function _matchesAt__oak_qm(s=null,substr=null,idx=null){return ((__oak_cond)=>__oak_eq(__oak_cond,0)?true:__oak_eq(__oak_cond,1)?__oak_eq(__oak_acc(s,__oak_obj_key((idx))),substr):((max,sub)=>((max=len(substr)),sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,max)?true:((__oak_cond)=>__oak_eq(__oak_cond,__oak_acc(substr,__oak_obj_key((i))))?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1)):false)(__oak_acc(s,__oak_obj_key((__as_oak_string(idx+i))))))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))())(len(substr))},indexOf=function indexOf(s=null,substr=null){return ((max,sub)=>((max=(len(s)-len(substr))),sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?i:((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+1)):-1)((i<max)))(_matchesAt__oak_qm(s,substr,i))}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},sub(0)))()},contains__oak_qm=function contains__oak_qm(s=null,substr=null){return (indexOf(s,substr)>=0)},cut=function cut(s=null,sep=null){let idx;return ((__oak_cond)=>__oak_eq(__oak_cond,-1)?[s,__Oak_String(``)]:[slice(s,0,idx),slice(s,__as_oak_string(idx+len(sep)))])((idx=indexOf(s,sep)))},lower=function lower(s=null){return reduce(s,__Oak_String(``),function _(acc=null,c=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_push(acc,char(__as_oak_string(codepoint(c)+32))):__oak_push(acc,c))(upper__oak_qm(c))})},upper=function upper(s=null){return reduce(s,__Oak_String(``),function _(acc=null,c=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_push(acc,char((codepoint(c)-32))):__oak_push(acc,c))(lower__oak_qm(c))})},_replaceNonEmpty=function _replaceNonEmpty(s=null,old=null,__oak_js_new=null){return ((lnew,lold,sub)=>((lold=len(old)),(lnew=len(__oak_js_new)),sub=function sub(acc=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(__as_oak_string(slice(acc,0,i)+__oak_js_new)+slice(acc,__as_oak_string(i+lold))),__as_oak_string(i+lnew)):((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,acc,__as_oak_string(i+1)):acc)((i<len(acc))))(_matchesAt__oak_qm(acc,old,i))}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i)))()},sub(s,0)))()},replace=function replace(s=null,old=null,__oak_js_new=null){return ((__oak_cond)=>__oak_eq(__oak_cond,__Oak_String(``))?s:_replaceNonEmpty(s,old,__oak_js_new))(old)},_splitNonEmpty=function _splitNonEmpty(s=null,sep=null){return ((coll,lsep,sub)=>((coll=[]),(lsep=len(sep)),sub=function sub(acc=null,i=null,last=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(acc=null,i=null,last=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?(__oak_push(coll,slice(acc,last,i)),__oak_trampoline(__oak_trampolined_sub,acc,__as_oak_string(i+lsep),__as_oak_string(i+lsep))):((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,acc,__as_oak_string(i+1),last):__oak_push(coll,slice(acc,last)))((i<len(acc))))(_matchesAt__oak_qm(acc,sep,i))}),__oak_resolve_trampoline(__oak_trampolined_sub,acc,i,last)))()},sub(s,0,0)))()},split=function split(s=null,sep=null){return ((__oak_cond)=>__oak_eq(__oak_cond,null)?reduce(s,[],function _(acc=null,c=null){return __oak_push(acc,c)}):__oak_eq(__oak_cond,__Oak_String(``))?reduce(s,[],function _(acc=null,c=null){return __oak_push(acc,c)}):_splitNonEmpty(s,sep))(sep)},_extend=function _extend(pad=null,n=null){return ((part,sub,times)=>((times=int((n/len(pad)))),(part=(n%len(pad))),sub=function sub(base=null,i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(base=null,i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,0)?__oak_push(base,slice(pad,0,part)):__oak_trampoline(__oak_trampolined_sub,__oak_push(base,pad),(i-1)))(i)}),__oak_resolve_trampoline(__oak_trampolined_sub,base,i)))()},sub(__Oak_String(``),times)))()},padStart=function padStart(s=null,n=null,pad=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?s:__oak_push(_extend(pad,(n-len(s))),s))((len(s)>=n))},padEnd=function padEnd(s=null,n=null,pad=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?s:__as_oak_string(s+_extend(pad,(n-len(s)))))((len(s)>=n))},_trimStartNonEmpty=function _trimStartNonEmpty(s=null,prefix=null){return ((idx,lpref,max,sub)=>((max=len(s)),(lpref=len(prefix)),sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,__as_oak_string(i+lpref)):i)(_matchesAt__oak_qm(s,prefix,i)):i)((i<max))}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},(idx=sub(0)),slice(s,idx)))()},trimStart=function trimStart(s=null,prefix=null){return ((__oak_cond)=>__oak_eq(__oak_cond,__Oak_String(``))?s:_trimStartNonEmpty(s,prefix))(prefix)},_trimEndNonEmpty=function _trimEndNonEmpty(s=null,suffix=null){return ((idx,lsuf,sub)=>((lsuf=len(suffix)),sub=function sub(i=null){return ((__oak_trampolined_sub)=>((__oak_trampolined_sub=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_sub,(i-lsuf)):i)(_matchesAt__oak_qm(s,suffix,(i-lsuf))):i)((i>-1))}),__oak_resolve_trampoline(__oak_trampolined_sub,i)))()},(idx=sub(len(s))),slice(s,0,idx)))()},trimEnd=function trimEnd(s=null,suffix=null){return ((__oak_cond)=>__oak_eq(__oak_cond,__Oak_String(``))?s:_trimEndNonEmpty(s,suffix))(suffix)},_trimSpace=function _trimSpace(s=null){return ((__oak_cond)=>__oak_eq(__oak_cond,__Oak_String(``))?s:((firstNonSpace,lastNonSpace,subEnd,subStart)=>(subStart=function subStart(i=null){return ((__oak_trampolined_subStart)=>((__oak_trampolined_subStart=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_subStart,__as_oak_string(i+1)):i)(space__oak_qm(__oak_acc(s,__oak_obj_key((i)))))}),__oak_resolve_trampoline(__oak_trampolined_subStart,i)))()},(firstNonSpace=subStart(0)),subEnd=function subEnd(i=null){return ((__oak_trampolined_subEnd)=>((__oak_trampolined_subEnd=function _(i=null){return ((__oak_cond)=>__oak_eq(__oak_cond,true)?__oak_trampoline(__oak_trampolined_subEnd,(i-1)):i)(space__oak_qm(__oak_acc(s,__oak_obj_key((i)))))}),__oak_resolve_trampoline(__oak_trampolined_subEnd,i)))()},(lastNonSpace=subEnd((len(s)-1))),slice(s,firstNonSpace,__as_oak_string(lastNonSpace+1))))())(s)},trim=function trim(s=null,part=null){return ((__oak_cond)=>__oak_eq(__oak_cond,null)?_trimSpace(s):trimStart(trimEnd(s,part),part))(part)},({_extend,_matchesAt__oak_qm,_replaceNonEmpty,_splitNonEmpty,_trimEndNonEmpty,_trimSpace,_trimStartNonEmpty,checkRange,contains__oak_qm,cut,__oak_js_default,digit__oak_qm,endsWith__oak_qm,indexOf,join,letter__oak_qm,lower,lower__oak_qm,padEnd,padStart,reduce,replace,slice,space__oak_qm,split,startsWith__oak_qm,trim,trimEnd,trimStart,upper,upper__oak_qm,word__oak_qm})))()}),(__Oak_Import_Aliases=({})),__oak_module_import(__Oak_String(``)))