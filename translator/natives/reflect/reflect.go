// +build js

package reflect

import (
	"github.com/gopherjs/gopherjs/js"
	"unsafe"
)

// temporary
func init() {
	a := false
	if a {
		isWrapped(nil)
		copyStruct(nil, nil, nil)
	}
}

func jsType(typ Type) js.Object {
	return js.InternalObject(typ).Get("jsType")
}

func reflectType(typ js.Object) *rtype {
	return typ.Call("reflectType").Interface().(*rtype)
}

func isWrapped(typ Type) bool {
	switch typ.Kind() {
	case Bool, Int, Int8, Int16, Int32, Uint, Uint8, Uint16, Uint32, Uintptr, Float32, Float64, Array, Map, Func, String, Struct:
		return true
	case Ptr:
		return typ.Elem().Kind() == Array
	}
	return false
}

func copyStruct(dst, src js.Object, typ Type) {
	fields := jsType(typ).Get("fields")
	for i := 0; i < fields.Length(); i++ {
		name := fields.Index(i).Index(0).Str()
		dst.Set(name, src.Get(name))
	}
}

func zeroVal(typ Type) js.Object {
	switch typ.Kind() {
	case Bool:
		return js.InternalObject(false)
	case Int, Int8, Int16, Int32, Uint, Uint8, Uint16, Uint32, Uintptr, Float32, Float64:
		return js.InternalObject(0)
	case Int64, Uint64, Complex64, Complex128:
		return jsType(typ).New(0, 0)
	case Array:
		elemType := typ.Elem()
		return js.Global.Call("go$makeNativeArray", jsType(elemType).Get("kind"), typ.Len(), func() js.Object { return zeroVal(elemType) })
	case Func:
		return js.Global.Get("go$throwNilPointerError")
	case Interface:
		return nil
	case Map:
		return js.InternalObject(false)
	case Chan, Ptr, Slice:
		return jsType(typ).Get("nil")
	case String:
		return js.InternalObject("")
	case Struct:
		return jsType(typ).Get("Ptr").New()
	default:
		panic(&ValueError{"reflect.Zero", typ.Kind()})
	}
}

func makeIword(t Type, v js.Object) iword {
	if t.Size() > ptrSize && t.Kind() != Array && t.Kind() != Struct {
		return iword(js.Global.Call("go$newDataPointer", v, jsType(PtrTo(t))).Unsafe())
	}
	return iword(v.Unsafe())
}

func makeValue(t Type, v js.Object, fl flag) Value {
	rt := t.(*rtype)
	if t.Size() > ptrSize && t.Kind() != Array && t.Kind() != Struct {
		return Value{rt, unsafe.Pointer(js.Global.Call("go$newDataPointer", v, jsType(rt.ptrTo())).Unsafe()), fl | flag(t.Kind())<<flagKindShift | flagIndir}
	}
	return Value{rt, unsafe.Pointer(v.Unsafe()), fl | flag(t.Kind())<<flagKindShift}
}

func MakeSlice(typ Type, len, cap int) Value {
	if typ.Kind() != Slice {
		panic("reflect.MakeSlice of non-slice type")
	}
	if len < 0 {
		panic("reflect.MakeSlice: negative len")
	}
	if cap < 0 {
		panic("reflect.MakeSlice: negative cap")
	}
	if len > cap {
		panic("reflect.MakeSlice: len > cap")
	}

	return makeValue(typ, jsType(typ).Call("make", len, cap, func() js.Object { return zeroVal(typ.Elem()) }), 0)
}

func jsObject() *rtype {
	return reflectType(js.Global.Get("go$packages").Get("github.com/gopherjs/gopherjs/js").Get("Object"))
}

func TypeOf(i interface{}) Type {
	if i == nil {
		return nil
	}
	c := js.InternalObject(i).Get("constructor")
	if c.Get("kind").IsUndefined() { // js.Object
		return jsObject()
	}
	return reflectType(c)
}

