package gmdbx

//#include "mdbxgo.h"
import "C"
import (
	"unsafe"

	"github.com/sunvim/gmdbx/unsafecgo"
)

type TxFlags uint32

const (
	// TxReadWrite Start read-write transaction.
	//
	// Only one write transaction may be active at a time. Writes are fully
	// serialized, which guarantees that writers can never deadlock.
	TxReadWrite = TxFlags(C.MDBX_TXN_READWRITE)

	// TxReadOnly Start read-only transaction.
	//
	// There can be multiple read-only transactions simultaneously that do not
	// block each other and a write transactions.
	TxReadOnly = TxFlags(C.MDBX_TXN_RDONLY)

	// TxReadOnlyPrepare Prepare but not start read-only transaction.
	//
	// Transaction will not be started immediately, but created transaction handle
	// will be ready for use with ref mdbx_txn_renew(). This flag allows to
	// preallocate memory and assign a reader slot, thus avoiding these operations
	// at the next start of the transaction.
	TxReadOnlyPrepare = TxFlags(C.MDBX_TXN_RDONLY_PREPARE)

	// TxTry Do not block when starting a write transaction.
	TxTry = TxFlags(C.MDBX_TXN_TRY)

	// TxNoMetaSync Exactly the same as ref MDBX_NOMETASYNC,
	// but for this transaction only
	TxNoMetaSync = TxFlags(C.MDBX_TXN_NOMETASYNC)

	// TxNoSync Exactly the same as ref MDBX_SAFE_NOSYNC,
	// but for this transaction only
	TxNoSync = TxFlags(C.MDBX_TXN_NOSYNC)
)

type PutFlags uint32

const (
	// PutUpsert Upsertion by default (without any other flags)
	PutUpsert = PutFlags(C.MDBX_UPSERT)

	// PutNoOverwrite For insertion: Don't write if the key already exists.
	PutNoOverwrite = PutFlags(C.MDBX_NOOVERWRITE)

	// PutNoDupData Has effect only for ref MDBX_DUPSORT databases.
	// For upsertion: don't write if the key-value pair already exist.
	// For deletion: remove all values for key.
	PutNoDupData = PutFlags(C.MDBX_NODUPDATA)

	// PutCurrent For upsertion: overwrite the current key/data pair.
	// MDBX allows this flag for ref mdbx_put() for explicit overwrite/update
	// without insertion.
	// For deletion: remove only single entry at the current cursor position.
	PutCurrent = PutFlags(C.MDBX_CURRENT)

	// PutAllDups Has effect only for ref MDBX_DUPSORT databases.
	// For deletion: remove all multi-values (aka duplicates) for given key.
	// For upsertion: replace all multi-values for given key with a new one.
	PutAllDups = PutFlags(C.MDBX_ALLDUPS)

	// PutReserve For upsertion: Just reserve space for data, don't copy it.
	// Return a pointer to the reserved space.
	PutReserve = PutFlags(C.MDBX_RESERVE)

	// PutAppend Data is being appended.
	// Don't split full pages, continue on a new instead.
	PutAppend = PutFlags(C.MDBX_APPEND)

	// PutAppendDup Has effect only for ref MDBX_DUPSORT databases.
	// Duplicate data is being appended.
	// Don't split full pages, continue on a new instead.
	PutAppendDup = PutFlags(C.MDBX_APPENDDUP)

	// PutMultiple Only for ref MDBX_DUPFIXED.
	// Store multiple data items in one call.
	PutMultiple = PutFlags(C.MDBX_MULTIPLE)
)

type DBI uint32

type Tx struct {
	env       *Env
	txn       *C.MDBX_txn
	shared    bool
	reset     bool
	aborted   bool
	committed bool
}

func NewTransaction(env *Env) *Tx {
	txn := &Tx{}
	txn.env = env
	txn.shared = true
	return txn
}

func (tx *Tx) IsReset() bool {
	return tx.reset
}

func (tx *Tx) IsAborted() bool {
	return tx.aborted
}

func (tx *Tx) IsCommitted() bool {
	return tx.committed
}

func (env *Env) Begin(txn *Tx, flags TxFlags) Error {
	txn.env = env
	txn.txn = nil
	txn.reset = false
	txn.aborted = false
	txn.committed = false
	args := struct {
		env     uintptr
		parent  uintptr
		txn     uintptr
		context uintptr
		flags   TxFlags
		result  Error
	}{
		env:    uintptr(unsafe.Pointer(env.env)),
		parent: 0,
		txn:    uintptr(unsafe.Pointer(&txn.txn)),
		flags:  flags,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_begin_ex), ptr, 0)
	return args.result
}

// TxInfo Information about the transaction
type TxInfo struct {
	// The ID of the transaction. For a READ-ONLY transaction, this corresponds to the snapshot being read.
	ID uint64

	// For READ-ONLY transaction: the lag from a recent MVCC-snapshot, i.e. the
	// number of committed transaction since read transaction started.
	// For WRITE transaction (provided if `scan_rlt=true`): the lag of the oldest
	// reader from current transaction (i.e. at least 1 if any reader running).
	ReaderLag uint64

	// Used space by this transaction, i.e. corresponding to the last used database page.
	SpaceUsed uint64

	// Current size of database file.
	SpaceLimitSoft uint64

	// Upper bound for size the database file, i.e. the value `size_upper`
	// argument of the appropriate call of ref mdbx_env_set_geometry().
	SpaceLimitHard uint64

	// For READ-ONLY transaction: The total size of the database pages that were
	// retired by committed write transactions after the reader's MVCC-snapshot,
	// i.e. the space which would be freed after the Reader releases the
	// MVCC-snapshot for reuse by completion read transaction.
	//
	// For WRITE transaction: The summarized size of the database pages that were
	// retired for now due Copy-On-Write during this transaction.
	SpaceRetired uint64

	// For READ-ONLY transaction: the space available for writer(s) and that
	// must be exhausted for reason to call the Handle-Slow-Readers callback for
	// this read transaction.
	//
	// For WRITE transaction: the space inside transaction
	// that left to `MDBX_TXN_FULL` error.
	SpaceLeftover uint64

	// For READ-ONLY transaction (provided if `scan_rlt=true`): The space that
	// actually become available for reuse when only this transaction will be finished.
	//
	// For WRITE transaction: The summarized size of the dirty database
	// pages that generated during this transaction.
	SpaceDirty uint64
}

// Info Return information about the MDBX transaction.
// ingroup c_statinfo
//
// param [in] txn        A transaction handle returned by ref mdbx_txn_begin()
// param [out] info      The address of an ref MDBX_txn_info structure
//
//	where the information will be copied.
//
// param [in] scan_rlt   The boolean flag controls the scan of the read lock
//
//	table to provide complete information. Such scan
//	is relatively expensive and you can avoid it
//	if corresponding fields are not needed.
//	See description of ref MDBX_txn_info.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) Info(info *TxInfo) Error {
	args := struct {
		txn     uintptr
		info    uintptr
		scanRlt int32
		result  Error
	}{
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		info: uintptr(unsafe.Pointer(info)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_info), ptr, 0)
	return args.result
}

