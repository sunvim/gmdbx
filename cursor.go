package gmdbx

//#include "mdbxgo.h"
import "C"
import (
	"unsafe"

	"github.com/sunvim/gmdbx/unsafecgo"
)

type CursorOp int32

const (
	// CursorFirst Position at first key/data item
	CursorFirst = CursorOp(C.MDBX_FIRST)

	// CursorFirstDup ref MDBX_DUPSORT -only: Position at first data item of current key.
	CursorFirstDup = CursorOp(C.MDBX_FIRST_DUP)

	// CursorGetBoth ref MDBX_DUPSORT -only: Position at key/data pair.
	CursorGetBoth = CursorOp(C.MDBX_GET_BOTH)

	// CursorGetBothRange ref MDBX_DUPSORT -only: Position at given key and at first data greater
	// than or equal to specified data.
	CursorGetBothRange = CursorOp(C.MDBX_GET_BOTH_RANGE)

	// CursorGetCurrent Return key/data at current cursor position
	CursorGetCurrent = CursorOp(C.MDBX_GET_CURRENT)

	// CursorGetMultiple ref MDBX_DUPFIXED -only: Return up to a page of duplicate data items
	// from current cursor position. Move cursor to prepare
	// for ref MDBX_NEXT_MULTIPLE.
	CursorGetMultiple = CursorOp(C.MDBX_GET_MULTIPLE)

	// CursorLast Position at last key/data item
	CursorLast = CursorOp(C.MDBX_LAST)

	// CursorLastDup ref MDBX_DUPSORT -only: Position at last data item of current key.
	CursorLastDup = CursorOp(C.MDBX_LAST_DUP)

	// CursorNext Position at next data item
	CursorNext = CursorOp(C.MDBX_NEXT)

	// CursorNextDup ref MDBX_DUPSORT -only: Position at next data item of current key.
	CursorNextDup = CursorOp(C.MDBX_NEXT_DUP)

	// CursorNextMultiple ref MDBX_DUPFIXED -only: Return up to a page of duplicate data items
	// from next cursor position. Move cursor to prepare
	// for `MDBX_NEXT_MULTIPLE`.
	CursorNextMultiple = CursorOp(C.MDBX_NEXT_MULTIPLE)

	// CursorNextNoDup Position at first data item of next key
	CursorNextNoDup = CursorOp(C.MDBX_NEXT_NODUP)

	// CursorPrev Position at previous data item
	CursorPrev = CursorOp(C.MDBX_PREV)

	// CursorPrevDup ref MDBX_DUPSORT -only: Position at previous data item of current key.
	CursorPrevDup = CursorOp(C.MDBX_PREV_DUP)

	// CursorPrevNoDup Position at last data item of previous key
	CursorPrevNoDup = CursorOp(C.MDBX_PREV_NODUP)

	// CursorSet Position at specified key
	CursorSet = CursorOp(C.MDBX_SET)

	// CursorSetKey Position at specified key, return both key and data
	CursorSetKey = CursorOp(C.MDBX_SET_KEY)

	// CursorSetRange Position at first key greater than or equal to specified key.
	CursorSetRange = CursorOp(C.MDBX_SET_RANGE)

	// CursorPrevMultiple ref MDBX_DUPFIXED -only: Position at previous page and return up to
	// a page of duplicate data items.
	CursorPrevMultiple = CursorOp(C.MDBX_PREV_MULTIPLE)

	// CursorSetLowerBound Positions cursor at first key-value pair greater than or equal to
	// specified, return both key and data, and the return code depends on whether
	// a exact match.
	//
	// For non DUPSORT-ed collections this work the same to ref MDBX_SET_RANGE,
	// but returns ref MDBX_SUCCESS if key found exactly or
	// ref MDBX_RESULT_TRUE if greater key was found.
	//
	// For DUPSORT-ed a data value is taken into account for duplicates,
	// i.e. for a pairs/tuples of a key and an each data value of duplicates.
	// Returns ref MDBX_SUCCESS if key-value pair found exactly or
	// ref MDBX_RESULT_TRUE if the next pair was returned.///
	CursorSetLowerBound = CursorOp(C.MDBX_SET_LOWERBOUND)

	// CursorSetUpperBound Positions cursor at first key-value pair greater than specified,
	// return both key and data, and the return code depends on whether a
	// upper-bound was found.
	//
	// For non DUPSORT-ed collections this work the same to ref MDBX_SET_RANGE,
	// but returns ref MDBX_SUCCESS if the greater key was found or
	// ref MDBX_NOTFOUND otherwise.
	//
	// For DUPSORT-ed a data value is taken into account for duplicates,
	// i.e. for a pairs/tuples of a key and an each data value of duplicates.
	// Returns ref MDBX_SUCCESS if the greater pair was returned or
	// ref MDBX_NOTFOUND otherwise.
	CursorSetUpperBound = CursorOp(C.MDBX_SET_UPPERBOUND)
)

type Cursor C.MDBX_cursor

// NewCursor Create a cursor handle but not bind it to transaction nor DBI handle.
// ingroup c_cursors
//
// An capable of operation cursor is associated with a specific transaction and
// database. A cursor cannot be used when its database handle is closed. Nor
// when its transaction has ended, except with ref mdbx_cursor_bind() and
// ref mdbx_cursor_renew().
// Also it can be discarded with ref mdbx_cursor_close().
//
// A cursor must be closed explicitly always, before or after its transaction
// ends. It can be reused with ref mdbx_cursor_bind()
// or ref mdbx_cursor_renew() before finally closing it.
//
// note In contrast to LMDB, the MDBX required that any opened cursors can be
// reused and must be freed explicitly, regardless ones was opened in a
// read-only or write transaction. The REASON for this is eliminates ambiguity
// which helps to avoid errors such as: use-after-free, double-free, i.e.
// memory corruption and segfaults.
//
// param [in] context A pointer to application context to be associated with
//
//	created cursor and could be retrieved by
//	ref mdbx_cursor_get_userctx() until cursor closed.
//
// returns Created cursor handle or NULL in case out of memory.
func NewCursor() *Cursor {
	args := struct {
		context uintptr
		cursor  uintptr
	}{}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_create), ptr, 0)
	return (*Cursor)(unsafe.Pointer(args.cursor))
}
