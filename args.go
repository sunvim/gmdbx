package gmdbx

/*
#include "mdbxgo.h"
*/
import "C"

const (
	MaxDBI      = uint32(C.MDBX_MAX_DBI)
	MaxDataSize = uint32(C.MDBX_MAXDATASIZE)
	MinPageSize = int(C.MDBX_MIN_PAGESIZE)
	MaxPageSize = int(C.MDBX_MAX_PAGESIZE)
)

type LogLevel int32

const (
	// LogFatal Critical conditions, i.e. assertion failures
	LogFatal = LogLevel(C.MDBX_LOG_FATAL)

	// LogError Enables logging for error conditions and ref MDBX_LOG_FATAL
	LogError = LogLevel(C.MDBX_LOG_ERROR)

	// LogWarn Enables logging for warning conditions and ref MDBX_LOG_ERROR ...
	// ref MDBX_LOG_FATAL
	LogWarn = LogLevel(C.MDBX_LOG_WARN)

	// LogNotice Enables logging for normal but significant condition and
	// ref MDBX_LOG_WARN ... ref MDBX_LOG_FATAL
	LogNotice = LogLevel(C.MDBX_LOG_NOTICE)

	// LogVerbose Enables logging for verbose informational and ref MDBX_LOG_NOTICE ...
	// ref MDBX_LOG_FATAL
	LogVerbose = LogLevel(C.MDBX_LOG_VERBOSE)

	// LogDebug Enables logging for debug-level messages and ref MDBX_LOG_VERBOSE ...
	// ref MDBX_LOG_FATAL
	LogDebug = LogLevel(C.MDBX_LOG_DEBUG)

	// LogTrace Enables logging for trace debug-level messages and ref MDBX_LOG_DEBUG ...
	// ref MDBX_LOG_FATAL
	LogTrace = LogLevel(C.MDBX_LOG_TRACE)

	// LogExtra Enables extra debug-level messages (dump pgno lists) and all other log-messages
	LogExtra = LogLevel(C.MDBX_LOG_EXTRA)
	LogMax   = LogLevel(7)

	// LogDontChange for ref mdbx_setup_debug() only: Don't change current settings
	LogDontChange = LogLevel(C.MDBX_LOG_DONTCHANGE)
)