// Flags Return the transaction's flags.
// ingroup c_transactions
//
// This returns the flags associated with this transaction.
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A transaction flags, valid if input is an valid transaction,
//
//	otherwise -1.
func (tx *Tx) Flags() int32 {
	args := struct {
		txn   uintptr
		flags int32
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_flags), ptr, 0)
	return args.flags
}

// ID Return the transaction's ID.
// ingroup c_statinfo
//
// This returns the identifier associated with this transaction. For a
// read-only transaction, this corresponds to the snapshot being read;
// concurrent readers will frequently have the same transaction ID.
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A transaction ID, valid if input is an active transaction,
//
//	otherwise 0.
func (tx *Tx) ID() uint64 {
	args := struct {
		txn uintptr
		id  uint64
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_id), ptr, 0)
	return args.id
}

// CommitLatency of commit stages in 1/65536 of seconds units.
// warning This structure may be changed in future releases.
// see mdbx_txn_commit_ex()
type CommitLatency struct {
	// Duration of preparation (commit child transactions, update sub-databases records and cursors destroying).
	Preparation uint32
	// Duration of GC/freeDB handling & updation.
	GC uint32
	// Duration of internal audit if enabled.
	Audit uint32
	// Duration of writing dirty/modified data pages.
	Write uint32
	// Duration of syncing written data to the dist/storage.
	Sync uint32
	// Duration of transaction ending (releasing resources).
	Ending uint32
	// The total duration of a commit.
	Whole uint32
}

// CommitEx commit all the operations of a transaction into the database and
// collect latency information.
// see mdbx_txn_commit()
// ingroup c_statinfo
// warning This function may be changed in future releases.
func (tx *Tx) CommitEx(latency *CommitLatency) Error {
	args := struct {
		txn     uintptr
		latency uintptr
		result  Error
	}{
		txn:     uintptr(unsafe.Pointer(tx.txn)),
		latency: uintptr(unsafe.Pointer(latency)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_commit_ex), ptr, 0)
	return args.result
}

// Commit all the operations of a transaction into the database.
// ingroup c_transactions
//
// If the current thread is not eligible to manage the transaction then
// the ref MDBX_THREAD_MISMATCH error will returned. Otherwise the transaction
// will be committed and its handle is freed. If the transaction cannot
// be committed, it will be aborted with the corresponding error returned.
//
// Thus, a result other than ref MDBX_THREAD_MISMATCH means that the
// transaction is terminated:
//   - Resources are released;
//   - Transaction handle is invalid;
//   - Cursor(s) associated with transaction must not be used, except with
//     mdbx_cursor_renew() and ref mdbx_cursor_close().
//     Such cursor(s) must be closed explicitly by ref mdbx_cursor_close()
//     before or after transaction commit, either can be reused with
//     ref mdbx_cursor_renew() until it will be explicitly closed by
//     ref mdbx_cursor_close().
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_RESULT_TRUE      Transaction was aborted since it should
//
//	be aborted due to previous errors.
//
// retval MDBX_PANIC            A fatal error occurred earlier
//
//	and the environment must be shut down.
//
// retval MDBX_BAD_TXN          Transaction is already finished or never began.
// retval MDBX_EBADSIGN         Transaction object has invalid signature,
//
//	e.g. transaction was already terminated
//	or memory was corrupted.
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL           Transaction handle is NULL.
// retval MDBX_ENOSPC           No more disk space.
// retval MDBX_EIO              A system-level I/O error occurred.
// retval MDBX_ENOMEM           Out of memory.
func (tx *Tx) Commit() Error {
	tx.committed = true
	return tx.CommitEx(nil)
}

// Abort Abandon all the operations of the transaction instead of saving them.
// ingroup c_transactions
//
// The transaction handle is freed. It and its cursors must not be used again
// after this call, except with ref mdbx_cursor_renew() and
// ref mdbx_cursor_close().
//
// If the current thread is not eligible to manage the transaction then
// the ref MDBX_THREAD_MISMATCH error will returned. Otherwise the transaction
// will be aborted and its handle is freed. Thus, a result other than
// ref MDBX_THREAD_MISMATCH means that the transaction is terminated:
//   - Resources are released;
//   - Transaction handle is invalid;
//   - Cursor(s) associated with transaction must not be used, except with
//     ref mdbx_cursor_renew() and ref mdbx_cursor_close().
//     Such cursor(s) must be closed explicitly by ref mdbx_cursor_close()
//     before or after transaction abort, either can be reused with
//     ref mdbx_cursor_renew() until it will be explicitly closed by
//     ref mdbx_cursor_close().
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_PANIC            A fatal error occurred earlier and
//
//	the environment must be shut down.
//
// retval MDBX_BAD_TXN          Transaction is already finished or never began.
// retval MDBX_EBADSIGN         Transaction object has invalid signature,
//
//	e.g. transaction was already terminated
//	or memory was corrupted.
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL           Transaction handle is NULL.
func (tx *Tx) Abort() Error {
	args := struct {
		txn    uintptr
		result Error
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	tx.aborted = true
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_abort), ptr, 0)
	return args.result
}

// Break Marks transaction as broken.
// ingroup c_transactions
//
// Function keeps the transaction handle and corresponding locks, but makes
// impossible to perform any operations within a broken transaction.
// Broken transaction must then be aborted explicitly later.
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// see mdbx_txn_abort() see mdbx_txn_reset() see mdbx_txn_commit()
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) Break() Error {
	args := struct {
		txn    uintptr
		result Error
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_break), ptr, 0)
	return args.result
}

// Reset a read-only transaction.
// ingroup c_transactions
//
// Abort the read-only transaction like ref mdbx_txn_abort(), but keep the
// transaction handle. Therefore ref mdbx_txn_renew() may reuse the handle.
// This saves allocation overhead if the process will start a new read-only
// transaction soon, and also locking overhead if ref MDBX_NOTLS is in use. The
// reader table lock is released, but the table slot stays tied to its thread
// or ref MDBX_txn. Use ref mdbx_txn_abort() to discard a reset handle, and to
// free its lock table slot if ref MDBX_NOTLS is in use.
//
// Cursors opened within the transaction must not be used again after this
// call, except with ref mdbx_cursor_renew() and ref mdbx_cursor_close().
//
// Reader locks generally don't interfere with writers, but they keep old
// versions of database pages allocated. Thus they prevent the old pages from
// being reused when writers commit new data, and so under heavy load the
// database size may grow much more rapidly than otherwise.
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_PANIC            A fatal error occurred earlier and
//
//	the environment must be shut down.
//
// retval MDBX_BAD_TXN          Transaction is already finished or never began.
// retval MDBX_EBADSIGN         Transaction object has invalid signature,
//
//	e.g. transaction was already terminated
//	or memory was corrupted.
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL           Transaction handle is NULL.
func (tx *Tx) Reset() Error {
	args := struct {
		txn    uintptr
		result Error
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	tx.reset = true
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_reset), ptr, 0)
	return args.result
}

