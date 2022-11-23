package gmdbx

//#include "mdbxgo.h"
import "C"
import (
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/sunvim/gmdbx/unsafecgo"
)

func init() {
	sz0 := unsafe.Sizeof(C.MDBX_envinfo{})
	sz1 := unsafe.Sizeof(EnvInfo{})
	if sz0 != sz1 {
		panic("sizeof(C.MDBX_envinfo) != sizeof(EnvInfo{})")
	}
}

type EnvInfo struct {
	Geo struct {
		Lower   uint64 // Lower limit for datafile size
		Upper   uint64 // Upper limit for datafile size
		Current uint64 // Current datafile size
		Shrink  uint64 // Shrink threshold for datafile
		Grow    uint64 // Growth step for datafile
	}
	MapSize               uint64 // Size of the data memory map
	LastPageNumber        uint64 // Number of the last used page
	RecentTxnID           uint64 // ID of the last committed transaction
	LatterReaderTxnID     uint64 // ID of the last reader transaction
	SelfLatterReaderTxnID uint64 // ID of the last reader transaction of caller process

	Meta0TxnID, MIMeta0Sign uint64
	Meta1TxnID, MIMeta1Sign uint64
	Meta2TxnID, MIMeta2Sign uint64

	MaxReaders  uint32 // Total reader slots in the environment
	NumReaders  uint32 // Max reader slots used in the environment
	DXBPageSize uint32 // Database pagesize
	SysPageSize uint32 // System pagesize

	// BootID A mostly unique ID that is regenerated on each boot.
	// As such it can be used to identify the local machine's current boot. MDBX
	// uses such when open the database to determine whether rollback required to
	// the last steady sync point or not. I.e. if current bootid is differ from the
	// value within a database then the system was rebooted and all changes since
	// last steady sync must be reverted for data integrity. Zeros mean that no
	// relevant information is available from the system.
	BootID struct {
		Current, Meta0, Meta1, Meta2 struct{ X, Y uint64 }
	}

	UnSyncVolume                   uint64 // Bytes not explicitly synchronized to disk.
	AutoSyncThreshold              uint64 // Current auto-sync threshold, see ref mdbx_env_set_syncbytes().
	SinceSyncSeconds16Dot16        uint32 // Time since the last steady sync in 1/65536 of second
	AutoSyncPeriodSeconds16Dot16   uint32 // Current auto-sync period in 1/65536 of second, see ref mdbx_env_set_syncperiod().
	SinceReaderCheckSeconds16Dot16 uint32 // Time since the last readers check in 1/65536 of second, see ref mdbx_reader_check()
	Mode                           uint32 // Current environment mode. The same as ref mdbx_env_get_flags() returns.

	// Statistics of page operations.
	// details Overall statistics of page operations of all (running, completed
	// and aborted) transactions in the current multi-process session (since the
	// first process opened the database after everyone had previously closed it).
	PGOpStat struct {
		Newly                 uint64 // Quantity of a new pages added
		Cow                   uint64 // Quantity of pages copied for update
		Clone                 uint64 // Quantity of parent's dirty pages clones for nested transactions
		Split                 uint64 // Page splits
		Merge                 uint64 // Page merges
		Spill                 uint64 // Quantity of spilled dirty pages
		UnSpill               uint64 // Quantity of unspilled/reloaded pages
		Wops                  uint64 // Number of explicit write operations (not a pages) to a disk
		GcrtimeSeconds16dot16 uint64 //Time spent loading and searching inside GC (aka FreeDB) in 1/65536 of second.
	}
}

type EnvFlags uint32

