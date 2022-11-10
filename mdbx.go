package gmdbx

/*
#cgo !windows CFLAGS: -O2 -g -DMDBX_BUILD_FLAGS='' -DMDBX_DEBUG=0 -DNDEBUG=1 -fPIC -ffast-math -std=gnu11 -fvisibility=hidden -pthread
#cgo linux LDFLAGS: -lrt
#include "mdbxgo.h"
*/
import "C"