// Renew a read-only transaction.
// ingroup c_transactions
//
// This acquires a new reader lock for a transaction handle that had been
// released by ref mdbx_txn_reset(). It must be called before a reset
// transaction may be used again.
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_PANIC            A fatal error occurred earlier and
//
//	the environment must be shut down.
//
// retval MDBX_BAD_TXN          Transaction is already finished or never began.
// retval MDBX_EBADSIGN         Transaction object has invalid signature,
//
//	e.g. transaction was already terminated
//	or memory was corrupted.
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL           Transaction handle is NULL.
func (tx *Tx) Renew() Error {
	args := struct {
		txn    uintptr
		result Error
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
	}
	tx.reset = false
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_txn_renew), ptr, 0)
	return args.result
}

type Canary struct {
	X, Y, Z, V uint64
}

// PutCanary Set integers markers (aka "canary") associated with the environment.
// ingroup c_crud
// see mdbx_canary_get()
//
// param [in] txn     A transaction handle returned by ref mdbx_txn_begin()
// param [in] canary  A optional pointer to ref MDBX_canary structure for `x`,
//
//	  `y` and `z` values from.
//	- If canary is NOT NULL then the `x`, `y` and `z` values will be
//	  updated from given canary argument, but the 'v' be always set
//	  to the current transaction number if at least one `x`, `y` or
//	  `z` values have changed (i.e. if `x`, `y` and `z` have the same
//	  values as currently present then nothing will be changes or
//	  updated).
//	- if canary is NULL then the `v` value will be explicitly update
//	  to the current transaction number without changes `x`, `y` nor
//	  `z`.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) PutCanary(canary *Canary) Error {
	args := struct {
		txn    uintptr
		canary uintptr
		result Error
	}{
		txn:    uintptr(unsafe.Pointer(tx.txn)),
		canary: uintptr(unsafe.Pointer(canary)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_canary_put), ptr, 0)
	return args.result
}

// GetCanary Returns fours integers markers (aka "canary") associated with the
// environment.
// ingroup c_crud
// see mdbx_canary_set()
//
// param [in] txn     A transaction handle returned by ref mdbx_txn_begin().
// param [in] canary  The address of an MDBX_canary structure where the
//
//	information will be copied.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) GetCanary(canary *Canary) Error {
	args := struct {
		txn    uintptr
		canary uintptr
		result Error
	}{
		txn:    uintptr(unsafe.Pointer(tx.txn)),
		canary: uintptr(unsafe.Pointer(canary)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_canary_get), ptr, 0)
	return args.result
}

// EnvInfo Return information about the MDBX environment.
// ingroup c_statinfo
//
// At least one of env or txn argument must be non-null. If txn is passed
// non-null then stat will be filled accordingly to the given transaction.
// Otherwise, if txn is null, then stat will be populated by a snapshot from
// the last committed write transaction, and at next time, other information
// can be returned.
//
// Legacy ref mdbx_env_info() correspond to calling ref mdbx_env_info_ex()
// with the null `txn` argument.
//
// param [in] env     An environment handle returned by ref mdbx_env_create()
// param [in] txn     A transaction handle returned by ref mdbx_txn_begin()
// param [out] info   The address of an ref MDBX_envinfo structure
//
//	where the information will be copied
//
// param [in] bytes   The size of ref MDBX_envinfo.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) EnvInfo(info *EnvInfo) Error {
	if info == nil {
		return ErrInvalid
	}
	args := struct {
		env    uintptr
		txn    uintptr
		info   uintptr
		size   uintptr
		result int32
	}{
		env:  uintptr(unsafe.Pointer(tx.env.env)),
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		info: uintptr(unsafe.Pointer(info)),
		size: unsafe.Sizeof(C.MDBX_envinfo{}),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_env_info_ex), ptr, 0)
	return Error(args.result)
}

// OpenDBI Open or Create a database in the environment.
// ingroup c_dbi
//
// A database handle denotes the name and parameters of a database,
// independently of whether such a database exists. The database handle may be
// discarded by calling ref mdbx_dbi_close(). The old database handle is
// returned if the database was already open. The handle may only be closed
// once.
//
// note A notable difference between MDBX and LMDB is that MDBX make handles
// opened for existing databases immediately available for other transactions,
// regardless this transaction will be aborted or reset. The REASON for this is
// to avoiding the requirement for multiple opening a same handles in
// concurrent read transactions, and tracking of such open but hidden handles
// until the completion of read transactions which opened them.
//
// Nevertheless, the handle for the NEWLY CREATED database will be invisible
// for other transactions until the this write transaction is successfully
// committed. If the write transaction is aborted the handle will be closed
// automatically. After a successful commit the such handle will reside in the
// shared environment, and may be used by other transactions.
//
// In contrast to LMDB, the MDBX allow this function to be called from multiple
// concurrent transactions or threads in the same process.
//
// To use named database (with name != NULL), ref mdbx_env_set_maxdbs()
// must be called before opening the environment. Table names are
// keys in the internal unnamed database, and may be read but not written.
//
// param [in] txn    transaction handle returned by ref mdbx_txn_begin().
// param [in] name   The name of the database to open. If only a single
//
//	database is needed in the environment,
//	this value may be NULL.
//
// param [in] flags  Special options for this database. This parameter must
//
//	                  be set to 0 or by bitwise OR'ing together one or more
//	                  of the values described here:
//	- ref MDBX_REVERSEKEY
//	    Keys are strings to be compared in reverse order, from the end
//	    of the strings to the beginning. By default, Keys are treated as
//	    strings and compared from beginning to end.
//	- ref MDBX_INTEGERKEY
//	    Keys are binary integers in native byte order, either uint32_t or
//	    uint64_t, and will be sorted as such. The keys must all be of the
//	    same size and must be aligned while passing as arguments.
//	- ref MDBX_DUPSORT
//	    Duplicate keys may be used in the database. Or, from another point of
//	    view, keys may have multiple data items, stored in sorted order. By
//	    default keys must be unique and may have only a single data item.
//	- ref MDBX_DUPFIXED
//	    This flag may only be used in combination with ref MDBX_DUPSORT. This
//	    option tells the library that the data items for this database are
//	    all the same size, which allows further optimizations in storage and
//	    retrieval. When all data items are the same size, the
//	    ref MDBX_GET_MULTIPLE, ref MDBX_NEXT_MULTIPLE and
//	    ref MDBX_PREV_MULTIPLE cursor operations may be used to retrieve
//	    multiple items at once.
//	- ref MDBX_INTEGERDUP
//	    This option specifies that duplicate data items are binary integers,
//	    similar to ref MDBX_INTEGERKEY keys. The data values must all be of the
//	    same size and must be aligned while passing as arguments.
//	- ref MDBX_REVERSEDUP
//	    This option specifies that duplicate data items should be compared as
//	    strings in reverse order (the comparison is performed in the direction
//	    from the last byte to the first).
//	- ref MDBX_CREATE
//	    Create the named database if it doesn't exist. This option is not
//	    allowed in a read-only transaction or a read-only environment.
//
// param [out] dbi     Address where the new ref MDBX_dbi handle
//
//	will be stored.
//
// For ref mdbx_dbi_open_ex() additional arguments allow you to set custom
// comparison functions for keys and values (for multimaps).
// see avoid_custom_comparators
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_NOTFOUND   The specified database doesn't exist in the
//
//	environment and ref MDBX_CREATE was not specified.
//
// retval MDBX_DBS_FULL   Too many databases have been opened.
//
//	see mdbx_env_set_maxdbs()
//
// retval MDBX_INCOMPATIBLE  Database is incompatible with given flags,
//
//	i.e. the passed flags is different with which the
//	database was created, or the database was already
//	opened with a different comparison function(s).
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
func (tx *Tx) OpenDBI(name string, flags DBFlags) (DBI, Error) {
	if len(name) == 0 {
		var dbi DBI
		err := Error(C.mdbx_dbi_open(tx.txn, nil, (C.MDBX_db_flags_t)(flags), (*C.MDBX_dbi)(unsafe.Pointer(&dbi))))
		return dbi, err
	} else {
		n := C.CString(name)
		defer C.free(unsafe.Pointer(n))
		var dbi DBI
		err := Error(C.mdbx_dbi_open(tx.txn, n, (C.MDBX_db_flags_t)(flags), (*C.MDBX_dbi)(unsafe.Pointer(&dbi))))
		return dbi, err
	}
}