const (
	EnvEnvDefaults EnvFlags = 0

	// EnvNoSubDir No environment directory.
	//
	// By default, MDBX creates its environment in a directory whose pathname is
	// given in path, and creates its data and lock files under that directory.
	// With this option, path is used as-is for the database rootDB data file.
	// The database lock file is the path with "-lck" appended.
	//
	// - with `MDBX_NOSUBDIR` = in a filesystem we have the pair of MDBX-files
	//   which names derived from given pathname by appending predefined suffixes.
	//
	// - without `MDBX_NOSUBDIR` = in a filesystem we have the MDBX-directory with
	//   given pathname, within that a pair of MDBX-files with predefined names.
	//
	// This flag affects only at new environment creating by ref mdbx_env_open(),
	// otherwise at opening an existing environment libmdbx will choice this
	// automatically.
	EnvNoSubDir = EnvFlags(C.MDBX_NOSUBDIR)

	// EnvReadOnly Read only mode.
	//
	// Open the environment in read-only mode. No write operations will be
	// allowed. MDBX will still modify the lock file - except on read-only
	// filesystems, where MDBX does not use locks.
	//
	// - with `MDBX_RDONLY` = open environment in read-only mode.
	//   MDBX supports pure read-only mode (i.e. without opening LCK-file) only
	//   when environment directory and/or both files are not writable (and the
	//   LCK-file may be missing). In such case allowing file(s) to be placed
	//   on a network read-only share.
	//
	// - without `MDBX_RDONLY` = open environment in read-write mode.
	//
	// This flag affects only at environment opening but can't be changed after.
	EnvReadOnly = EnvFlags(C.MDBX_RDONLY)

	// EnvExclusive Open environment in exclusive/monopolistic mode.
	//
	// `MDBX_EXCLUSIVE` flag can be used as a replacement for `MDB_NOLOCK`,
	// which don't supported by MDBX.
	// In this way, you can get the minimal overhead, but with the correct
	// multi-process and multi-thread locking.
	//
	// - with `MDBX_EXCLUSIVE` = open environment in exclusive/monopolistic mode
	//   or return ref MDBX_BUSY if environment already used by other process.
	//   The rootDB feature of the exclusive mode is the ability to open the
	//   environment placed on a network share.
	//
	// - without `MDBX_EXCLUSIVE` = open environment in cooperative mode,
	//   i.e. for multi-process access/interaction/cooperation.
	//   The rootDB requirements of the cooperative mode are:
	//
	//   1. data files MUST be placed in the LOCAL file system,
	//      but NOT on a network share.
	//   2. environment MUST be opened only by LOCAL processes,
	//      but NOT over a network.
	//   3. OS kernel (i.e. file system and memory mapping implementation) and
	//      all processes that open the given environment MUST be running
	//      in the physically single RAM with cache-coherency. The only
	//      exception for cache-consistency requirement is Linux on MIPS
	//      architecture, but this case has not been tested for a long time).
	//
	// This flag affects only at environment opening but can't be changed after.
	EnvExclusive = EnvFlags(C.MDBX_EXCLUSIVE)

	// EnvAccede Using database/environment which already opened by another process(es).
	//
	// The `MDBX_ACCEDE` flag is useful to avoid ref MDBX_INCOMPATIBLE error
	// while opening the database/environment which is already used by another
	// process(es) with unknown mode/flags. In such cases, if there is a
	// difference in the specified flags (ref MDBX_NOMETASYNC,
	// ref MDBX_SAFE_NOSYNC, ref MDBX_UTTERLY_NOSYNC, ref MDBX_LIFORECLAIM,
	// ref MDBX_COALESCE and ref MDBX_NORDAHEAD), instead of returning an error,
	// the database will be opened in a compatibility with the already used mode.
	//
	// `MDBX_ACCEDE` has no effect if the current process is the only one either
	// opening the DB in read-only mode or other process(es) uses the DB in
	// read-only mode.
	EnvAccede = EnvFlags(C.MDBX_ACCEDE)

	// EnvWriteMap Map data into memory with write permission.
	//
	// Use a writeable memory map unless ref MDBX_RDONLY is set. This uses fewer
	// mallocs and requires much less work for tracking database pages, but
	// loses protection from application bugs like wild pointer writes and other
	// bad updates into the database. This may be slightly faster for DBs that
	// fit entirely in RAM, but is slower for DBs larger than RAM. Also adds the
	// possibility for stray application writes thru pointers to silently
	// corrupt the database.
	//
	// - with `MDBX_WRITEMAP` = all data will be mapped into memory in the
	//   read-write mode. This offers a significant performance benefit, since the
	//   data will be modified directly in mapped memory and then flushed to disk
	//   by single system call, without any memory management nor copying.
	//
	// - without `MDBX_WRITEMAP` = data will be mapped into memory in the
	//   read-only mode. This requires stocking all modified database pages in
	//   memory and then writing them to disk through file operations.
	//
	// warning On the other hand, `MDBX_WRITEMAP` adds the possibility for stray
	// application writes thru pointers to silently corrupt the database.
	//
	// note The `MDBX_WRITEMAP` mode is incompatible with nested transactions,
	// since this is unreasonable. I.e. nested transactions requires mallocation
	// of database pages and more work for tracking ones, which neuters a
	// performance boost caused by the `MDBX_WRITEMAP` mode.
	//
	// This flag affects only at environment opening but can't be changed after.
	EnvWriteMap = EnvFlags(C.MDBX_WRITEMAP)

	// EnvNoTLS Tie reader locktable slots to read-only transactions
	// instead of to threads.
	//
	// Don't use Thread-Local Storage, instead tie reader locktable slots to
	// ref MDBX_txn objects instead of to threads. So, ref mdbx_txn_reset()
	// keeps the slot reserved for the ref MDBX_txn object. A thread may use
	// parallel read-only transactions. And a read-only transaction may span
	// threads if you synchronizes its use.
	//
	// Applications that multiplex many user threads over individual OS threads
	// need this option. Such an application must also serialize the write
	// transactions in an OS thread, since MDBX's write locking is unaware of
	// the user threads.
	//
	// note Regardless to `MDBX_NOTLS` flag a write transaction entirely should
	// always be used in one thread from start to finish. MDBX checks this in a
	// reasonable manner and return the ref MDBX_THREAD_MISMATCH error in rules
	// violation.
	//
	// This flag affects only at environment opening but can't be changed after.
	EnvNoTLS = EnvFlags(C.MDBX_NOTLS)
	//MDBX_NOTLS = UINT32_C(0x200000)

	// EnvNoReadAhead Don't do readahead.
	//
	// Turn off readahead. Most operating systems perform readahead on read
	// requests by default. This option turns it off if the OS supports it.
	// Turning it off may help random read performance when the DB is larger
	// than RAM and system RAM is full.
	//
	// By default libmdbx dynamically enables/disables readahead depending on
	// the actual database size and currently available memory. On the other
	// hand, such automation has some limitation, i.e. could be performed only
	// when DB size changing but can't tracks and reacts changing a free RAM
	// availability, since it changes independently and asynchronously.
	//
	// note The mdbx_is_readahead_reasonable() function allows to quickly find
	// out whether to use readahead or not based on the size of the data and the
	// amount of available memory.
	//
	// This flag affects only at environment opening and can't be changed after.
	EnvNoReadAhead = EnvFlags(C.MDBX_NORDAHEAD)

	// EnvNoMemInit Don't initialize malloc'ed memory before writing to datafile.
	//
	// Don't initialize malloc'ed memory before writing to unused spaces in the
	// data file. By default, memory for pages written to the data file is
	// obtained using malloc. While these pages may be reused in subsequent
	// transactions, freshly malloc'ed pages will be initialized to zeroes before
	// use. This avoids persisting leftover data from other code (that used the
	// heap and subsequently freed the memory) into the data file.
	//
	// Note that many other system libraries may allocate and free memory from
	// the heap for arbitrary uses. E.g., stdio may use the heap for file I/O
	// buffers. This initialization step has a modest performance cost so some
	// applications may want to disable it using this flag. This option can be a
	// problem for applications which handle sensitive data like passwords, and
	// it makes memory checkers like Valgrind noisy. This flag is not needed
	// with ref MDBX_WRITEMAP, which writes directly to the mmap instead of using
	// malloc for pages. The initialization is also skipped if ref MDBX_RESERVE
	// is used; the caller is expected to overwrite all of the memory that was
	// reserved in that case.
	//
	// This flag may be changed at any time using `mdbx_env_set_flags()`.
	EnvNoMemInit = EnvFlags(C.MDBX_NOMEMINIT)

	// EnvCoalesce Aims to coalesce a Garbage Collection items.
	//
	// With `MDBX_COALESCE` flag MDBX will aims to coalesce items while recycling
	// a Garbage Collection. Technically, when possible short lists of pages
	// will be combined into longer ones, but to fit on one database page. As a
	// result, there will be fewer items in Garbage Collection and a page lists
	// are longer, which slightly increases the likelihood of returning pages to
	// Unallocated space and reducing the database file.
	//
	// This flag may be changed at any time using mdbx_env_set_flags().
	EnvCoalesce = EnvFlags(C.MDBX_COALESCE)

	// EnvLIFOReclaim LIFO policy for recycling a Garbage Collection items.
	//
	// `MDBX_LIFORECLAIM` flag turns on LIFO policy for recycling a Garbage
	// Collection items, instead of FIFO by default. On systems with a disk
	// write-back cache, this can significantly increase write performance, up
	// to several times in a best case scenario.
	//
	// LIFO recycling policy means that for reuse pages will be taken which became
	// unused the lastest (i.e. just now or most recently). Therefore the loop of
	// database pages circulation becomes as short as possible. In other words,
	// the number of pages, that are overwritten in memory and on disk during a
	// series of write transactions, will be as small as possible. Thus creates
	// ideal conditions for the efficient operation of the disk write-back cache.
	//
	// ref MDBX_LIFORECLAIM is compatible with all no-sync flags, but gives NO
	// noticeable impact in combination with ref MDBX_SAFE_NOSYNC or
	// ref MDBX_UTTERLY_NOSYN-Because MDBX will reused pages only before the
	// last "steady" MVCC-snapshot, i.e. the loop length of database pages
	// circulation will be mostly defined by frequency of calling
	// ref mdbx_env_sync() rather than LIFO and FIFO difference.
	//
	// This flag may be changed at any time using mdbx_env_set_flags().
	EnvLIFOReclaim = EnvFlags(C.MDBX_LIFORECLAIM)

	// EnvPagPerTurb Debugging option, fill/perturb released pages.
	EnvPagePerTurb = EnvFlags(C.MDBX_PAGEPERTURB)

	// SYNC MODES

	// defgroup sync_modes SYNC MODES
	//
	// attention Using any combination of ref MDBX_SAFE_NOSYNC, ref
	// MDBX_NOMETASYNC and especially ref MDBX_UTTERLY_NOSYNC is always a deal to
	// reduce durability for gain write performance. You must know exactly what
	// you are doing and what risks you are taking!
	//
	// note for LMDB users: ref MDBX_SAFE_NOSYNC is NOT similar to LMDB_NOSYNC,
	// but ref MDBX_UTTERLY_NOSYNC is exactly match LMDB_NOSYN-See details
	// below.
	//
	// THE SCENE:
	// - The DAT-file contains several MVCC-snapshots of B-tree at same time,
	//   each of those B-tree has its own root page.
	// - Each of meta pages at the beginning of the DAT file contains a
	//   pointer to the root page of B-tree which is the result of the particular
	//   transaction, and a number of this transaction.
	// - For data durability, MDBX must first write all MVCC-snapshot data
	//   pages and ensure that are written to the disk, then update a meta page
	//   with the new transaction number and a pointer to the corresponding new
	//   root page, and flush any buffers yet again.
	// - Thus during commit a I/O buffers should be flushed to the disk twice;
	//   i.e. fdatasync(), FlushFileBuffers() or similar syscall should be
	//   called twice for each commit. This is very expensive for performance,
	//   but guaranteed durability even on unexpected system failure or power
	//   outage. Of course, provided that the operating system and the
	//   underlying hardware (e.g. disk) work correctly.
	//
	// TRADE-OFF:
	// By skipping some stages described above, you can significantly benefit in
	// speed, while partially or completely losing in the guarantee of data
	// durability and/or consistency in the event of system or power failure.
	// Moreover, if for any reason disk write order is not preserved, then at
	// moment of a system crash, a meta-page with a pointer to the new B-tree may
	// be written to disk, while the itself B-tree not yet. In that case, the
	// database will be corrupted!
	//
	// see MDBX_SYNC_DURABLE see MDBX_NOMETASYNC see MDBX_SAFE_NOSYNC
	// see MDBX_UTTERLY_NOSYNC
	//
	// @{

	// EnvSyncDurable Default robust and durable sync mode.
	//
	// Metadata is written and flushed to disk after a data is written and
	// flushed, which guarantees the integrity of the database in the event
	// of a crash at any time.
	//
	// attention Please do not use other modes until you have studied all the
	// details and are sure. Otherwise, you may lose your users' data, as happens
	// in [Miranda NG](https://www.miranda-ng.org/) messenger.
	EnvSyncDurable = EnvFlags(C.MDBX_SYNC_DURABLE)

	// EnvNoMetaSync Don't sync the meta-page after commit.
	//
	// Flush system buffers to disk only once per transaction commit, omit the
	// metadata flush. Defer that until the system flushes files to disk,
	// or next non-ref MDBX_RDONLY commit or ref mdbx_env_sync(). Depending on
	// the platform and hardware, with ref MDBX_NOMETASYNC you may get a doubling
	// of write performance.
	//
	// This trade-off maintains database integrity, but a system crash may
	// undo the last committed transaction. I.e. it preserves the ACI
	// (atomicity, consistency, isolation) but not D (durability) database
	// property.
	//
	// `MDBX_NOMETASYNC` flag may be changed at any time using
	// ref mdbx_env_set_flags() or by passing to ref mdbx_txn_begin() for
	// particular write transaction. see sync_modes
	EnvNoMetaSync = EnvFlags(C.MDBX_NOMETASYNC)

	// EnvSafeNoSync Don't sync anything but keep previous steady commits.
	//
	// Like ref MDBX_UTTERLY_NOSYNC the `MDBX_SAFE_NOSYNC` flag disable similarly
	// flush system buffers to disk when committing a transaction. But there is a
	// huge difference in how are recycled the MVCC snapshots corresponding to
	// previous "steady" transactions (see below).
	//
	// With ref MDBX_WRITEMAP the `MDBX_SAFE_NOSYNC` instructs MDBX to use
	// asynchronous mmap-flushes to disk. Asynchronous mmap-flushes means that
	// actually all writes will scheduled and performed by operation system on it
	// own manner, i.e. unordered. MDBX itself just notify operating system that
	// it would be nice to write data to disk, but no more.
	//
	// Depending on the platform and hardware, with `MDBX_SAFE_NOSYNC` you may get
	// a multiple increase of write performance, even 10 times or more.
	//
	// In contrast to ref MDBX_UTTERLY_NOSYNC mode, with `MDBX_SAFE_NOSYNC` flag
	// MDBX will keeps untouched pages within B-tree of the last transaction
	// "steady" which was synced to disk completely. This has big implications for
	// both data durability and (unfortunately) performance:
	//  - a system crash can't corrupt the database, but you will lose the last
	//    transactions; because MDBX will rollback to last steady commit since it
	//    kept explicitly.
	//  - the last steady transaction makes an effect similar to "long-lived" read
	//    transaction (see above in the ref restrictions section) since prevents
	//    reuse of pages freed by newer write transactions, thus the any data
	//    changes will be placed in newly allocated pages.
	//  - to avoid rapid database growth, the system will sync data and issue
	//    a steady commit-point to resume reuse pages, each time there is
	//    insufficient space and before increasing the size of the file on disk.
	//
	// In other words, with `MDBX_SAFE_NOSYNC` flag MDBX insures you from the
	// whole database corruption, at the cost increasing database size and/or
	// number of disk IOPs. So, `MDBX_SAFE_NOSYNC` flag could be used with
	// ref mdbx_env_sync() as alternatively for batch committing or nested
	// transaction (in some cases). As well, auto-sync feature exposed by
	// ref mdbx_env_set_syncbytes() and ref mdbx_env_set_syncperiod() functions
	// could be very useful with `MDBX_SAFE_NOSYNC` flag.
	//
	// The number and volume of of disk IOPs with MDBX_SAFE_NOSYNC flag will
	// exactly the as without any no-sync flags. However, you should expect a
	// larger process's [work set](https://bit.ly/2kA2tFX) and significantly worse
	// a [locality of reference](https://bit.ly/2mbYq2J), due to the more
	// intensive allocation of previously unused pages and increase the size of
	// the database.
	//
	// `MDBX_SAFE_NOSYNC` flag may be changed at any time using
	// ref mdbx_env_set_flags() or by passing to ref mdbx_txn_begin() for
	// particular write transaction.
	EnvSafeNoSync = EnvFlags(C.MDBX_SAFE_NOSYNC)

	// EnvUtterlyNoSync Don't sync anything and wipe previous steady commits.
	//
	// Don't flush system buffers to disk when committing a transaction. This
	// optimization means a system crash can corrupt the database, if buffers are
	// not yet flushed to disk. Depending on the platform and hardware, with
	// `MDBX_UTTERLY_NOSYNC` you may get a multiple increase of write performance,
	// even 100 times or more.
	//
	// If the filesystem preserves write order (which is rare and never provided
	// unless explicitly noted) and the ref MDBX_WRITEMAP and ref
	// MDBX_LIFORECLAIM flags are not used, then a system crash can't corrupt the
	// database, but you can lose the last transactions, if at least one buffer is
	// not yet flushed to disk. The risk is governed by how often the system
	// flushes dirty buffers to disk and how often ref mdbx_env_sync() is called.
	// So, transactions exhibit ACI (atomicity, consistency, isolation) properties
	// and only lose `D` (durability). I.e. database integrity is maintained, but
	// a system crash may undo the final transactions.
	//
	// Otherwise, if the filesystem not preserves write order (which is
	// typically) or ref MDBX_WRITEMAP or ref MDBX_LIFORECLAIM flags are used,
	// you should expect the corrupted database after a system crash.
	//
	// So, most important thing about `MDBX_UTTERLY_NOSYNC`:
	//  - a system crash immediately after commit the write transaction
	//    high likely lead to database corruption.
	//  - successful completion of mdbx_env_sync(force = true) after one or
	//    more committed transactions guarantees consistency and durability.
	//  - BUT by committing two or more transactions you back database into
	//    a weak state, in which a system crash may lead to database corruption!
	//    In case single transaction after mdbx_env_sync, you may lose transaction
	//    itself, but not a whole database.
	//
	// Nevertheless, `MDBX_UTTERLY_NOSYNC` provides "weak" durability in case
	// of an application crash (but no durability on system failure), and
	// therefore may be very useful in scenarios where data durability is
	// not required over a system failure (e.g for short-lived data), or if you
	// can take such risk.
	//
	// `MDBX_UTTERLY_NOSYNC` flag may be changed at any time using
	// ref mdbx_env_set_flags(), but don't has effect if passed to
	// ref mdbx_txn_begin() for particular write transaction. see sync_modes
	EnvUtterlyNoSync = EnvFlags(C.MDBX_UTTERLY_NOSYNC)
)