func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}
	c := js.InternalObject(i).Get("constructor")
	if c.Get("kind").IsUndefined() { // js.Object
		return Value{jsObject(), unsafe.Pointer(js.InternalObject(i).Unsafe()), flag(Interface) << flagKindShift}
	}
	return makeValue(reflectType(c), js.InternalObject(i).Get("go$val"), 0)
}

func arrayOf(count int, elem Type) Type {
	return reflectType(js.Global.Call("go$arrayType", jsType(elem), count))
}

func ChanOf(dir ChanDir, t Type) Type {
	return reflectType(js.Global.Call("go$chanType", jsType(t), dir == SendDir, dir == RecvDir))
}

func MapOf(key, elem Type) Type {
	switch key.Kind() {
	case Func, Map, Slice:
		panic("reflect.MapOf: invalid key type " + key.String())
	}

	return reflectType(js.Global.Call("go$mapType", jsType(key), jsType(elem)))
}

func (t *rtype) ptrTo() *rtype {
	return reflectType(js.Global.Call("go$ptrType", jsType(t)))
}

func SliceOf(t Type) Type {
	return reflectType(js.Global.Call("go$sliceType", jsType(t)))
}

func Zero(typ Type) Value {
	return Value{typ.(*rtype), unsafe.Pointer(zeroVal(typ).Unsafe()), flag(typ.Kind()) << flagKindShift}
}

func unsafe_New(typ *rtype) unsafe.Pointer {
	switch typ.Kind() {
	case Struct:
		return unsafe.Pointer(jsType(typ).Get("Ptr").New().Unsafe())
	case Array:
		return unsafe.Pointer(zeroVal(typ).Unsafe())
	default:
		return unsafe.Pointer(js.Global.Call("go$newDataPointer", zeroVal(typ), jsType(typ.ptrTo())).Unsafe())
	}
}

func makechan(typ *rtype, size uint64) (ch iword) {
	return iword(jsType(typ).New().Unsafe())
}

func chancap(ch iword) int {
	js.Global.Call("go$notSupported", "channels")
	panic("unreachable")
}

func chanclose(ch iword) {
	js.Global.Call("go$notSupported", "channels")
	panic("unreachable")
}

func chanlen(ch iword) int {
	js.Global.Call("go$notSupported", "channels")
	panic("unreachable")
}

func chanrecv(t *rtype, ch iword, nb bool) (val iword, selected, received bool) {
	js.Global.Call("go$notSupported", "channels")
	panic("unreachable")
}

func chansend(t *rtype, ch iword, val iword, nb bool) bool {
	js.Global.Call("go$notSupported", "channels")
	panic("unreachable")
}

func makemap(t *rtype) (m iword) {
	return iword(js.Global.Get("Go$Map").New().Unsafe())
}

func mapaccess(t *rtype, m iword, key iword) (val iword, ok bool) {
	k := js.InternalObject(key)
	if !k.Get("go$key").IsUndefined() {
		k = k.Call("go$key")
	}
	entry := js.InternalObject(m).Get(k.Str())
	if entry.IsUndefined() {
		return nil, false
	}
	return makeIword(t.Elem(), entry.Get("v")), true
}

func mapassign(t *rtype, m iword, key, val iword, ok bool) {
	k := js.InternalObject(key)
	if !k.Get("go$key").IsUndefined() {
		k = k.Call("go$key")
	}
	if !ok {
		js.InternalObject(m).Delete(k.Str())
		return
	}
	jsVal := js.InternalObject(val)
	if t.Elem().Kind() == Struct {
		newVal := js.Global.Get("Object").New()
		copyStruct(newVal, jsVal, t.Elem())
		jsVal = newVal
	}
	entry := js.Global.Get("Object").New()
	entry.Set("k", key)
	entry.Set("v", jsVal)
	js.InternalObject(m).Set(k.Str(), entry)
}

type mapIter struct {
	t    Type
	m    js.Object
	keys js.Object
	i    int
}