type Stats struct {
	PageSize      uint32 // Size of a database page. This is the same for all databases.
	Depth         uint32 // Depth (height) of the B-tree
	BranchPages   uint64 // Number of internal (non-leaf) pages
	LeafPages     uint64 // Number of leaf pages
	OverflowPages uint64 // Number of overflow pages
	Entries       uint64 // Number of data items
	ModTxnID      uint64 // Transaction ID of committed last modification
}

// DBIStat Retrieve statistics for a database.
// ingroup c_statinfo
//
// param [in] txn     A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi     A database handle returned by ref mdbx_dbi_open().
// param [out] stat   The address of an ref MDBX_stat structure where
//
//	the statistics will be copied.
//
// param [in] bytes   The size of ref MDBX_stat.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL   An invalid parameter was specified.
func (tx *Tx) DBIStat(dbi DBI, stat *Stats) Error {
	args := struct {
		txn    uintptr
		stat   uintptr
		size   uintptr
		dbi    uint32
		result Error
	}{
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		stat: uintptr(unsafe.Pointer(stat)),
		size: unsafe.Sizeof(Stats{}),
		dbi:  uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_dbi_stat), ptr, 0)
	return args.result
}

// DBIFlags Retrieve the DB flags and status for a database handle.
// ingroup c_statinfo
//
// param [in] txn     A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi     A database handle returned by ref mdbx_dbi_open().
// param [out] flags  Address where the flags will be returned.
// param [out] state  Address where the state will be returned.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) DBIFlags(dbi DBI) (DBFlags, DBIState, Error) {
	var flags DBFlags
	var state DBIState

	args := struct {
		txn    uintptr
		flags  uintptr
		state  uintptr
		dbi    uint32
		result Error
	}{
		txn:   uintptr(unsafe.Pointer(tx.txn)),
		flags: uintptr(unsafe.Pointer(&flags)),
		state: uintptr(unsafe.Pointer(&state)),
		dbi:   uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_dbi_flags_ex), ptr, 0)
	return flags, state, args.result
}

// Drop Empty or delete and close a database.
// ingroup c_crud
//
// see mdbx_dbi_close() see mdbx_dbi_open()
//
// param [in] txn  A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi  A database handle returned by ref mdbx_dbi_open().
// param [in] del  `false` to empty the DB, `true` to delete it
//
//	from the environment and close the DB handle.
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) Drop(dbi DBI, del bool) Error {
	args := struct {
		txn    uintptr
		del    uintptr
		dbi    uint32
		result Error
	}{
		txn: uintptr(unsafe.Pointer(tx.txn)),
		dbi: uint32(dbi),
	}
	if del {
		args.del = 1
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_drop), ptr, 0)
	return args.result
}

// Get items from a database.
// ingroup c_crud
//
// This function retrieves key/data pairs from the database. The address
// and length of the data associated with the specified key are returned
// in the structure to which data refers.
// If the database supports duplicate keys (ref MDBX_DUPSORT) then the
// first data item for the key will be returned. Retrieval of other
// items requires the use of ref mdbx_cursor_get().
//
// note The memory pointed to by the returned values is owned by the
// database. The caller need not dispose of the memory, and may not
// modify it in any way. For values returned in a read-only transaction
// any modification attempts will cause a `SIGSEGV`.
//
// note Values returned from the database are valid only until a
// subsequent update operation, or the end of the transaction.
//
// param [in] txn       A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi       A database handle returned by ref mdbx_dbi_open().
// param [in] key       The key to search for in the database.
// param [in,out] data  The data corresponding to the key.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_NOTFOUND  The key was not in the database.
// retval MDBX_EINVAL    An invalid parameter was specified.
func (tx *Tx) Get(dbi DBI, key *Val, data *Val) Error {
	args := struct {
		txn    uintptr
		key    uintptr
		data   uintptr
		dbi    uint32
		result Error
	}{
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		key:  uintptr(unsafe.Pointer(key)),
		data: uintptr(unsafe.Pointer(data)),
		dbi:  uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_get), ptr, 0)
	return args.result
}

// GetEqualOrGreat Get equal or great item from a database.
// ingroup c_crud
//
// Briefly this function does the same as ref mdbx_get() with a few
// differences:
//  1. Return equal or great (due comparison function) key-value
//     pair, but not only exactly matching with the key.
//  2. On success return ref MDBX_SUCCESS if key found exactly,
//     and ref MDBX_RESULT_TRUE otherwise. Moreover, for databases with
//     ref MDBX_DUPSORT flag the data argument also will be used to match over
//     multi-value/duplicates, and ref MDBX_SUCCESS will be returned only when
//     BOTH the key and the data match exactly.
//  3. Updates BOTH the key and the data for pointing to the actual key-value
//     pair inside the database.
//
// param [in] txn           A transaction handle returned
//
//	by ref mdbx_txn_begin().
//
// param [in] dbi           A database handle returned by ref mdbx_dbi_open().
// param [in,out] key       The key to search for in the database.
// param [in,out] data      The data corresponding to the key.
//
// returns A non-zero error value on failure and ref MDBX_RESULT_FALSE
//
//	or ref MDBX_RESULT_TRUE on success (as described above).
//	Some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_NOTFOUND      The key was not in the database.
// retval MDBX_EINVAL        An invalid parameter was specified.
func (tx *Tx) GetEqualOrGreat(dbi DBI, key *Val, data *Val) Error {
	args := struct {
		txn    uintptr
		key    uintptr
		data   uintptr
		dbi    uint32
		result Error
	}{
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		key:  uintptr(unsafe.Pointer(key)),
		data: uintptr(unsafe.Pointer(data)),
		dbi:  uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_get_equal_or_great), ptr, 0)
	return args.result
}