type DBFlags uint32

const (
	DBDefaults = DBFlags(C.MDBX_DB_DEFAULTS)

	// DBReverseKey Use reverse string keys
	DBReverseKey = DBFlags(C.MDBX_REVERSEKEY)

	// DBDupSort Use sorted duplicates, i.e. allow multi-values
	DBDupSort = DBFlags(C.MDBX_DUPSORT)

	// DBIntegerKey Numeric keys in native byte order either uint32_t or uint64_t. The keys
	// must all be of the same size and must be aligned while passing as
	// arguments.
	DBIntegerKey = DBFlags(C.MDBX_INTEGERKEY)

	// DBDupFixed With ref MDBX_DUPSORT; sorted dup items have fixed size
	DBDupFixed = DBFlags(C.MDBX_DUPFIXED)

	// DBIntegerGroup With ref MDBX_DUPSORT and with ref MDBX_DUPFIXED; dups are fixed size
	// ref MDBX_INTEGERKEY -style integers. The data values must all be of the
	// same size and must be aligned while passing as arguments.
	DBIntegerGroup = DBFlags(C.MDBX_INTEGERDUP)

	// DBReverseDup With ref MDBX_DUPSORT; use reverse string comparison
	DBReverseDup = DBFlags(C.MDBX_REVERSEDUP)

	// DBCreate Create DB if not already existing
	DBCreate = DBFlags(C.MDBX_CREATE)

	// DBAccede Opens an existing sub-database created with unknown flags.
	//
	// The `MDBX_DB_ACCEDE` flag is intend to open a existing sub-database which
	// was created with unknown flags (ref MDBX_REVERSEKEY, ref MDBX_DUPSORT,
	// ref MDBX_INTEGERKEY, ref MDBX_DUPFIXED, ref MDBX_INTEGERDUP and
	// ref MDBX_REVERSEDUP).
	//
	// In such cases, instead of returning the ref MDBX_INCOMPATIBLE error, the
	// sub-database will be opened with flags which it was created, and then an
	// application could determine the actual flags by ref mdbx_dbi_flags().
	DBAccede = DBFlags(C.MDBX_DB_ACCEDE)
)

type CopyFlags uint32

const (
	CopyDefaults = CopyFlags(C.MDBX_CP_DEFAULTS)

	// CopyCompact Copy and compact: Omit free space from copy and renumber all
	// pages sequentially
	CopyCompact = CopyFlags(C.MDBX_CP_COMPACT)

	// CopyForceDynamicSize Force to make resizeable copy, i.e. dynamic size instead of fixed
	CopyForceDynamicSize = CopyFlags(C.MDBX_CP_FORCE_DYNAMIC_SIZE)
)

type Opt int32