func mapiterinit(t *rtype, m iword) *byte {
	return (*byte)(unsafe.Pointer(&mapIter{t, js.InternalObject(m), js.Global.Call("go$keys", m), 0}))
}

func mapiterkey(it *byte) (key iword, ok bool) {
	iter := js.InternalObject(it)
	k := iter.Get("keys").Index(iter.Get("i").Int())
	return makeIword(iter.Get("t").Interface().(*rtype).Key(), iter.Get("m").Get(k.Str()).Get("k")), true
	// k := iter.keys.Index(iter.i)
	// return makeIword(iter.t.Key(), iter.m.G
}

func mapiternext(it *byte) {
	iter := js.InternalObject(it)
	iter.Set("i", iter.Get("i").Int()+1)
}

func maplen(m iword) int {
	return js.Global.Call("go$keys", m).Length()
}

func Copy(dst, src Value) int {
	dk := dst.kind()
	if dk != Array && dk != Slice {
		panic(&ValueError{"reflect.Copy", dk})
	}
	if dk == Array {
		dst.mustBeAssignable()
	}
	dst.mustBeExported()

	sk := src.kind()
	if sk != Array && sk != Slice {
		panic(&ValueError{"reflect.Copy", sk})
	}
	src.mustBeExported()

	typesMustMatch("reflect.Copy", dst.typ.Elem(), src.typ.Elem())

	dstVal := js.InternalObject(dst.iword())
	if dk == Array {
		dstVal = jsType(SliceOf(dst.typ.Elem())).New(dstVal)
	}

	srcVal := js.InternalObject(src.iword())
	if sk == Array {
		srcVal = jsType(SliceOf(src.typ.Elem())).New(srcVal)
	}

	return js.Global.Call("go$copySlice", dstVal, srcVal).Int()
}

func (v Value) iword() iword {
	if v.flag&flagIndir != 0 && v.typ.Kind() != Array && v.typ.Kind() != Struct {
		val := js.InternalObject(v.val).Call("go$get")
		if v.typ.Kind() == Uint64 || v.typ.Kind() == Int64 {
			val = jsType(v.typ).New(val.Get("high"), val.Get("low"))
		}
		if v.typ.Kind() == Complex64 || v.typ.Kind() == Complex128 {
			val = jsType(v.typ).New(val.Get("real"), val.Get("imag"))
		}
		return iword(val.Unsafe())
	}
	return iword(v.val)
}

func (v Value) Cap() int {
	k := v.kind()
	switch k {
	case Array:
		return v.typ.Len()
	// case Chan:
	// 	return int(chancap(v.iword()))
	case Slice:
		return js.InternalObject(v.iword()).Get("capacity").Int()
	}
	panic(&ValueError{"reflect.Value.Cap", k})
}

func (v Value) Elem() Value {
	switch k := v.kind(); k {
	case Interface:
		val := js.InternalObject(v.iword())
		if val.IsNull() {
			return Value{}
		}
		typ := reflectType(val.Get("constructor"))
		return makeValue(typ, val.Get("go$val"), v.flag&flagRO)

	case Ptr:
		if v.IsNil() {
			return Value{}
		}
		val := v.iword()
		tt := (*ptrType)(unsafe.Pointer(v.typ))
		fl := v.flag&flagRO | flagIndir | flagAddr
		fl |= flag(tt.elem.Kind()) << flagKindShift
		return Value{tt.elem, unsafe.Pointer(val), fl}

	default:
		panic(&ValueError{"reflect.Value.Elem", k})
	}
}

func (v Value) IsNil() bool {
	switch k := v.kind(); k {
	case Chan, Ptr, Slice:
		return v.iword() == iword(jsType(v.typ).Get("nil").Unsafe())
	case Func:
		return v.iword() == iword(js.Global.Get("go$throwNilPointerError").Unsafe())
	case Map:
		return v.iword() == iword(js.InternalObject(false).Unsafe())
	case Interface:
		return js.InternalObject(v.iword()).IsNull()
	default:
		panic(&ValueError{"reflect.Value.IsNil", k})
	}
}