// GetEx Get items from a database
// and optionally number of data items for a given key.
//
// ingroup c_crud
//
// Briefly this function does the same as ref mdbx_get() with a few
// differences:
//  1. If values_count is NOT NULL, then returns the count
//     of multi-values/duplicates for a given key.
//  2. Updates BOTH the key and the data for pointing to the actual key-value
//     pair inside the database.
//
// param [in] txn           A transaction handle returned
//
//	by ref mdbx_txn_begin().
//
// param [in] dbi           A database handle returned by ref mdbx_dbi_open().
// param [in,out] key       The key to search for in the database.
// param [in,out] data      The data corresponding to the key.
// param [out] values_count The optional address to return number of values
//
//	associated with given key:
//	 = 0 - in case ref MDBX_NOTFOUND error;
//	 = 1 - exactly for databases
//	       WITHOUT ref MDBX_DUPSORT;
//	 >= 1 for databases WITH ref MDBX_DUPSORT.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_NOTFOUND  The key was not in the database.
// retval MDBX_EINVAL    An invalid parameter was specified.
func (tx *Tx) GetEx(dbi DBI, key *Val, data *Val) (int, Error) {
	var valuesCount uintptr
	args := struct {
		txn         uintptr
		key         uintptr
		data        uintptr
		valuesCount uintptr
		dbi         uint32
		result      Error
	}{
		txn:         uintptr(unsafe.Pointer(tx.txn)),
		key:         uintptr(unsafe.Pointer(key)),
		data:        uintptr(unsafe.Pointer(data)),
		valuesCount: uintptr(unsafe.Pointer(&valuesCount)),
		dbi:         uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_get_ex), ptr, 0)
	return int(valuesCount), args.result
}

// Put Store items into a database.
// ingroup c_crud
//
// This function stores key/data pairs in the database. The default behavior
// is to enter the new key/data pair, replacing any previously existing key
// if duplicates are disallowed, or adding a duplicate data item if
// duplicates are allowed (see ref MDBX_DUPSORT).
//
// param [in] txn        A transaction handle returned
//
//	by ref mdbx_txn_begin().
//
// param [in] dbi        A database handle returned by ref mdbx_dbi_open().
// param [in] key        The key to store in the database.
// param [in,out] data   The data to store.
// param [in] flags      Special options for this operation.
//
//	                      This parameter must be set to 0 or by bitwise OR'ing
//	                      together one or more of the values described here:
//	 - ref MDBX_NODUPDATA
//	    Enter the new key-value pair only if it does not already appear
//	    in the database. This flag may only be specified if the database
//	    was opened with ref MDBX_DUPSORT. The function will return
//	    ref MDBX_KEYEXIST if the key/data pair already appears in the database.
//
//	- ref MDBX_NOOVERWRITE
//	    Enter the new key/data pair only if the key does not already appear
//	    in the database. The function will return ref MDBX_KEYEXIST if the key
//	    already appears in the database, even if the database supports
//	    duplicates (see ref  MDBX_DUPSORT). The data parameter will be set
//	    to point to the existing item.
//
//	- ref MDBX_CURRENT
//	    Update an single existing entry, but not add new ones. The function will
//	    return ref MDBX_NOTFOUND if the given key not exist in the database.
//	    In case multi-values for the given key, with combination of
//	    the ref MDBX_ALLDUPS will replace all multi-values,
//	    otherwise return the ref MDBX_EMULTIVAL.
//
//	- ref MDBX_RESERVE
//	    Reserve space for data of the given size, but don't copy the given
//	    data. Instead, return a pointer to the reserved space, which the
//	    caller can fill in later - before the next update operation or the
//	    transaction ends. This saves an extra memcpy if the data is being
//	    generated later. MDBX does nothing else with this memory, the caller
//	    is expected to modify all of the space requested. This flag must not
//	    be specified if the database was opened with ref MDBX_DUPSORT.
//
//	- ref MDBX_APPEND
//	    Append the given key/data pair to the end of the database. This option
//	    allows fast bulk loading when keys are already known to be in the
//	    correct order. Loading unsorted keys with this flag will cause
//	    a ref MDBX_EKEYMISMATCH error.
//
//	- ref MDBX_APPENDDUP
//	    As above, but for sorted dup data.
//
//	- ref MDBX_MULTIPLE
//	    Store multiple contiguous data elements in a single request. This flag
//	    may only be specified if the database was opened with
//	    ref MDBX_DUPFIXED. With combination the ref MDBX_ALLDUPS
//	    will replace all multi-values.
//	    The data argument must be an array of two ref MDBX_val. The `iov_len`
//	    of the first ref MDBX_val must be the size of a single data element.
//	    The `iov_base` of the first ref MDBX_val must point to the beginning
//	    of the array of contiguous data elements which must be properly aligned
//	    in case of database with ref MDBX_INTEGERDUP flag.
//	    The `iov_len` of the second ref MDBX_val must be the count of the
//	    number of data elements to store. On return this field will be set to
//	    the count of the number of elements actually written. The `iov_base` of
//	    the second ref MDBX_val is unused.
//
// see ref c_crud_hints "Quick reference for Insert/Update/Delete operations"
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_KEYEXIST  The key/value pair already exists in the database.
// retval MDBX_MAP_FULL  The database is full, see ref mdbx_env_set_mapsize().
// retval MDBX_TXN_FULL  The transaction has too many dirty pages.
// retval MDBX_EACCES    An attempt was made to write
//
//	in a read-only transaction.
//
// retval MDBX_EINVAL    An invalid parameter was specified.
func (tx *Tx) Put(dbi DBI, key *Val, data *Val, flags PutFlags) Error {
	args := struct {
		txn    uintptr
		key    uintptr
		data   uintptr
		dbi    uint32
		flags  uint32
		result Error
	}{
		txn:   uintptr(unsafe.Pointer(tx.txn)),
		key:   uintptr(unsafe.Pointer(key)),
		data:  uintptr(unsafe.Pointer(data)),
		dbi:   uint32(dbi),
		flags: uint32(flags),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_put), ptr, 0)
	return args.result
}