const (
	// OptMaxDB brief Controls the maximum number of named databases for the environment.
	//
	// details By default only unnamed key-value database could used and
	// appropriate value should set by `MDBX_opt_max_db` to using any more named
	// subDB(s). To reduce overhead, use the minimum sufficient value. This option
	// may only set after ref mdbx_env_create() and before ref mdbx_env_open().
	//
	// see mdbx_env_set_maxdbs() see mdbx_env_get_maxdbs()
	OptMaxDB = Opt(C.MDBX_opt_max_db)

	// OptMaxReaders brief Defines the maximum number of threads/reader slots
	// for all processes interacting with the database.
	//
	// details This defines the number of slots in the lock table that is used to
	// track readers in the the environment. The default is about 100 for 4K
	// system page size. Starting a read-only transaction normally ties a lock
	// table slot to the current thread until the environment closes or the thread
	// exits. If ref MDBX_NOTLS is in use, ref mdbx_txn_begin() instead ties the
	// slot to the ref MDBX_txn object until it or the ref MDBX_env object is
	// destroyed. This option may only set after ref mdbx_env_create() and before
	// ref mdbx_env_open(), and has an effect only when the database is opened by
	// the first process interacts with the database.
	//
	// see mdbx_env_set_maxreaders() see mdbx_env_get_maxreaders()
	OptMaxReaders = Opt(C.MDBX_opt_max_readers)

	// OptSyncBytes brief Controls interprocess/shared threshold to force flush the data
	// buffers to disk, if ref MDBX_SAFE_NOSYNC is used.
	//
	// see mdbx_env_set_syncbytes() see mdbx_env_get_syncbytes()
	OptSyncBytes = Opt(C.MDBX_opt_sync_bytes)

	// OptSyncPeriod brief Controls interprocess/shared relative period since the last
	// unsteady commit to force flush the data buffers to disk,
	// if ref MDBX_SAFE_NOSYNC is used.
	// see mdbx_env_set_syncperiod() see mdbx_env_get_syncperiod()
	OptSyncPeriod = Opt(C.MDBX_opt_sync_period)

	// OptRpAugmentLimit brief Controls the in-process limit to grow a list of reclaimed/recycled
	// page's numbers for finding a sequence of contiguous pages for large data
	// items.
	//
	// details A long values requires allocation of contiguous database pages.
	// To find such sequences, it may be necessary to accumulate very large lists,
	// especially when placing very long values (more than a megabyte) in a large
	// databases (several tens of gigabytes), which is much expensive in extreme
	// cases. This threshold allows you to avoid such costs by allocating new
	// pages at the end of the database (with its possible growth on disk),
	// instead of further accumulating/reclaiming Garbage Collection records.
	//
	// On the other hand, too small threshold will lead to unreasonable database
	// growth, or/and to the inability of put long values.
	//
	// The `MDBX_opt_rp_augment_limit` controls described limit for the current
	// process. Default is 262144, it is usually enough for most cases.
	OptRpAugmentLimit = Opt(C.MDBX_opt_rp_augment_limit)

	// OptLooseLimit brief Controls the in-process limit to grow a cache of dirty
	// pages for reuse in the current transaction.
	//
	// details A 'dirty page' refers to a page that has been updated in memory
	// only, the changes to a dirty page are not yet stored on disk.
	// To reduce overhead, it is reasonable to release not all such pages
	// immediately, but to leave some ones in cache for reuse in the current
	// transaction.
	//
	// The `MDBX_opt_loose_limit` allows you to set a limit for such cache inside
	// the current process. Should be in the range 0..255, default is 64.
	OptLooseLimit = Opt(C.MDBX_opt_loose_limit)

	// OptDpReserveLimit brief Controls the in-process limit of a pre-allocated memory items
	// for dirty pages.
	//
	// details A 'dirty page' refers to a page that has been updated in memory
	// only, the changes to a dirty page are not yet stored on disk.
	// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and
	// released when a transaction is committed. To reduce overhead, it is
	// reasonable to release not all ones, but to leave some allocations in
	// reserve for reuse in the next transaction(s).
	//
	// The `MDBX_opt_dp_reserve_limit` allows you to set a limit for such reserve
	// inside the current process. Default is 1024.
	OptDpReserveLimit = Opt(C.MDBX_opt_dp_reserve_limit)

	// OptTxnDpLimit brief Controls the in-process limit of dirty pages
	// for a write transaction.
	//
	// details A 'dirty page' refers to a page that has been updated in memory
	// only, the changes to a dirty page are not yet stored on disk.
	// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and will
	// be busy until are written to disk. Therefore for a large transactions is
	// reasonable to limit dirty pages collecting above an some threshold but
	// spill to disk instead.
	//
	// The `MDBX_opt_txn_dp_limit` controls described threshold for the current
	// process. Default is 65536, it is usually enough for most cases.
	OptTxnDpLimit = Opt(C.MDBX_opt_txn_dp_limit)

	// OptTxnDpInitial brief Controls the in-process initial allocation size for dirty pages
	// list of a write transaction. Default is 1024.
	OptTxnDpInitial = Opt(C.MDBX_opt_txn_dp_initial)

	// OptSpillMaxDenomiator brief Controls the in-process how maximal part of the dirty pages may be
	// spilled when necessary.
	//
	// details The `MDBX_opt_spill_max_denominator` defines the denominator for
	// limiting from the top for part of the current dirty pages may be spilled
	// when the free room for a new dirty pages (i.e. distance to the
	// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
	// operation.
	// Exactly `max_pages_to_spill = dirty_pages - dirty_pages / N`,
	// where `N` is the value set by `MDBX_opt_spill_max_denominator`.
	//
	// Should be in the range 0..255, where zero means no limit, i.e. all dirty
	// pages could be spilled. Default is 8, i.e. no more than 7/8 of the current
	// dirty pages may be spilled when reached the condition described above.
	OptSpillMaxDenomiator = Opt(C.MDBX_opt_spill_max_denominator)

	// OptSpillMinDenomiator brief Controls the in-process how minimal part of the dirty pages should
	// be spilled when necessary.
	//
	// details The `MDBX_opt_spill_min_denominator` defines the denominator for
	// limiting from the bottom for part of the current dirty pages should be
	// spilled when the free room for a new dirty pages (i.e. distance to the
	// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
	// operation.
	// Exactly `min_pages_to_spill = dirty_pages / N`,
	// where `N` is the value set by `MDBX_opt_spill_min_denominator`.
	//
	// Should be in the range 0..255, where zero means no restriction at the
	// bottom. Default is 8, i.e. at least the 1/8 of the current dirty pages
	// should be spilled when reached the condition described above.
	OptSpillMinDenomiator = Opt(C.MDBX_opt_spill_min_denominator)

	// OptSpillParent4ChildDenominator brief Controls the in-process how much of the parent transaction dirty
	// pages will be spilled while start each child transaction.
	//
	// details The `MDBX_opt_spill_parent4child_denominator` defines the
	// denominator to determine how much of parent transaction dirty pages will be
	// spilled explicitly while start each child transaction.
	// Exactly `pages_to_spill = dirty_pages / N`,
	// where `N` is the value set by `MDBX_opt_spill_parent4child_denominator`.
	//
	// For a stack of nested transactions each dirty page could be spilled only
	// once, and parent's dirty pages couldn't be spilled while child
	// transaction(s) are running. Therefore a child transaction could reach
	// ref MDBX_TXN_FULL when parent(s) transaction has  spilled too less (and
	// child reach the limit of dirty pages), either when parent(s) has spilled
	// too more (since child can't spill already spilled pages). So there is no
	// universal golden ratio.
	//
	// Should be in the range 0..255, where zero means no explicit spilling will
	// be performed during starting nested transactions.
	// Default is 0, i.e. by default no spilling performed during starting nested
	// transactions, that correspond historically behaviour.
	OptSpillParent4ChildDenominator = Opt(C.MDBX_opt_spill_parent4child_denominator)

	// OptMergeThreshold16Dot16Percent brief Controls the in-process threshold of semi-empty pages merge.
	// warning This is experimental option and subject for change or removal.
	// details This option controls the in-process threshold of minimum page
	// fill, as used space of percentage of a page. Neighbour pages emptier than
	// this value are candidates for merging. The threshold value is specified
	// in 1/65536 of percent, which is equivalent to the 16-dot-16 fixed point
	// format. The specified value must be in the range from 12.5% (almost empty)
	// to 50% (half empty) which corresponds to the range from 8192 and to 32768
	// in units respectively.
	OptMergeThreshold16Dot16Percent = Opt(C.MDBX_opt_merge_threshold_16dot16_percent)
)

type DeleteMode int32

const (
	// DeleteModeJustDelete brief Just delete the environment's files and directory if any.
	// note On POSIX systems, processes already working with the database will
	// continue to work without interference until it close the environment.
	// note On Windows, the behavior of `MDB_ENV_JUST_DELETE` is different
	// because the system does not support deleting files that are currently
	// memory mapped.
	DeleteModeJustDelete = DeleteMode(C.MDBX_ENV_JUST_DELETE)

	// DeleteModeEnsureUnused brief Make sure that the environment is not being used by other
	// processes, or return an error otherwise.
	DeleteModeEnsureUnused = DeleteMode(C.MDBX_ENV_ENSURE_UNUSED)

	// DeleteModeWaitForUnused brief Wait until other processes closes the environment before deletion.
	DeleteModeWaitForUnused = DeleteMode(C.MDBX_ENV_WAIT_FOR_UNUSED)
)

type DBIState uint32

const (
	DBIStateDirty = DBIState(C.MDBX_DBI_DIRTY) // DB was written in this txn
	DBIStateState = DBIState(C.MDBX_DBI_STALE) // Named-DB record is older than txnID
	DBIStateFresh = DBIState(C.MDBX_DBI_FRESH) // Named-DB handle opened in this txn
	DBIStateCreat = DBIState(C.MDBX_DBI_CREAT) // Named-DB handle created in this txn
)

// Delete brief Delete the environment's files in a proper and multiprocess-safe way.
// ingroup c_extra
//
// param [in] pathname  The pathname for the database or the directory in which
//
//	the database files reside.
//
// param [in] mode      Special deletion mode for the environment. This
//
//	parameter must be set to one of the values described
//	above in the ref MDBX_env_delete_mode_t section.
//
// note The ref MDBX_ENV_JUST_DELETE don't supported on Windows since system
// unable to delete a memory-mapped files.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_RESULT_TRUE   No corresponding files or directories were found,
//
//	so no deletion was performed.
func Delete(path string, mode DeleteMode) Error {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	return Error(C.mdbx_env_delete(p, (C.MDBX_env_delete_mode_t)(mode)))
}

type Env struct {
	env    *C.MDBX_env
	opened int64
	info   EnvInfo
	closed int64
	mu     sync.Mutex
}

// NewEnv brief Create an MDBX environment instance.
// ingroup c_opening
//
// This function allocates memory for a ref MDBX_env structure. To release
// the allocated memory and discard the handle, call ref mdbx_env_close().
// Before the handle may be used, it must be opened using ref mdbx_env_open().
//
// Various other options may also need to be set before opening the handle,
// e.g. ref mdbx_env_set_geometry(), ref mdbx_env_set_maxreaders(),
// ref mdbx_env_set_maxdbs(), depending on usage requirements.
//
// param [out] penv  The address where the new handle will be stored.
//
// returns a non-zero error value on failure and 0 on success.
func NewEnv() (*Env, Error) {
	env := &Env{}
	err := Error(C.mdbx_env_create((**C.MDBX_env)(unsafe.Pointer(&env.env))))
	if err != ErrSuccess {
		return nil, err
	}
	return env, err
}

// FD returns the open file descriptor (or Windows file handle) for the given
// environment.  An error is returned if the environment has not been
// successfully Opened (where C API just retruns an invalid handle).
//
// See mdbx_env_get_fd.
func (env *Env) FD() (uintptr, error) {
	// fdInvalid is the value -1 as a uintptr, which is used by MDBX in the
	// case that env has not been opened yet.  the strange construction is done
	// to avoid constant value overflow errors at compile time.
	const fdInvalid = ^uintptr(0)

	var mf C.mdbx_filehandle_t
	err := Error(C.mdbx_env_get_fd(env.env, &mf))
	//err := operrno("mdbx_env_get_fd", ret)
	if err != ErrSuccess {
		return 0, err
	}
	fd := uintptr(mf)

	if fd == fdInvalid {
		return 0, os.ErrClosed
	}
	return fd, nil
}

// ReaderList dumps the contents of the reader lock table as text.  Readers
// start on the second line as space-delimited fields described by the first
// line.
//
// See mdbx_reader_list.
//func (env *Env) ReaderList(fn func(string) error) error {
//	ctx, done := newMsgFunc(fn)
//	defer done()
//	if fn == nil {
//		ctx = 0
//	}
//
//	ret := C.mdbxgo_reader_list(env._env, C.size_t(ctx))
//	if ret >= 0 {
//		return nil
//	}
//	if ret < 0 && ctx != 0 {
//		err := ctx.get().err
//		if err != nil {
//			return err
//		}
//	}
//	return operrno("mdbx_reader_list", ret)
//}

// ReaderCheck clears stale entries from the reader lock table and returns the
// number of entries cleared.
//
// See mdbx_reader_check()
func (env *Env) ReaderCheck() (int, error) {
	var dead C.int
	err := Error(C.mdbx_reader_check(env.env, &dead))
	if err != ErrSuccess {
		return int(dead), err
	}
	return int(dead), nil
}

