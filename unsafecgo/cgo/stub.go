package cgo

/*
#include <stdio.h>
#include <time.h>
#include <unistd.h>

void mdbx_unsafecgo_stub() {}

typedef void mdbx_unsafecgo_trampoline_handler(size_t arg0, size_t arg1);

void mdbx_unsafecgo_cgo_call(size_t fn, size_t arg0, size_t arg1) {
	((mdbx_unsafecgo_trampoline_handler*)fn)(arg0, arg1);
}
*/
import "C"
import "unsafe"

var (
	Stub = C.mdbx_unsafecgo_stub
)

func NonBlocking(fn *byte, arg0, arg1 uintptr) {
	Blocking(fn, arg0, arg1)
}

func Blocking(fn *byte, arg0, arg1 uintptr) {
	C.mdbx_unsafecgo_cgo_call((C.size_t)(uintptr(unsafe.Pointer(fn))), (C.size_t)(arg0), (C.size_t)(arg1))
}