// Replace items in a database.
// ingroup c_crud
//
// This function allows to update or delete an existing value at the same time
// as the previous value is retrieved. If the argument new_data equal is NULL
// zero, the removal is performed, otherwise the update/insert.
//
// The current value may be in an already changed (aka dirty) page. In this
// case, the page will be overwritten during the update, and the old value will
// be lost. Therefore, an additional buffer must be passed via old_data
// argument initially to copy the old value. If the buffer passed in is too
// small, the function will return ref MDBX_RESULT_TRUE by setting iov_len
// field pointed by old_data argument to the appropriate value, without
// performing any changes.
//
// For databases with non-unique keys (i.e. with ref MDBX_DUPSORT flag),
// another use case is also possible, when by old_data argument selects a
// specific item from multi-value/duplicates with the same key for deletion or
// update. To select this scenario in flags should simultaneously specify
// ref MDBX_CURRENT and ref MDBX_NOOVERWRITE. This combination is chosen
// because it makes no sense, and thus allows you to identify the request of
// such a scenario.
//
// param [in] txn           A transaction handle returned
//
//	by ref mdbx_txn_begin().
//
// param [in] dbi           A database handle returned by ref mdbx_dbi_open().
// param [in] key           The key to store in the database.
// param [in] new_data      The data to store, if NULL then deletion will
//
//	be performed.
//
// param [in,out] old_data  The buffer for retrieve previous value as describe
//
//	above.
//
// param [in] flags         Special options for this operation.
//
//	This parameter must be set to 0 or by bitwise
//	OR'ing together one or more of the values
//	described in ref mdbx_put() description above,
//	and additionally
//	(ref MDBX_CURRENT | ref MDBX_NOOVERWRITE)
//	combination for selection particular item from
//	multi-value/duplicates.
//
// see ref c_crud_hints "Quick reference for Insert/Update/Delete operations"
//
// returns A non-zero error value on failure and 0 on success.
func (tx *Tx) Replace(dbi DBI, key *Val, data *Val, oldData *Val, flags PutFlags) Error {
	args := struct {
		txn     uintptr
		key     uintptr
		data    uintptr
		oldData uintptr
		dbi     uint32
		flags   uint32
		result  Error
	}{
		txn:     uintptr(unsafe.Pointer(tx.txn)),
		key:     uintptr(unsafe.Pointer(key)),
		data:    uintptr(unsafe.Pointer(data)),
		oldData: uintptr(unsafe.Pointer(oldData)),
		dbi:     uint32(dbi),
		flags:   uint32(flags),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_replace), ptr, 0)
	return args.result
}

// Delete items from a database.
// ingroup c_crud
//
// This function removes key/data pairs from the database.
//
// note The data parameter is NOT ignored regardless the database does
// support sorted duplicate data items or not. If the data parameter
// is non-NULL only the matching data item will be deleted. Otherwise, if data
// parameter is NULL, any/all value(s) for specified key will be deleted.
//
// This function will return ref MDBX_NOTFOUND if the specified key/data
// pair is not in the database.
//
// see ref c_crud_hints "Quick reference for Insert/Update/Delete operations"
//
// param [in] txn   A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi   A database handle returned by ref mdbx_dbi_open().
// param [in] key   The key to delete from the database.
// param [in] data  The data to delete.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_EACCES   An attempt was made to write
//
//	in a read-only transaction.
//
// retval MDBX_EINVAL   An invalid parameter was specified.
func (tx *Tx) Delete(dbi DBI, key *Val, data *Val) Error {
	args := struct {
		txn    uintptr
		key    uintptr
		data   uintptr
		dbi    uint32
		result Error
	}{
		txn:  uintptr(unsafe.Pointer(tx.txn)),
		key:  uintptr(unsafe.Pointer(key)),
		data: uintptr(unsafe.Pointer(data)),
		dbi:  uint32(dbi),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_del), ptr, 0)
	return args.result
}

// Bind cursor to specified transaction and DBI handle.
// ingroup c_cursors
//
// Using of the `mdbx_cursor_bind()` is equivalent to calling
// ref mdbx_cursor_renew() but with specifying an arbitrary dbi handle.
//
// An capable of operation cursor is associated with a specific transaction and
// database. The cursor may be associated with a new transaction,
// and referencing a new or the same database handle as it was created with.
// This may be done whether the previous transaction is live or dead.
//
// note In contrast to LMDB, the MDBX required that any opened cursors can be
// reused and must be freed explicitly, regardless ones was opened in a
// read-only or write transaction. The REASON for this is eliminates ambiguity
// which helps to avoid errors such as: use-after-free, double-free, i.e.
// memory corruption and segfaults.
//
// param [in] txn      A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi      A database handle returned by ref mdbx_dbi_open().
// param [out] cursor  A cursor handle returned by ref mdbx_cursor_create().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL  An invalid parameter was specified.
func (tx *Tx) Bind(cursor *Cursor, dbi DBI) Error {
	args := struct {
		txn    uintptr
		cursor uintptr
		dbi    DBI
		result Error
	}{
		txn:    uintptr(unsafe.Pointer(tx.txn)),
		cursor: uintptr(unsafe.Pointer(cursor)),
		dbi:    dbi,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_bind), ptr, 0)
	return args.result
}

// OpenCursor Create a cursor handle for the specified transaction and DBI handle.
// ingroup c_cursors
//
// Using of the `mdbx_cursor_open()` is equivalent to calling
// ref mdbx_cursor_create() and then ref mdbx_cursor_bind() functions.
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
// param [in] txn      A transaction handle returned by ref mdbx_txn_begin().
// param [in] dbi      A database handle returned by ref mdbx_dbi_open().
// param [out] cursor  Address where the new ref MDBX_cursor handle will be
//
//	stored.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL  An invalid parameter was specified.
func (tx *Tx) OpenCursor(dbi DBI) (*Cursor, Error) {
	var cursor *C.MDBX_cursor
	args := struct {
		txn    uintptr
		cursor uintptr
		dbi    DBI
		result Error
	}{
		txn:    uintptr(unsafe.Pointer(tx.txn)),
		cursor: uintptr(unsafe.Pointer(&cursor)),
		dbi:    dbi,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_open), ptr, 0)
	return (*Cursor)(unsafe.Pointer(cursor)), args.result
}

// Close a cursor handle.
// ingroup c_cursors
//
// The cursor handle will be freed and must not be used again after this call,
// but its transaction may still be live.
//
// note In contrast to LMDB, the MDBX required that any opened cursors can be
// reused and must be freed explicitly, regardless ones was opened in a
// read-only or write transaction. The REASON for this is eliminates ambiguity
// which helps to avoid errors such as: use-after-free, double-free, i.e.
// memory corruption and segfaults.
//
// param [in] cursor  A cursor handle returned by ref mdbx_cursor_open()
//
//	or ref mdbx_cursor_create().
func (cur *Cursor) Close() Error {
	ptr := uintptr(unsafe.Pointer(cur))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_close), ptr, 0)
	return ErrSuccess
}