// Path returns the path argument passed to Open.  Path returns a non-nil error
// if env.Open() was not previously called.
//
// See mdbx_env_get_path.
func (env *Env) Path() (string, error) {
	var cpath *C.char
	err := Error(C.mdbx_env_get_path(env.env, &cpath))
	if err != ErrSuccess {
		return "", err
	}
	if cpath == nil {
		return "", os.ErrNotExist
	}
	return C.GoString(cpath), nil
}

// MaxKeySize returns the maximum allowed length for a key.
//
// See mdbx_env_get_maxkeysize.
func (env *Env) MaxKeySize() int {
	if env == nil {
		return int(C.mdbx_env_get_maxkeysize_ex(nil, 0))
	}
	return int(C.mdbx_env_get_maxkeysize_ex(env.env, 0))
}

// Close the environment and release the memory map.
// ingroup c_opening
//
// Only a single thread may call this function. All transactions, databases,
// and cursors must already be closed before calling this function. Attempts
// to use any such handles after calling this function will cause a `SIGSEGV`.
// The environment handle will be freed and must not be used again after this
// call.
//
// param [in] env        An environment handle returned by
//
//	ref mdbx_env_create().
//
// param [in] dont_sync  A dont'sync flag, if non-zero the last checkpoint
//
//	will be kept "as is" and may be still "weak" in the
//	ref MDBX_SAFE_NOSYNC or ref MDBX_UTTERLY_NOSYNC
//	modes. Such "weak" checkpoint will be ignored on
//	opening next time, and transactions since the last
//	non-weak checkpoint (meta-page update) will rolledback
//	for consistency guarantee.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_BUSY   The write transaction is running by other thread,
//
//	in such case ref MDBX_env instance has NOT be destroyed
//	not released!
//	note If any OTHER error code was returned then
//	given MDBX_env instance has been destroyed and released.
//
// retval MDBX_EBADSIGN  Environment handle already closed or not valid,
//
//	i.e. ref mdbx_env_close() was already called for the
//	`env` or was not created by ref mdbx_env_create().
//
// retval MDBX_PANIC  If ref mdbx_env_close_ex() was called in the child
//
//	process after `fork()`. In this case ref MDBX_PANIC
//	is expected, i.e. ref MDBX_env instance was freed in
//	proper manner.
//
// retval MDBX_EIO    An error occurred during synchronization.
func (env *Env) Close(dontSync bool) Error {
	env.mu.Lock()
	defer env.mu.Unlock()
	if env.closed > 0 {
		return ErrSuccess
	}
	err := Error(C.mdbx_env_close_ex(env.env, (C.bool)(dontSync)))
	if err != ErrSuccess {
		return err
	}
	env.closed = time.Now().UnixNano()
	return err
}

// SetFlags Set environment flags.
// ingroup c_settings
//
// This may be used to set some flags in addition to those from
// mdbx_env_open(), or to unset these flags.
// see mdbx_env_get_flags()
//
// note In contrast to LMDB, the MDBX serialize threads via mutex while
// changing the flags. Therefore this function will be blocked while a write
// transaction running by other thread, or ref MDBX_BUSY will be returned if
// function called within a write transaction.
//
// param [in] env      An environment handle returned
//
//	by ref mdbx_env_create().
//
// param [in] flags    The ref env_flags to change, bitwise OR'ed together.
// param [in] onoff    A non-zero value sets the flags, zero clears them.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_EINVAL  An invalid parameter was specified.
func (env *Env) SetFlags(flags EnvFlags, onoff bool) Error {
	return Error(C.mdbx_env_set_flags(env.env, (C.MDBX_env_flags_t)(flags), (C.bool)(onoff)))
}

// GetFlags Get environment flags.
// ingroup c_statinfo
// see mdbx_env_set_flags()
//
// param [in] env     An environment handle returned by ref mdbx_env_create().
// param [out] flags  The address of an integer to store the flags.
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_EINVAL An invalid parameter was specified.
func (env *Env) GetFlags() (EnvFlags, Error) {
	flags := C.unsigned(0)
	err := Error(C.mdbx_env_get_flags(env.env, &flags))
	return EnvFlags(flags), err
}

// Copy an MDBX environment to the specified path, with options.
// ingroup c_extra
//
// This function may be used to make a backup of an existing environment.
// No lockfile is created, since it gets recreated at need.
// note This call can trigger significant file size growth if run in
// parallel with write transactions, because it employs a read-only
// transaction. See long-lived transactions under ref restrictions section.
//
// param [in] env    An environment handle returned by mdbx_env_create().
//
//	It must have already been opened successfully.
//
// param [in] dest   The pathname of a file in which the copy will reside.
//
//	This file must not be already exist, but parent directory
//	must be writable.
//
// param [in] flags  Special options for this operation. This parameter must
//
//	                  be set to 0 or by bitwise OR'ing together one or more
//	                  of the values described here:
//
//	- ref MDBX_CP_COMPACT
//	    Perform compaction while copying: omit free pages and sequentially
//	    renumber all pages in output. This option consumes little bit more
//	    CPU for processing, but may running quickly than the default, on
//	    account skipping free pages.
//
//	- ref MDBX_CP_FORCE_DYNAMIC_SIZE
//	    Force to make resizeable copy, i.e. dynamic size instead of fixed.
//
// returns A non-zero error value on failure and 0 on success.
func (env *Env) Copy(dest string, flags CopyFlags) Error {
	if env.env == nil {
		return ErrSuccess
	}
	d := C.CString(dest)
	defer C.free(unsafe.Pointer(d))
	return Error(C.mdbx_env_copy(env.env, d, (C.MDBX_copy_flags_t)(flags)))
}

// Open brief Open an environment instance.
// ingroup c_opening
//
// Indifferently this function will fails or not, the ref mdbx_env_close() must
// be called later to discard the ref MDBX_env handle and release associated
// resources.
//
// param [in] env       An environment handle returned
//
//	by ref mdbx_env_create()
//
// param [in] pathname  The pathname for the database or the directory in which
//
//	the database files reside. In the case of directory it
//	must already exist and be writable.
//
// param [in] flags     Special options for this environment. This parameter
//
//	must be set to 0 or by bitwise OR'ing together one
//	or more of the values described above in the
//	ref env_flags and ref sync_modes sections.
//
// Flags set by mdbx_env_set_flags() are also used:
//
//   - ref MDBX_NOSUBDIR, ref MDBX_RDONLY, ref MDBX_EXCLUSIVE,
//     ref MDBX_WRITEMAP, ref MDBX_NOTLS, ref MDBX_NORDAHEAD,
//     ref MDBX_NOMEMINIT, ref MDBX_COALESCE, ref MDBX_LIFORECLAIM.
//     See ref env_flags section.
//
//   - ref MDBX_NOMETASYNC, ref MDBX_SAFE_NOSYNC, ref MDBX_UTTERLY_NOSYNC.
//     See ref sync_modes section.
//
// note `MDB_NOLOCK` flag don't supported by MDBX,
//
//	try use ref MDBX_EXCLUSIVE as a replacement.
//
// note MDBX don't allow to mix processes with different ref MDBX_SAFE_NOSYNC
//
//	flags on the same environment.
//	In such case ref MDBX_INCOMPATIBLE will be returned.
//
// If the database is already exist and parameters specified early by
// ref mdbx_env_set_geometry() are incompatible (i.e. for instance, different
// page size) then ref mdbx_env_open() will return ref MDBX_INCOMPATIBLE
// error.
//
// param [in] mode   The UNIX permissions to set on created files.
//
//	Zero value means to open existing, but do not create.
//
// return A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_VERSION_MISMATCH The version of the MDBX library doesn't match
//
//	the version that created the database environment.
//
// retval MDBX_INVALID       The environment file headers are corrupted.
// retval MDBX_ENOENT        The directory specified by the path parameter
//
//	doesn't exist.
//
// retval MDBX_EACCES        The user didn't have permission to access
//
//	the environment files.
//
// retval MDBX_EAGAIN        The environment was locked by another process.
// retval MDBX_BUSY          The ref MDBX_EXCLUSIVE flag was specified and the
//
//	environment is in use by another process,
//	or the current process tries to open environment
//	more than once.
//
// retval MDBX_INCOMPATIBLE  Environment is already opened by another process,
//
//	but with different set of ref MDBX_SAFE_NOSYNC,
//	ref MDBX_UTTERLY_NOSYNC flags.
//	Or if the database is already exist and parameters
//	specified early by ref mdbx_env_set_geometry()
//	are incompatible (i.e. different pagesize, etc).
//
// retval MDBX_WANNA_RECOVERY The ref MDBX_RDONLY flag was specified but
//
//	read-write access is required to rollback
//	inconsistent state after a system crash.
//
// retval MDBX_TOO_LARGE      Database is too large for this process,
//
//	i.e. 32-bit process tries to open >4Gb database.
func (env *Env) Open(path string, flags EnvFlags, mode os.FileMode) Error {
	if env.opened > 0 {
		return ErrSuccess
	}

	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))

	err := Error(C.mdbx_env_open(
		(*C.MDBX_env)(unsafe.Pointer(env.env)),
		p,
		(C.MDBX_env_flags_t)(flags),
		(C.mdbx_mode_t)(mode),
	))
	if err != ErrSuccess {
		return err
	}

	env.opened = time.Now().UnixNano()
	return err
}

type Geometry struct {
	env             uintptr
	SizeLower       uintptr
	SizeNow         uintptr
	SizeUpper       uintptr
	GrowthStep      uintptr
	ShrinkThreshold uintptr
	PageSize        uintptr
	err             Error
}

