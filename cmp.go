package gmdbx

//#include "mdbxgo.h"
import "C"

type Cmp C.MDBX_cmp_func

var (
	CmpU16                    = (*Cmp)(C.mdbx_cmp_u16)
	CmpU32                    = (*Cmp)(C.mdbx_cmp_u32)
	CmpU64                    = (*Cmp)(C.mdbx_cmp_u64)
	CmpU16PrefixLexical       = (*Cmp)(C.mdbx_cmp_u16_prefix_lexical)
	CmpU16PrefixU64           = (*Cmp)(C.mdbx_cmp_u16_prefix_u64)
	CmpU32PrefixLexical       = (*Cmp)(C.mdbx_cmp_u32_prefix_lexical)
	CmpU32PrefixU64           = (*Cmp)(C.mdbx_cmp_u32_prefix_u64)
	CmpU64PrefixLexical       = (*Cmp)(C.mdbx_cmp_u64_prefix_lexical)
	CmpU64PrefixU64           = (*Cmp)(C.mdbx_cmp_u64_prefix_u64)
	CmpU32PrefixU64DupLexical = (*Cmp)(C.mdbx_cmp_u32_prefix_u64_dup_lexical)
	CmpU32PrefixU64DupU64     = (*Cmp)(C.mdbx_cmp_u32_prefix_u64_dup_u64)
	CmpU64PrefixU64DupLexical = (*Cmp)(C.mdbx_cmp_u64_prefix_u64_dup_lexical)
	CmpU64PrefixU64DupU64     = (*Cmp)(C.mdbx_cmp_u64_prefix_u64_dup_u64)
)