// Renew a cursor handle.
// ingroup c_cursors
//
// An capable of operation cursor is associated with a specific transaction and
// database. The cursor may be associated with a new transaction,
// and referencing a new or the same database handle as it was created with.
// This may be done whether the previous transaction is live or dead.
//
// Using of the `mdbx_cursor_renew()` is equivalent to calling
// ref mdbx_cursor_bind() with the DBI handle that previously
// the cursor was used with.
//
// note In contrast to LMDB, the MDBX allow any cursor to be re-used by using
// ref mdbx_cursor_renew(), to avoid unnecessary malloc/free overhead until it
// freed by ref mdbx_cursor_close().
//
// param [in] txn      A transaction handle returned by ref mdbx_txn_begin().
// param [in] cursor   A cursor handle returned by ref mdbx_cursor_open().
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL  An invalid parameter was specified.
func (cur *Cursor) Renew(tx *Tx) Error {
	args := struct {
		txn    uintptr
		cursor uintptr
		result Error
	}{
		txn:    uintptr(unsafe.Pointer(tx.txn)),
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_renew), ptr, 0)
	return args.result
}

// Tx Return the cursor's transaction handle.
// ingroup c_cursors
//
// param [in] cursor A cursor handle returned by ref mdbx_cursor_open().
func (cur *Cursor) Tx() *C.MDBX_txn {
	args := struct {
		cursor uintptr
		txn    uintptr
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_txn), ptr, 0)
	return (*C.MDBX_txn)(unsafe.Pointer(args.txn))
}

// DBI Return the cursor's database handle.
// ingroup c_cursors
//
// param [in] cursor  A cursor handle returned by ref mdbx_cursor_open().
func (cur *Cursor) DBI() DBI {
	args := struct {
		cursor uintptr
		dbi    DBI
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_dbi), ptr, 0)
	return args.dbi
}

// Copy cursor position and state.
// ingroup c_cursors
//
// param [in] src       A source cursor handle returned
// by ref mdbx_cursor_create() or ref mdbx_cursor_open().
//
// param [in,out] dest  A destination cursor handle returned
// by ref mdbx_cursor_create() or ref mdbx_cursor_open().
//
// returns A non-zero error value on failure and 0 on success.
func (cur *Cursor) Copy(dest *Cursor) Error {
	args := struct {
		src    uintptr
		dest   uintptr
		result Error
	}{
		src:  uintptr(unsafe.Pointer(cur)),
		dest: uintptr(unsafe.Pointer(dest)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_copy), ptr, 0)
	return args.result
}

// Get Retrieve by cursor.
// ingroup c_crud
//
// This function retrieves key/data pairs from the database. The address and
// length of the key are returned in the object to which key refers (except
// for the case of the ref MDBX_SET option, in which the key object is
// unchanged), and the address and length of the data are returned in the object
// to which data refers.
// see mdbx_get()
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
// param [in,out] key   The key for a retrieved item.
// param [in,out] data  The data of a retrieved item.
// param [in] op        A cursor operation ref MDBX_cursor_op.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_NOTFOUND  No matching key found.
// retval MDBX_EINVAL    An invalid parameter was specified.
func (cur *Cursor) Get(key *Val, data *Val, op CursorOp) Error {
	args := struct {
		cursor uintptr
		key    uintptr
		data   uintptr
		op     CursorOp
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
		key:    uintptr(unsafe.Pointer(key)),
		data:   uintptr(unsafe.Pointer(data)),
		op:     op,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_get), ptr, 0)
	return args.result
}

// Put Store by cursor.
// ingroup c_crud
//
// This function stores key/data pairs into the database. The cursor is
// positioned at the new item, or on failure usually near it.
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
// param [in] key       The key operated on.
// param [in,out] data  The data operated on.
// param [in] flags     Options for this operation. This parameter
//
//	                     must be set to 0 or by bitwise OR'ing together
//	                     one or more of the values described here:
//	- ref MDBX_CURRENT
//	    Replace the item at the current cursor position. The key parameter
//	    must still be provided, and must match it, otherwise the function
//	    return ref MDBX_EKEYMISMATCH. With combination the
//	    ref MDBX_ALLDUPS will replace all multi-values.
//
//	    note MDBX allows (unlike LMDB) you to change the size of the data and
//	    automatically handles reordering for sorted duplicates
//	    (see ref MDBX_DUPSORT).
//
//	- ref MDBX_NODUPDATA
//	    Enter the new key-value pair only if it does not already appear in the
//	    database. This flag may only be specified if the database was opened
//	    with ref MDBX_DUPSORT. The function will return ref MDBX_KEYEXIST
//	    if the key/data pair already appears in the database.
//
//	- ref MDBX_NOOVERWRITE
//	    Enter the new key/data pair only if the key does not already appear
//	    in the database. The function will return ref MDBX_KEYEXIST if the key
//	    already appears in the database, even if the database supports
//	    duplicates (ref MDBX_DUPSORT).
//
//	- ref MDBX_RESERVE
//	    Reserve space for data of the given size, but don't copy the given
//	    data. Instead, return a pointer to the reserved space, which the
//	    caller can fill in later - before the next update operation or the
//	    transaction ends. This saves an extra memcpy if the data is being
//	    generated later. This flag must not be specified if the database
//	    was opened with ref MDBX_DUPSORT.
//
//	- ref MDBX_APPEND
//	    Append the given key/data pair to the end of the database. No key
//	    comparisons are performed. This option allows fast bulk loading when
//	    keys are already known to be in the correct order. Loading unsorted
//	    keys with this flag will cause a ref MDBX_KEYEXIST error.
//
//	- ref MDBX_APPENDDUP
//	    As above, but for sorted dup data.
//
//	- ref MDBX_MULTIPLE
//	    Store multiple contiguous data elements in a single request. This flag
//	    may only be specified if the database was opened with
//	    ref MDBX_DUPFIXED. With combination the ref MDBX_ALLDUPS
//	    will replace all multi-values.
//	    The data argument must be an array of two ref MDBX_val. The `iov_len`
//	    of the first ref MDBX_val must be the size of a single data element.
//	    The `iov_base` of the first ref MDBX_val must point to the beginning
//	    of the array of contiguous data elements which must be properly aligned
//	    in case of database with ref MDBX_INTEGERDUP flag.
//	    The `iov_len` of the second ref MDBX_val must be the count of the
//	    number of data elements to store. On return this field will be set to
//	    the count of the number of elements actually written. The `iov_base` of
//	    the second ref MDBX_val is unused.
//
// see ref c_crud_hints "Quick reference for Insert/Update/Delete operations"
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EKEYMISMATCH  The given key value is mismatched to the current
//
//	cursor position
//
// retval MDBX_MAP_FULL      The database is full,
//
//	see ref mdbx_env_set_mapsize().
//
// retval MDBX_TXN_FULL      The transaction has too many dirty pages.
// retval MDBX_EACCES        An attempt was made to write in a read-only
//
//	transaction.
//
// retval MDBX_EINVAL        An invalid parameter was specified.
func (cur *Cursor) Put(key *Val, data *Val, flags PutFlags) Error {
	args := struct {
		cursor uintptr
		key    uintptr
		data   uintptr
		flags  PutFlags
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
		key:    uintptr(unsafe.Pointer(key)),
		data:   uintptr(unsafe.Pointer(data)),
		flags:  flags,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_put), ptr, 0)
	return args.result
}