// SetGeometry Set all size-related parameters of environment, including page size
// and the min/max size of the memory map. ingroup c_settings
//
// In contrast to LMDB, the MDBX provide automatic size management of an
// database according the given parameters, including shrinking and resizing
// on the fly. From user point of view all of these just working. Nevertheless,
// it is reasonable to know some details in order to make optimal decisions
// when choosing parameters.
//
// Both ref mdbx_env_info_ex() and legacy ref mdbx_env_info() are inapplicable
// to read-only opened environment.
//
// Both ref mdbx_env_info_ex() and legacy ref mdbx_env_info() could be called
// either before or after ref mdbx_env_open(), either within the write
// transaction running by current thread or not:
//
//   - In case ref mdbx_env_info_ex() or legacy ref mdbx_env_info() was called
//     BEFORE ref mdbx_env_open(), i.e. for closed environment, then the
//     specified parameters will be used for new database creation, or will be
//     applied during opening if database exists and no other process using it.
//
//     If the database is already exist, opened with ref MDBX_EXCLUSIVE or not
//     used by any other process, and parameters specified by
//     ref mdbx_env_set_geometry() are incompatible (i.e. for instance,
//     different page size) then ref mdbx_env_open() will return
//     ref MDBX_INCOMPATIBLE error.
//
//     In another way, if database will opened read-only or will used by other
//     process during calling ref mdbx_env_open() that specified parameters will
//     silently discarded (open the database with ref MDBX_EXCLUSIVE flag
//     to avoid this).
//
//   - In case ref mdbx_env_info_ex() or legacy ref mdbx_env_info() was called
//     after ref mdbx_env_open() WITHIN the write transaction running by current
//     thread, then specified parameters will be applied as a part of write
//     transaction, i.e. will not be visible to any others processes until the
//     current write transaction has been committed by the current process.
//     However, if transaction will be aborted, then the database file will be
//     reverted to the previous size not immediately, but when a next transaction
//     will be committed or when the database will be opened next time.
//
//   - In case ref mdbx_env_info_ex() or legacy ref mdbx_env_info() was called
//     after ref mdbx_env_open() but OUTSIDE a write transaction, then MDBX will
//     execute internal pseudo-transaction to apply new parameters (but only if
//     anything has been changed), and changes be visible to any others processes
//     immediately after succesful completion of function.
//
// Essentially a concept of "automatic size management" is simple and useful:
//   - There are the lower and upper bound of the database file size;
//   - There is the growth step by which the database file will be increased,
//     in case of lack of space.
//   - There is the threshold for unused space, beyond which the database file
//     will be shrunk.
//   - The size of the memory map is also the maximum size of the database.
//   - MDBX will automatically manage both the size of the database and the size
//     of memory map, according to the given parameters.
//
// So, there some considerations about choosing these parameters:
//   - The lower bound allows you to prevent database shrinking below some
//     rational size to avoid unnecessary resizing costs.
//   - The upper bound allows you to prevent database growth above some rational
//     size. Besides, the upper bound defines the linear address space
//     reservation in each process that opens the database. Therefore changing
//     the upper bound is costly and may be required reopening environment in
//     case of ref MDBX_UNABLE_EXTEND_MAPSIZE errors, and so on. Therefore, this
//     value should be chosen reasonable as large as possible, to accommodate
//     future growth of the database.
//   - The growth step must be greater than zero to allow the database to grow,
//     but also reasonable not too small, since increasing the size by little
//     steps will result a large overhead.
//   - The shrink threshold must be greater than zero to allow the database
//     to shrink but also reasonable not too small (to avoid extra overhead) and
//     not less than growth step to avoid up-and-down flouncing.
//   - The current size (i.e. size_now argument) is an auxiliary parameter for
//     simulation legacy ref mdbx_env_set_mapsize() and as workaround Windows
//     issues (see below).
//
// Unfortunately, Windows has is a several issues
// with resizing of memory-mapped file:
//   - Windows unable shrinking a memory-mapped file (i.e memory-mapped section)
//     in any way except unmapping file entirely and then map again. Moreover,
//     it is impossible in any way if a memory-mapped file is used more than
//     one process.
//   - Windows does not provide the usual API to augment a memory-mapped file
//     (that is, a memory-mapped partition), but only by using "Native API"
//     in an undocumented way.
//
// MDBX bypasses all Windows issues, but at a cost:
//   - Ability to resize database on the fly requires an additional lock
//     and release `SlimReadWriteLock during` each read-only transaction.
//   - During resize all in-process threads should be paused and then resumed.
//   - Shrinking of database file is performed only when it used by single
//     process, i.e. when a database closes by the last process or opened
//     by the first.
//     = Therefore, the size_now argument may be useful to set database size
//     by the first process which open a database, and thus avoid expensive
//     remapping further.
//
// For create a new database with particular parameters, including the page
// size, ref mdbx_env_set_geometry() should be called after
// ref mdbx_env_create() and before mdbx_env_open(). Once the database is
// created, the page size cannot be changed. If you do not specify all or some
// of the parameters, the corresponding default values will be used. For
// instance, the default for database size is 10485760 bytes.
//
// If the mapsize is increased by another process, MDBX silently and
// transparently adopt these changes at next transaction start. However,
// ref mdbx_txn_begin() will return ref MDBX_UNABLE_EXTEND_MAPSIZE if new
// mapping size could not be applied for current process (for instance if
// address space is busy).  Therefore, in the case of
// ref MDBX_UNABLE_EXTEND_MAPSIZE error you need close and reopen the
// environment to resolve error.
//
// note Actual values may be different than your have specified because of
// rounding to specified database page size, the system page size and/or the
// size of the system virtual memory management unit. You can get actual values
// by ref mdbx_env_sync_ex() or see by using the tool `mdbx_chk` with the `-v`
// option.
//
// Legacy ref mdbx_env_set_mapsize() correspond to calling
// ref mdbx_env_set_geometry() with the arguments `size_lower`, `size_now`,
// `size_upper` equal to the `size` and `-1` (i.e. default) for all other
// parameters.
//
// param [in] env         An environment handle returned
//
//	by ref mdbx_env_create()
//
// param [in] size_lower  The lower bound of database size in bytes.
//
//	Zero value means "minimal acceptable",
//	and negative means "keep current or use default".
//
// param [in] size_now    The size in bytes to setup the database size for
//
//	now. Zero value means "minimal acceptable", and
//	negative means "keep current or use default". So,
//	it is recommended always pass -1 in this argument
//	except some special cases.
//
// param [in] size_upper The upper bound of database size in bytes.
//
//	Zero value means "minimal acceptable",
//	and negative means "keep current or use default".
//	It is recommended to avoid change upper bound while
//	database is used by other processes or threaded
//	(i.e. just pass -1 in this argument except absolutely
//	necessary). Otherwise you must be ready for
//	ref MDBX_UNABLE_EXTEND_MAPSIZE error(s), unexpected
//	pauses during remapping and/or system errors like
//	"address busy", and so on. In other words, there
//	is no way to handle a growth of the upper bound
//	robustly because there may be a lack of appropriate
//	system resources (which are extremely volatile in
//	a multi-process multi-threaded environment).
//
// param [in] growth_step  The growth step in bytes, must be greater than
//
//	zero to allow the database to grow. Negative value
//	means "keep current or use default".
//
// param [in] shrink_threshold  The shrink threshold in bytes, must be greater
//
//	than zero to allow the database to shrink and
//	greater than growth_step to avoid shrinking
//	right after grow.
//	Negative value means "keep current
//	or use default". Default is 2*growth_step.
//
// param [in] pagesize          The database page size for new database
//
//	creation or -1 otherwise. Must be power of 2
//	in the range between ref MDBX_MIN_PAGESIZE and
//	ref MDBX_MAX_PAGESIZE. Zero value means
//	"minimal acceptable", and negative means
//	"keep current or use default".
//
// returns A non-zero error value on failure and 0 on success,
//
//	some possible errors are:
//
// retval MDBX_EINVAL    An invalid parameter was specified,
//
//	or the environment has an active write transaction.
//
// retval MDBX_EPERM     Specific for Windows: Shrinking was disabled before
//
//	and now it wanna be enabled, but there are reading
//	threads that don't use the additional `SRWL` (that
//	is required to avoid Windows issues).
//
// retval MDBX_EACCESS   The environment opened in read-only.
// retval MDBX_MAP_FULL  Specified size smaller than the space already
//
//	consumed by the environment.
//
// retval MDBX_TOO_LARGE Specified size is too large, i.e. too many pages for
//
//	given size, or a 32-bit process requests too much
//	bytes for the 32-bit address space.
func (env *Env) SetGeometry(args Geometry) Error {
	args.env = uintptr(unsafe.Pointer(env.env))
	ptr := uintptr(unsafe.Pointer(&args))
	unsafecgo.NonBlocking((*byte)(C.do_mdbx_env_set_geometry), ptr, 0)
	return args.err
}

// GetOption brief Gets the value of runtime options from an environment.
// ingroup c_settings
//
// param [in] env     An environment handle returned by ref mdbx_env_create().
// param [in] option  The option from ref MDBX_option_t to get value of it.
// param [out] pvalue The address where the option's value will be stored.
//
// see MDBX_option_t
// see mdbx_env_get_option()
// returns A non-zero error value on failure and 0 on success.
func (env *Env) GetOption(option Opt) (uint64, Error) {
	value := uint64(0)
	err := Error(C.mdbx_env_get_option(
		(*C.MDBX_env)(unsafe.Pointer(env.env)),
		(C.MDBX_option_t)(option),
		(*C.uint64_t)(unsafe.Pointer(&value))),
	)
	return value, err
}

// SetOption brief Sets the value of a runtime options for an environment.
// ingroup c_settings
//
// param [in] env     An environment handle returned by ref mdbx_env_create().
// param [in] option  The option from ref MDBX_option_t to set value of it.
// param [in] value   The value of option to be set.
//
// see MDBX_option_t
// see mdbx_env_get_option()
// returns A non-zero error value on failure and 0 on success.
func (env *Env) SetOption(option Opt, value uint64) Error {
	return Error(C.mdbx_env_set_option(
		(*C.MDBX_env)(unsafe.Pointer(env.env)),
		(C.MDBX_option_t)(option),
		C.uint64_t(value)),
	)
}