func (v Value) Len() int {
	switch k := v.kind(); k {
	case Array, Slice, String:
		return js.InternalObject(v.iword()).Length()
	// case Chan:
	// 	return chanlen(v.iword())
	case Map:
		return js.Global.Call("go$keys", v.iword()).Length()
	default:
		panic(&ValueError{"reflect.Value.Len", k})
	}
}

func (v Value) Pointer() uintptr {
	switch k := v.kind(); k {
	case Chan, Map, Ptr, Slice, UnsafePointer:
		if v.IsNil() {
			return 0
		}
		return uintptr(unsafe.Pointer(v.iword()))
	case Func:
		if v.IsNil() {
			return 0
		}
		return 1
	default:
		panic(&ValueError{"reflect.Value.Pointer", k})
	}
}

func (v Value) SetCap(n int) {
	v.mustBeAssignable()
	v.mustBe(Slice)
	s := js.InternalObject(v.val).Call("go$get")
	if n < s.Length() || n > s.Get("capacity").Int() {
		panic("reflect: slice capacity out of range in SetCap")
	}
	newSlice := jsType(v.typ).New(s.Get("array"))
	newSlice.Set("offset", s.Get("offset"))
	newSlice.Set("length", s.Get("length"))
	newSlice.Set("capacity", n)
	js.InternalObject(v.val).Call("go$set", newSlice)
}

func (v Value) SetLen(n int) {
	v.mustBeAssignable()
	v.mustBe(Slice)
	s := js.InternalObject(v.val).Call("go$get")
	if n < 0 || n > s.Get("capacity").Int() {
		panic("reflect: slice length out of range in SetLen")
	}
	newSlice := jsType(v.typ).New(s.Get("array"))
	newSlice.Set("offset", s.Get("offset"))
	newSlice.Set("length", n)
	newSlice.Set("capacity", s.Get("capacity"))
	js.InternalObject(v.val).Call("go$set", newSlice)
}

func (v Value) Slice(i, j int) Value {
	var (
		cap int
		typ Type
		s   js.Object
	)
	switch kind := v.kind(); kind {
	case Array:
		if v.flag&flagAddr == 0 {
			panic("reflect.Value.Slice: slice of unaddressable array")
		}
		tt := (*arrayType)(unsafe.Pointer(v.typ))
		cap = int(tt.len)
		typ = SliceOf(tt.elem)
		s = jsType(typ).New(v.iword())

	case Slice:
		typ = v.typ
		s = js.InternalObject(v.iword())
		cap = s.Get("capacity").Int()

	case String:
		str := *(*string)(v.val)
		if i < 0 || j < i || j > len(str) {
			panic("reflect.Value.Slice: string slice index out of bounds")
		}
		return ValueOf(str[i:j])

	default:
		panic(&ValueError{"reflect.Value.Slice", kind})
	}

	if i < 0 || j < i || j > cap {
		panic("reflect.Value.Slice: slice index out of bounds")
	}

	return makeValue(typ, js.Global.Call("go$subslice", s, i, j), v.flag&flagRO)
}

func (v Value) Slice3(i, j, k int) Value {
	var (
		cap int
		typ Type
		s   js.Object
	)
	switch kind := v.kind(); kind {
	case Array:
		if v.flag&flagAddr == 0 {
			panic("reflect.Value.Slice: slice of unaddressable array")
		}
		tt := (*arrayType)(unsafe.Pointer(v.typ))
		cap = int(tt.len)
		typ = SliceOf(tt.elem)
		s = jsType(typ).New(v.iword())

	case Slice:
		typ = v.typ
		s = js.InternalObject(v.iword())
		cap = s.Get("capacity").Int()

	default:
		panic(&ValueError{"reflect.Value.Slice3", kind})
	}

	if i < 0 || j < i || k < j || k > cap {
		panic("reflect.Value.Slice3: slice index out of bounds")
	}

	return makeValue(typ, js.Global.Call("go$subslice", s, i, j, k), v.flag&flagRO)
}