// Delete current key/data pair.
// ingroup c_crud
//
// This function deletes the key/data pair to which the cursor refers. This
// does not invalidate the cursor, so operations such as ref MDBX_NEXT can
// still be used on it. Both ref MDBX_NEXT and ref MDBX_GET_CURRENT will
// return the same record after this operation.
//
// param [in] cursor  A cursor handle returned by mdbx_cursor_open().
// param [in] flags   Options for this operation. This parameter must be set
// to one of the values described here.
//
//   - ref MDBX_CURRENT Delete only single entry at current cursor position.
//   - ref MDBX_ALLDUPS
//     or ref MDBX_NODUPDATA (supported for compatibility)
//     Delete all of the data items for the current key. This flag has effect
//     only for database(s) was created with ref MDBX_DUPSORT.
//
// see ref c_crud_hints "Quick reference for Insert/Update/Delete operations"
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_MAP_FULL      The database is full,
//
//	see ref mdbx_env_set_mapsize().
//
// retval MDBX_TXN_FULL      The transaction has too many dirty pages.
// retval MDBX_EACCES        An attempt was made to write in a read-only
//
//	transaction.
//
// retval MDBX_EINVAL        An invalid parameter was specified.
func (cur *Cursor) Delete(flags PutFlags) Error {
	args := struct {
		cursor uintptr
		flags  PutFlags
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
		flags:  flags,
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_del), ptr, 0)
	return args.result
}

// Count Return count of duplicates for current key.
// ingroup c_crud
//
// This call is valid for all databases, but reasonable only for that support
// sorted duplicate data items ref MDBX_DUPSORT.
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
// param [out] pcount   Address where the count will be stored.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_THREAD_MISMATCH  Given transaction is not owned
//
//	by current thread.
//
// retval MDBX_EINVAL   Cursor is not initialized, or an invalid parameter
//
//	was specified.
func (cur *Cursor) Count() (int, Error) {
	var count uintptr
	args := struct {
		cursor uintptr
		count  uintptr
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
		count:  uintptr(unsafe.Pointer(&count)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_count), ptr, 0)
	return int(count), args.result
}

// EOF Determines whether the cursor is pointed to a key-value pair or not,
// i.e. was not positioned or points to the end of data.
// ingroup c_cursors
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
//
// returns A ref MDBX_RESULT_TRUE or ref MDBX_RESULT_FALSE value,
//
//	otherwise the error code:
//
// retval MDBX_RESULT_TRUE    No more data available or cursor not
//
//	positioned
//
// retval MDBX_RESULT_FALSE   A data is available
// retval Otherwise the error code
func (cur *Cursor) EOF() Error {
	args := struct {
		cursor uintptr
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_eof), ptr, 0)
	return args.result
}

// First Determines whether the cursor is pointed to the first key-value pair
// or not. ingroup c_cursors
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
//
// returns A MDBX_RESULT_TRUE or MDBX_RESULT_FALSE value,
//
//	otherwise the error code:
//
// retval MDBX_RESULT_TRUE   Cursor positioned to the first key-value pair
// retval MDBX_RESULT_FALSE  Cursor NOT positioned to the first key-value
// pair retval Otherwise the error code
func (cur *Cursor) First() Error {
	args := struct {
		cursor uintptr
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_on_first), ptr, 0)
	return args.result
}

// Last Determines whether the cursor is pointed to the last key-value pair
// or not. ingroup c_cursors
//
// param [in] cursor    A cursor handle returned by ref mdbx_cursor_open().
//
// returns A ref MDBX_RESULT_TRUE or ref MDBX_RESULT_FALSE value,
//
//	otherwise the error code:
//
// retval MDBX_RESULT_TRUE   Cursor positioned to the last key-value pair
// retval MDBX_RESULT_FALSE  Cursor NOT positioned to the last key-value pair
// retval Otherwise the error code
func (cur *Cursor) Last() Error {
	args := struct {
		cursor uintptr
		result Error
	}{
		cursor: uintptr(unsafe.Pointer(cur)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_cursor_on_last), ptr, 0)
	return args.result
}

// EstimateDistance
// details note The estimation result varies greatly depending on the filling
// of specific pages and the overall balance of the b-tree:
//
// 1. The number of items is estimated by analyzing the height and fullness of
// the b-tree. The accuracy of the result directly depends on the balance of
// the b-tree, which in turn is determined by the history of previous
// insert/delete operations and the nature of the data (i.e. variability of
// keys length and so on). Therefore, the accuracy of the estimation can vary
// greatly in a particular situation.
//
// 2. To understand the potential spread of results, you should consider a
// possible situations basing on the general criteria for splitting and merging
// b-tree pages:
//  - the page is split into two when there is no space for added data;
//  - two pages merge if the result fits in half a page;
//  - thus, the b-tree can consist of an arbitrary combination of pages filled
//    both completely and only 1/4. Therefore, in the worst case, the result
//    can diverge 4 times for each level of the b-tree excepting the first and
//    the last.
//
// 3. In practice, the probability of extreme cases of the above situation is
// close to zero and in most cases the error does not exceed a few percent. On
// the other hand, it's just a chance you shouldn't overestimate.///

// EstimateDistance the distance between cursors as a number of elements.
// ingroup c_rqest
//
// This function performs a rough estimate based only on b-tree pages that are
// common for the both cursor's stacks. The results of such estimation can be
// used to build and/or optimize query execution plans.
//
// Please see notes on accuracy of the result in the details
// of ref c_rqest section.
//
// Both cursors must be initialized for the same database and the same
// transaction.
//
// param [in] first            The first cursor for estimation.
// param [in] last             The second cursor for estimation.
// param [out] distance_items  The pointer to store estimated distance value,
//
//	i.e. `*distance_items = distance(first, last)`.
//
// returns A non-zero error value on failure and 0 on success.
func EstimateDistance(first, last *Cursor) (int64, Error) {
	var distance int64
	args := struct {
		first    uintptr
		last     uintptr
		distance uintptr
		result   Error
	}{
		first:    uintptr(unsafe.Pointer(first)),
		last:     uintptr(unsafe.Pointer(last)),
		distance: uintptr(unsafe.Pointer(&distance)),
	}
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_estimate_distance), ptr, 0)
	return distance, args.result
}