// Sync Flush the environment data buffers to disk.
// ingroup c_extra
//
// Unless the environment was opened with no-sync flags (ref MDBX_NOMETASYNC,
// ref MDBX_SAFE_NOSYNC and ref MDBX_UTTERLY_NOSYNC), then
// data is always written an flushed to disk when ref mdbx_txn_commit() is
// called. Otherwise ref mdbx_env_sync() may be called to manually write and
// flush unsynced data to disk.
//
// Besides, ref mdbx_env_sync_ex() with argument `force=false` may be used to
// provide polling mode for lazy/asynchronous sync in conjunction with
// ref mdbx_env_set_syncbytes() and/or ref mdbx_env_set_syncperiod().
//
// note This call is not valid if the environment was opened with MDBX_RDONLY.
//
// param [in] env      An environment handle returned by ref mdbx_env_create()
// param [in] force    If non-zero, force a flush. Otherwise, If force is
//
//	zero, then will run in polling mode,
//	i.e. it will check the thresholds that were
//	set ref mdbx_env_set_syncbytes()
//	and/or ref mdbx_env_set_syncperiod() and perform flush
//	if at least one of the thresholds is reached.
//
// param [in] nonblock Don't wait if write transaction
//
//	is running by other thread.
//
// returns A non-zero error value on failure and ref MDBX_RESULT_TRUE or 0 on
//
//	success. The ref MDBX_RESULT_TRUE means no data pending for flush
//	to disk, and 0 otherwise. Some possible errors are:
//
// retval MDBX_EACCES   the environment is read-only.
// retval MDBX_BUSY     the environment is used by other thread
//
//	and `nonblock=true`.
//
// retval MDBX_EINVAL   an invalid parameter was specified.
// retval MDBX_EIO      an error occurred during synchronization.
func (env *Env) Sync(force, nonblock bool) Error {
	return Error(C.mdbx_env_sync_ex(env.env, (C.bool)(force), (C.bool)(nonblock)))
}

// CloseDBI Close a database handle. Normally unnecessary.
// ingroup c_dbi
//
// Closing a database handle is not necessary, but lets ref mdbx_dbi_open()
// reuse the handle value. Usually it's better to set a bigger
// ref mdbx_env_set_maxdbs(), unless that value would be large.
//
// note Use with care.
// This call is synchronized via mutex with ref mdbx_dbi_close(), but NOT with
// other transactions running by other threads. The "next" version of libmdbx
// (ref MithrilDB) will solve this issue.
//
// Handles should only be closed if no other threads are going to reference
// the database handle or one of its cursors any further. Do not close a handle
// if an existing transaction has modified its database. Doing so can cause
// misbehavior from database corruption to errors like ref MDBX_BAD_DBI
// (since the DB name is gone).
//
// param [in] env  An environment handle returned by ref mdbx_env_create().
// param [in] dbi  A database handle returned by ref mdbx_dbi_open().
//
// returns A non-zero error value on failure and 0 on success.
func (env *Env) CloseDBI(dbi DBI) Error {
	return Error(C.mdbx_dbi_close(env.env, (C.MDBX_dbi)(dbi)))
}

// GetMaxDBS Controls the maximum number of named databases for the environment.
//
// details By default only unnamed key-value database could used and
// appropriate value should set by `MDBX_opt_max_db` to using any more named
// subDB(s). To reduce overhead, use the minimum sufficient value. This option
// may only set after ref mdbx_env_create() and before ref mdbx_env_open().
//
// see mdbx_env_set_maxdbs() see mdbx_env_get_maxdbs()
func (env *Env) GetMaxDBS() (uint64, Error) {
	return env.GetOption(OptMaxDB)
}

// SetMaxDBS Controls the maximum number of named databases for the environment.
//
// details By default only unnamed key-value database could used and
// appropriate value should set by `MDBX_opt_max_db` to using any more named
// subDB(s). To reduce overhead, use the minimum sufficient value. This option
// may only set after ref mdbx_env_create() and before ref mdbx_env_open().
//
// see mdbx_env_set_maxdbs() see mdbx_env_get_maxdbs()
func (env *Env) SetMaxDBS(max uint16) Error {
	return env.SetOption(OptMaxDB, uint64(max))
}

// GetMaxReaders Defines the maximum number of threads/reader slots
// for all processes interacting with the database.
//
// details This defines the number of slots in the lock table that is used to
// track readers in the the environment. The default is about 100 for 4K
// system page size. Starting a read-only transaction normally ties a lock
// table slot to the current thread until the environment closes or the thread
// exits. If ref MDBX_NOTLS is in use, ref mdbx_txn_begin() instead ties the
// slot to the ref MDBX_txn object until it or the ref MDBX_env object is
// destroyed. This option may only set after ref mdbx_env_create() and before
// ref mdbx_env_open(), and has an effect only when the database is opened by
// the first process interacts with the database.
//
// see mdbx_env_set_maxreaders() see mdbx_env_get_maxreaders()
func (env *Env) GetMaxReaders() (uint64, Error) {
	return env.GetOption(OptMaxReaders)
}

// SetMaxReaders Defines the maximum number of threads/reader slots
// for all processes interacting with the database.
//
// details This defines the number of slots in the lock table that is used to
// track readers in the the environment. The default is about 100 for 4K
// system page size. Starting a read-only transaction normally ties a lock
// table slot to the current thread until the environment closes or the thread
// exits. If ref MDBX_NOTLS is in use, ref mdbx_txn_begin() instead ties the
// slot to the ref MDBX_txn object until it or the ref MDBX_env object is
// destroyed. This option may only set after ref mdbx_env_create() and before
// ref mdbx_env_open(), and has an effect only when the database is opened by
// the first process interacts with the database.
//
// see mdbx_env_set_maxreaders() see mdbx_env_get_maxreaders()
func (env *Env) SetMaxReaders(max uint64) Error {
	return env.SetOption(OptMaxReaders, max)
}

// GetSyncBytes Controls interprocess/shared threshold to force flush the data
// buffers to disk, if ref MDBX_SAFE_NOSYNC is used.
//
// see mdbx_env_set_syncbytes() see mdbx_env_get_syncbytes()
func (env *Env) GetSyncBytes() (uint64, Error) {
	return env.GetOption(OptSyncBytes)
}

// SetSyncBytes Controls interprocess/shared threshold to force flush the data
// buffers to disk, if ref MDBX_SAFE_NOSYNC is used.
//
// see mdbx_env_set_syncbytes() see mdbx_env_get_syncbytes()
func (env *Env) SetSyncBytes(bytes uint64) Error {
	return env.SetOption(OptSyncBytes, bytes)
}

// GetSyncPeriod Controls interprocess/shared relative period since the last
// unsteady commit to force flush the data buffers to disk,
// if ref MDBX_SAFE_NOSYNC is used.
// see mdbx_env_set_syncperiod() see mdbx_env_get_syncperiod()
func (env *Env) GetSyncPeriod() (uint64, Error) {
	return env.GetOption(OptSyncPeriod)
}

// SetSyncPeriod Controls interprocess/shared relative period since the last
// unsteady commit to force flush the data buffers to disk,
// if ref MDBX_SAFE_NOSYNC is used.
// see mdbx_env_set_syncperiod() see mdbx_env_get_syncperiod()
func (env *Env) SetSyncPeriod(period uint64) Error {
	return env.SetOption(OptSyncPeriod, period)
}

// GetRPAugmentLimit Controls the in-process limit to grow a list of reclaimed/recycled
// page's numbers for finding a sequence of contiguous pages for large data
// items.
//
// details A long values requires allocation of contiguous database pages.
// To find such sequences, it may be necessary to accumulate very large lists,
// especially when placing very long values (more than a megabyte) in a large
// databases (several tens of gigabytes), which is much expensive in extreme
// cases. This threshold allows you to avoid such costs by allocating new
// pages at the end of the database (with its possible growth on disk),
// instead of further accumulating/reclaiming Garbage Collection records.
//
// On the other hand, too small threshold will lead to unreasonable database
// growth, or/and to the inability of put long values.
//
// The `MDBX_opt_rp_augment_limit` controls described limit for the current
// process. Default is 262144, it is usually enough for most cases.
func (env *Env) GetRPAugmentLimit() (uint64, Error) {
	return env.GetOption(OptRpAugmentLimit)
}

// SetRPAugmentLimit Controls the in-process limit to grow a list of reclaimed/recycled
// page's numbers for finding a sequence of contiguous pages for large data
// items.
//
// details A long values requires allocation of contiguous database pages.
// To find such sequences, it may be necessary to accumulate very large lists,
// especially when placing very long values (more than a megabyte) in a large
// databases (several tens of gigabytes), which is much expensive in extreme
// cases. This threshold allows you to avoid such costs by allocating new
// pages at the end of the database (with its possible growth on disk),
// instead of further accumulating/reclaiming Garbage Collection records.
//
// On the other hand, too small threshold will lead to unreasonable database
// growth, or/and to the inability of put long values.
//
// The `MDBX_opt_rp_augment_limit` controls described limit for the current
// process. Default is 262144, it is usually enough for most cases.
func (env *Env) SetRPAugmentLimit(limit uint64) Error {
	return env.SetOption(OptRpAugmentLimit, limit)
}

// GetLooseLimit Controls the in-process limit to grow a cache of dirty
// pages for reuse in the current transaction.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// To reduce overhead, it is reasonable to release not all such pages
// immediately, but to leave some ones in cache for reuse in the current
// transaction.
//
// The `MDBX_opt_loose_limit` allows you to set a limit for such cache inside
// the current process. Should be in the range 0..255, default is 64.
func (env *Env) GetLooseLimit() (uint64, Error) {
	return env.GetOption(OptLooseLimit)
}

