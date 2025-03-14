package gmdbx

//#cgo !windows CFLAGS: -O2 -g -DMDBX_BUILD_FLAGS='' -DMDBX_DEBUG=0 -DNDEBUG=1 -fPIC -ffast-math -std=gnu11 -fvisibility=hidden -pthread
//#cgo linux LDFLAGS: -lrt
//#include "mdbxgo.h"
import "C"

import (
	"syscall"
	"unsafe"
)

type Val syscall.Iovec

func ToVal[T int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64](v T) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(&v)),
		Len:  uint64(unsafe.Sizeof(v)),
	}
}

func (v *Val) String() string {
	b := make([]byte, v.Len)
	bh := unsafe.Slice((*byte)(unsafe.Pointer(v.Base)), int(v.Len))
	copy(b, bh)
	return *(*string)(unsafe.Pointer(&b))
}

func (v *Val) UnsafeString() string {
	bh := unsafe.Slice((*byte)(unsafe.Pointer(v.Base)), int(v.Len))
	return *(*string)(unsafe.Pointer(&bh))
}

func (v *Val) Bytes() []byte {
	b := make([]byte, v.Len)
	bh := unsafe.Slice((*byte)(unsafe.Pointer(v.Base)), int(v.Len))
	copy(b, bh)
	return b
}

func (v *Val) UnsafeBytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(v.Base)), int(v.Len))
}

func (v *Val) Copy(dst []byte) []byte {
	src := v.UnsafeBytes()
	if cap(dst) >= int(v.Len) {
		dst = dst[0:v.Len]
		copy(dst, src)
		return dst
	}
	dst = make([]byte, v.Len)
	copy(dst, src)
	return dst
}

func U8(v *uint8) Val {
	return Val{
		Base: v,
		Len:  1,
	}
}

func I8(v *int8) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  1,
	}
}

func U16(v *uint16) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  2,
	}
}

func I16(v *int16) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  2,
	}
}

func U32(v *uint32) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  4,
	}
}

func I32(v *int32) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  4,
	}
}

func F32(v *float32) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  4,
	}
}

func U64(v *uint64) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  8,
	}
}

func I64(v *int64) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  8,
	}
}

func F64(v *float64) Val {
	return Val{
		Base: (*byte)(unsafe.Pointer(v)),
		Len:  8,
	}
}

func Bytes(b *[]byte) Val {
	return Val{
		Base: &(*b)[0],
		Len:  uint64(len(*b)),
	}
}

func String(s *string) Val {
	return Val{
		Base: unsafe.StringData(*s),
		Len:  uint64(len(*s)),
	}
}

// go:lint:ignore
func StringConst(s string) Val {
	return Val{
		Base: unsafe.StringData(s),
		Len:  uint64(len(s)),
	}
}

func (v *Val) I8() int8 {
	if v.Len < 1 {
		return 0
	}
	return *(*int8)(unsafe.Pointer(v.Base))
}

func (v *Val) U8() uint8 {
	if v.Len < 1 {
		return 0
	}
	return *v.Base
}

func (v *Val) I16() int16 {
	if v.Len < 2 {
		return 0
	}
	return *(*int16)(unsafe.Pointer(v.Base))
}

func (v *Val) U16() uint16 {
	if v.Len < 2 {
		return 0
	}
	return *(*uint16)(unsafe.Pointer(v.Base))
}

func (v *Val) I32() int32 {
	if v.Len < 4 {
		return 0
	}
	return *(*int32)(unsafe.Pointer(v.Base))
}

func (v *Val) U32() uint32 {
	if v.Len < 4 {
		return 0
	}
	return *(*uint32)(unsafe.Pointer(v.Base))
}

func (v *Val) I64() int64 {
	if v.Len < 8 {
		return 0
	}
	return *(*int64)(unsafe.Pointer(v.Base))
}

func (v *Val) U64() uint64 {
	if v.Len < 8 {
		return 0
	}
	return *(*uint64)(unsafe.Pointer(v.Base))
}

func (v *Val) F32() float32 {
	if v.Len < 4 {
		return 0
	}
	return *(*float32)(unsafe.Pointer(v.Base))
}

func (v *Val) F64() float64 {
	if v.Len < 8 {
		return 0
	}
	return *(*float64)(unsafe.Pointer(v.Base))
}
