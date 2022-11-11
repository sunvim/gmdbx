package gmdbx

/*
#include "mdbxgo.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/sunvim/gmdbx/unsafecgo"
)

var (
	NotFound = errors.New("not found")
)

type Error int32

func (e Error) Error() string {
	args := struct {
		result uintptr
		code   int32
	}{
		code: int32(e),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_strerror), ptr, 0)
	str := C.GoString((*C.char)(unsafe.Pointer(args.result)))
	return str
}

const (
	ErrSuccess     = Error(C.MDBX_SUCCESS)
	ErrResultFalse = ErrSuccess

	// ErrResultTrue Successful result with special meaning or a flag
	ErrResultTrue = Error(C.MDBX_RESULT_TRUE)

	// ErrKeyExist key/data pair already exist
	ErrKeyExist = Error(C.MDBX_KEYEXIST)

	// ErrFirstLMDBErrCode The first LMDB-compatible defined error code
	ErrFirstLMDBErrCode = ErrKeyExist

	// ErrNotFound key/data pair not found (EOF)
	ErrNotFound = Error(C.MDBX_NOTFOUND)

	// ErrPageNotFound Requested page not found -this usually indicates corruption
	ErrPageNotFound = Error(C.MDBX_PAGE_NOTFOUND)

	// ErrCorrupted Database is corrupted (page was wrong type and so on)
	ErrCorrupted = Error(C.MDBX_CORRUPTED)

	// ErrPanic Environment had fatal error, i.e. update of meta page failed and so on.
	ErrPanic = Error(C.MDBX_PANIC)

	// ErrVersionMismatch DB file version mismatch with libmdbx
	ErrVersionMismatch = Error(C.MDBX_VERSION_MISMATCH)

	// ErrInvalid File is not a valid MDBX file
	ErrInvalid = Error(C.MDBX_INVALID)

	// ErrMapFull Environment mapsize reached
	ErrMapFull = Error(C.MDBX_MAP_FULL)

	// ErrDBSFull Environment maxdbs reached
	ErrDBSFull = Error(C.MDBX_DBS_FULL)

	// ErrReadersFull Environment maxreaders reached
	ErrReadersFull = Error(C.MDBX_READERS_FULL)

	// ErrTXNFull Transaction has too many dirty pages, i.e transaction too big
	ErrTXNFull = Error(C.MDBX_TXN_FULL)

	// ErrCursorFull Cursor stack too deep -this usually indicates corruption, i.e branchC.pages loop
	ErrCursorFull = Error(C.MDBX_CURSOR_FULL)

	// ErrPageFull Page has not enough space -internal error
	ErrPageFull = Error(C.MDBX_PAGE_FULL)

	// ErrUnableExtendMapSize Database engine was unable to extend mapping, e.g. since address space
	// is unavailable or busy. This can mean:
	//  - Database size extended by other process beyond to environment mapsize
	//    and engine was unable to extend mapping while starting read
	//    transaction. Environment should be reopened to continue.
	//  - Engine was unable to extend mapping during write transaction
	//    or explicit call of ref mdbx_env_set_geometry().
	ErrUnableExtendMapSize = Error(C.MDBX_UNABLE_EXTEND_MAPSIZE)

	// ErrIncompatible Environment or database is not compatible with the requested operation
	// or the specified flags. This can mean:
	//  - The operation expects an ref MDBX_DUPSORT / ref MDBX_DUPFIXED
	//    database.
	//  - Opening a named DB when the unnamed DB has ref MDBX_DUPSORT /
	//    ref MDBX_INTEGERKEY.
	//  - Accessing a data record as a database, or vice versa.
	//  - The database was dropped and recreated with different flags.
	ErrIncompatible = Error(C.MDBX_INCOMPATIBLE)

	// ErrBadRSlot Invalid reuse of reader locktable slot
	// e.g. readC.transaction already run for current thread
	ErrBadRSlot = Error(C.MDBX_BAD_RSLOT)

	// ErrBadTXN Transaction is not valid for requested operation,
	// e.g. had errored and be must aborted, has a child, or is invalid
	ErrBadTXN = Error(C.MDBX_BAD_TXN)

	// ErrBadValSize Invalid size or alignment of key or data for target database,
	// either invalid subDB name
	ErrBadValSize = Error(C.MDBX_BAD_VALSIZE)

	// ErrBadDBI The specified DBIC.handle is invalid
	// or changed by another thread/transaction.
	ErrBadDBI = Error(C.MDBX_BAD_DBI)

	// ErrProblem Unexpected internal error, transaction should be aborted
	ErrProblem = Error(C.MDBX_PROBLEM)

	// ErrLastLMDBErrCode The last LMDBC.compatible defined error code
	ErrLastLMDBErrCode = ErrProblem

	// ErrBusy Another write transaction is running or environment is already used while
	// opening with ref MDBX_EXCLUSIVE flag
	ErrBusy              = Error(C.MDBX_BUSY)
	ErrFirstAddedErrCode = ErrBusy                 // The first of MDBXC.added error codes
	ErrEMultiVal         = Error(C.MDBX_EMULTIVAL) // The specified key has more than one associated value

	// ErrEBadSign Bad signature of a runtime object(s), this can mean:
	//  - memory corruption or doubleC.free;
	//  - ABI version mismatch (rare case);
	ErrEBadSign = Error(C.MDBX_EBADSIGN)

	// ErrWannaRecovery Database should be recovered, but this could NOT be done for now
	// since it opened in readC.only mode.
	ErrWannaRecovery = Error(C.MDBX_WANNA_RECOVERY)

	// ErrEKeyMismatch The given key value is mismatched to the current cursor position
	ErrEKeyMismatch = Error(C.MDBX_EKEYMISMATCH)

	// ErrTooLarge Database is too large for current system,
	// e.g. could NOT be mapped into RAM.
	ErrTooLarge = Error(C.MDBX_TOO_LARGE)

	// ErrThreadMismatch A thread has attempted to use a not owned object,
	// e.g. a transaction that started by another thread.
	ErrThreadMismatch = Error(C.MDBX_THREAD_MISMATCH)

	// ErrTXNOverlapping Overlapping read and write transactions for the current thread
	ErrTXNOverlapping = Error(C.MDBX_TXN_OVERLAPPING)

	ErrLastAddedErrcode = ErrTXNOverlapping

	ErrENODAT  = Error(C.MDBX_ENODATA)
	ErrEINVAL  = Error(C.MDBX_EINVAL)
	ErrEACCESS = Error(C.MDBX_EACCESS)
	ErrENOMEM  = Error(C.MDBX_ENOMEM)
	ErrEROFS   = Error(C.MDBX_EROFS)
	ErrENOSYS  = Error(C.MDBX_ENOSYS)
	ErrEIO     = Error(C.MDBX_EIO)
	ErrEPERM   = Error(C.MDBX_EPERM)
	ErrEINTR   = Error(C.MDBX_EINTR)
	ErrENOENT  = Error(C.MDBX_ENOFILE)
	ErrENOTBLK = Error(C.MDBX_EREMOTE)
)