// SetLooseLimit Controls the in-process limit to grow a cache of dirty
// pages for reuse in the current transaction.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// To reduce overhead, it is reasonable to release not all such pages
// immediately, but to leave some ones in cache for reuse in the current
// transaction.
//
// The `MDBX_opt_loose_limit` allows you to set a limit for such cache inside
// the current process. Should be in the range 0..255, default is 64.
func (env *Env) SetLooseLimit(limit uint64) Error {
	return env.SetOption(OptLooseLimit, limit)
}

// GetDPReserveLimit Controls the in-process limit of a pre-allocated memory items
// for dirty pages.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and
// released when a transaction is committed. To reduce overhead, it is
// reasonable to release not all ones, but to leave some allocations in
// reserve for reuse in the next transaction(s).
//
// The `MDBX_opt_dp_reserve_limit` allows you to set a limit for such reserve
// inside the current process. Default is 1024.
func (env *Env) GetDPReserveLimit() (uint64, Error) {
	return env.GetOption(OptDpReserveLimit)
}

// SetDPReserveLimit Controls the in-process limit of a pre-allocated memory items
// for dirty pages.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and
// released when a transaction is committed. To reduce overhead, it is
// reasonable to release not all ones, but to leave some allocations in
// reserve for reuse in the next transaction(s).
//
// The `MDBX_opt_dp_reserve_limit` allows you to set a limit for such reserve
// inside the current process. Default is 1024.
func (env *Env) SetDPReserveLimit(limit uint64) Error {
	return env.SetOption(OptDpReserveLimit, limit)
}

// GetTxDPLimit Controls the in-process limit of dirty pages
// for a write transaction.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and will
// be busy until are written to disk. Therefore for a large transactions is
// reasonable to limit dirty pages collecting above an some threshold but
// spill to disk instead.
//
// The `MDBX_opt_txn_dp_limit` controls described threshold for the current
// process. Default is 65536, it is usually enough for most cases.
func (env *Env) GetTxDPLimit() (uint64, Error) {
	return env.GetOption(OptTxnDpLimit)
}

// SetTxDPLimit Controls the in-process limit of dirty pages
// for a write transaction.
//
// details A 'dirty page' refers to a page that has been updated in memory
// only, the changes to a dirty page are not yet stored on disk.
// Without ref MDBX_WRITEMAP dirty pages are allocated from memory and will
// be busy until are written to disk. Therefore for a large transactions is
// reasonable to limit dirty pages collecting above an some threshold but
// spill to disk instead.
//
// The `MDBX_opt_txn_dp_limit` controls described threshold for the current
// process. Default is 65536, it is usually enough for most cases.
func (env *Env) SetTxDPLimit(limit uint64) Error {
	return env.SetOption(OptTxnDpLimit, limit)
}

// GetTxDPInitial Controls the in-process initial allocation size for dirty pages
// list of a write transaction. Default is 1024.
func (env *Env) GetTxDPInitial() (uint64, Error) {
	return env.GetOption(OptTxnDpInitial)
}

// SetTxDPInitial Controls the in-process initial allocation size for dirty pages
// list of a write transaction. Default is 1024.
func (env *Env) SetTxDPInitial(initial uint64) Error {
	return env.SetOption(OptTxnDpInitial, initial)
}

// GetSpillMinDenominator Controls the in-process how minimal part of the dirty pages should
// be spilled when necessary.
//
// details The `MDBX_opt_spill_min_denominator` defines the denominator for
// limiting from the bottom for part of the current dirty pages should be
// spilled when the free room for a new dirty pages (i.e. distance to the
// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
// operation.
// Exactly `min_pages_to_spill = dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_min_denominator`.
//
// Should be in the range 0..255, where zero means no restriction at the
// bottom. Default is 8, i.e. at least the 1/8 of the current dirty pages
// should be spilled when reached the condition described above.
func (env *Env) GetSpillMinDenominator() (uint64, Error) {
	return env.GetOption(OptSpillMinDenomiator)
}

// SetSpillMinDenominator Controls the in-process how minimal part of the dirty pages should
// be spilled when necessary.
//
// details The `MDBX_opt_spill_min_denominator` defines the denominator for
// limiting from the bottom for part of the current dirty pages should be
// spilled when the free room for a new dirty pages (i.e. distance to the
// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
// operation.
// Exactly `min_pages_to_spill = dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_min_denominator`.
//
// Should be in the range 0..255, where zero means no restriction at the
// bottom. Default is 8, i.e. at least the 1/8 of the current dirty pages
// should be spilled when reached the condition described above.
func (env *Env) SetSpillMinDenominator(min uint64) Error {
	return env.SetOption(OptSpillMinDenomiator, min)
}

// GetSpillMaxDenominator Controls the in-process how maximal part of the dirty pages may be
// spilled when necessary.
//
// details The `MDBX_opt_spill_max_denominator` defines the denominator for
// limiting from the top for part of the current dirty pages may be spilled
// when the free room for a new dirty pages (i.e. distance to the
// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
// operation.
// Exactly `max_pages_to_spill = dirty_pages - dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_max_denominator`.
//
// Should be in the range 0..255, where zero means no limit, i.e. all dirty
// pages could be spilled. Default is 8, i.e. no more than 7/8 of the current
// dirty pages may be spilled when reached the condition described above.
func (env *Env) GetSpillMaxDenominator() (uint64, Error) {
	return env.GetOption(OptSpillMaxDenomiator)
}

// SetSpillMaxDenominator Controls the in-process how maximal part of the dirty pages may be
// spilled when necessary.
//
// details The `MDBX_opt_spill_max_denominator` defines the denominator for
// limiting from the top for part of the current dirty pages may be spilled
// when the free room for a new dirty pages (i.e. distance to the
// `MDBX_opt_txn_dp_limit` threshold) is not enough to perform requested
// operation.
// Exactly `max_pages_to_spill = dirty_pages - dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_max_denominator`.
//
// Should be in the range 0..255, where zero means no limit, i.e. all dirty
// pages could be spilled. Default is 8, i.e. no more than 7/8 of the current
// dirty pages may be spilled when reached the condition described above.
func (env *Env) SetSpillMaxDenominator(max uint64) Error {
	return env.SetOption(OptSpillMaxDenomiator, max)
}

// GetSpillParent4ChildDeominator Controls the in-process how much of the parent transaction dirty
// pages will be spilled while start each child transaction.
//
// details The `MDBX_opt_spill_parent4child_denominator` defines the
// denominator to determine how much of parent transaction dirty pages will be
// spilled explicitly while start each child transaction.
// Exactly `pages_to_spill = dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_parent4child_denominator`.
//
// For a stack of nested transactions each dirty page could be spilled only
// once, and parent's dirty pages couldn't be spilled while child
// transaction(s) are running. Therefore a child transaction could reach
// ref MDBX_TXN_FULL when parent(s) transaction has  spilled too less (and
// child reach the limit of dirty pages), either when parent(s) has spilled
// too more (since child can't spill already spilled pages). So there is no
// universal golden ratio.
//
// Should be in the range 0..255, where zero means no explicit spilling will
// be performed during starting nested transactions.
// Default is 0, i.e. by default no spilling performed during starting nested
// transactions, that correspond historically behaviour.
func (env *Env) GetSpillParent4ChildDeominator() (uint64, Error) {
	return env.GetOption(OptSpillParent4ChildDenominator)
}

// SetSpillParent4ChildDeominator Controls the in-process how much of the parent transaction dirty
// pages will be spilled while start each child transaction.
//
// details The `MDBX_opt_spill_parent4child_denominator` defines the
// denominator to determine how much of parent transaction dirty pages will be
// spilled explicitly while start each child transaction.
// Exactly `pages_to_spill = dirty_pages / N`,
// where `N` is the value set by `MDBX_opt_spill_parent4child_denominator`.
//
// For a stack of nested transactions each dirty page could be spilled only
// once, and parent's dirty pages couldn't be spilled while child
// transaction(s) are running. Therefore a child transaction could reach
// ref MDBX_TXN_FULL when parent(s) transaction has  spilled too less (and
// child reach the limit of dirty pages), either when parent(s) has spilled
// too more (since child can't spill already spilled pages). So there is no
// universal golden ratio.
//
// Should be in the range 0..255, where zero means no explicit spilling will
// be performed during starting nested transactions.
// Default is 0, i.e. by default no spilling performed during starting nested
// transactions, that correspond historically behaviour.
func (env *Env) SetSpillParent4ChildDeominator(value uint64) Error {
	return env.SetOption(OptSpillParent4ChildDenominator, value)
}

// GetMergeThreshold16Dot16Percent Controls the in-process threshold of semi-empty pages merge.
// warning This is experimental option and subject for change or removal.
// details This option controls the in-process threshold of minimum page
// fill, as used space of percentage of a page. Neighbour pages emptier than
// this value are candidates for merging. The threshold value is specified
// in 1/65536 of percent, which is equivalent to the 16-dot-16 fixed point
// format. The specified value must be in the range from 12.5% (almost empty)
// to 50% (half empty) which corresponds to the range from 8192 and to 32768
// in units respectively.
func (env *Env) GetMergeThreshold16Dot16Percent() (uint64, Error) {
	return env.GetOption(OptMergeThreshold16Dot16Percent)
}

// SetMergeThreshold16Dot16Percent Controls the in-process threshold of semi-empty pages merge.
// warning This is experimental option and subject for change or removal.
// details This option controls the in-process threshold of minimum page
// fill, as used space of percentage of a page. Neighbour pages emptier than
// this value are candidates for merging. The threshold value is specified
// in 1/65536 of percent, which is equivalent to the 16-dot-16 fixed point
// format. The specified value must be in the range from 12.5% (almost empty)
// to 50% (half empty) which corresponds to the range from 8192 and to 32768
// in units respectively.
func (env *Env) SetMergeThreshold16Dot16Percent(percent uint64) Error {
	return env.SetOption(OptMergeThreshold16Dot16Percent, percent)
}
