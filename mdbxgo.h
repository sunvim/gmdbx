#pragma once
#ifndef H_MDBX_GO
#define H_MDBX_GO

#include <stdlib.h>
#include <string.h>
#include <inttypes.h>
#include "mdbx.h"

#ifndef likely
#   if (defined(__GNUC__) || __has_builtin(__builtin_expect)) && !defined(__COVERITY__)
#       define likely(cond) __builtin_expect(!!(cond), 1)
#   else
#       define likely(x) (!!(x))
#   endif
#endif

#ifndef unlikely
#   if (defined(__GNUC__) || __has_builtin(__builtin_expect)) && !defined(__COVERITY__)
#       define unlikely(cond) __builtin_expect(!!(cond), 0)
#   else
#       define unlikely(x) (!!(x))
#   endif
#endif


int cmp_lexical(const MDBX_val *a, const MDBX_val *b);

int mdbx_cmp_u16(const MDBX_val *a, const MDBX_val *b);
int mdbx_cmp_u32(const MDBX_val *a, const MDBX_val *b);

int mdbx_cmp_u64(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u16_prefix_lexical(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u16_prefix_u64(const MDBX_val *a, const MDBX_val *b);

int mdbx_cmp_u32_prefix_u64_dup_lexical(const MDBX_val *a, const MDBX_val *b) ;
int mdbx_cmp_u64_prefix_u64_dup_lexical(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u32_prefix_u64_dup_u64(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u64_prefix_u64_dup_u64(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u32_prefix_lexical(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u32_prefix_u64(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u64_prefix_lexical(const MDBX_val *a, const MDBX_val *b) ;

int mdbx_cmp_u64_prefix_u64(const MDBX_val *a, const MDBX_val *b) ;

typedef struct mdbx_strerror_t {
	size_t result;
	int32_t code;
} mdbx_strerror_t;

void do_mdbx_strerror(size_t arg0, size_t arg1) ;

typedef struct mdbx_env_set_geometry_t {
	size_t env;
	size_t size_lower;
	size_t size_now;
	size_t size_upper;
	size_t growth_step;
	size_t shrink_threshold;
	size_t page_size;
	int32_t result;
} mdbx_env_set_geometry_t;

void do_mdbx_env_set_geometry(size_t arg0, size_t arg1) ;

typedef struct mdbx_env_info_t {
	size_t env;
	size_t txn;
	size_t info;
	size_t size;
	int32_t result;
} mdbx_env_info_t;

void do_mdbx_env_info_ex(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_info_t {
	size_t txn;
	size_t info;
	int32_t scan_rlt;
	int32_t result;
} mdbx_txn_info_t;

void do_mdbx_txn_info(size_t arg0, size_t arg1);
typedef struct mdbx_txn_flags_t {
	size_t txn;
	int32_t flags;
} mdbx_txn_flags_t;

void do_mdbx_txn_flags(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_id_t {
	size_t txn;
	uint64_t id;
} mdbx_txn_id_t;

void do_mdbx_txn_id(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_commit_ex_t {
	size_t txn;
	size_t latency;
	int32_t result;
} mdbx_txn_commit_ex_t;

void do_mdbx_txn_commit_ex(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_result_t {
	size_t txn;
	int32_t result;
} mdbx_txn_result_t;

void do_mdbx_txn_abort(size_t arg0, size_t arg1) ;

void do_mdbx_txn_break(size_t arg0, size_t arg1) ;

void do_mdbx_txn_reset(size_t arg0, size_t arg1) ;

void do_mdbx_txn_renew(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_canary_t {
	size_t txn;
	size_t canary;
	int32_t result;
} mdbx_txn_canary_t;

void do_mdbx_canary_put(size_t arg0, size_t arg1) ;

void do_mdbx_canary_get(size_t arg0, size_t arg1) ;

typedef struct mdbx_dbi_stat_t {
	size_t txn;
	size_t stat;
	size_t size;
	uint32_t dbi;
	int32_t result;
} mdbx_dbi_stat_t;

void do_mdbx_dbi_stat(size_t arg0, size_t arg1) ;

typedef struct mdbx_dbi_flags_t {
	size_t txn;
	size_t flags;
	size_t state;
	uint32_t dbi;
	int32_t result;
} mdbx_dbi_flags_t;

void do_mdbx_dbi_flags_ex(size_t arg0, size_t arg1) ;

typedef struct mdbx_drop_t {
	size_t txn;
	size_t del;
	uint32_t dbi;
	int32_t result;
} mdbx_drop_t;

void do_mdbx_drop(size_t arg0, size_t arg1) ;

typedef struct mdbx_get_t {
	size_t txn;
	size_t key;
	size_t data;
	uint32_t dbi;
	int32_t result;
} mdbx_get_t;

void do_mdbx_get(size_t arg0, size_t arg1) ;

void do_mdbx_get_equal_or_great(size_t arg0, size_t arg1) ;

typedef struct mdbx_get_ex_t {
	size_t txn;
	size_t key;
	size_t data;
	size_t values_count;
	uint32_t dbi;
	int32_t result;
} mdbx_get_ex_t;

void do_mdbx_get_ex(size_t arg0, size_t arg1) ;

typedef struct mdbx_put_t {
	size_t txn;
	size_t key;
	size_t data;
	uint32_t dbi;
	uint32_t flags;
	int32_t result;
} mdbx_put_t;

void do_mdbx_put(size_t arg0, size_t arg1) ;

typedef struct mdbx_replace_t {
	size_t txn;
	size_t key;
	size_t data;
	size_t old_data;
	uint32_t dbi;
	uint32_t flags;
	int32_t result;
} mdbx_replace_t;

void do_mdbx_replace(size_t arg0, size_t arg1) ;

typedef struct mdbx_del_t {
	size_t txn;
	size_t key;
	size_t data;
	uint32_t dbi;
	int32_t result;
} mdbx_del_t;

void do_mdbx_del(size_t arg0, size_t arg1) ;

typedef struct mdbx_txn_begin_t {
	size_t env;
	size_t parent;
	size_t txn;
	size_t context;
	uint32_t flags;
	int32_t result;
} mdbx_txn_begin_t;

void do_mdbx_txn_begin_ex(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_create_t {
	size_t context;
	size_t cursor;
} mdbx_cursor_create_t;

void do_mdbx_cursor_create(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_bind_t {
	size_t txn;
	size_t cursor;
	uint32_t dbi;
	int32_t result;
} mdbx_cursor_bind_t;

void do_mdbx_cursor_bind(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_open_t {
	size_t txn;
	size_t cursor;
	uint32_t dbi;
	int32_t result;
} mdbx_cursor_open_t;

void do_mdbx_cursor_open(size_t arg0, size_t arg1) ;
void do_mdbx_cursor_close(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_renew_t {
	size_t txn;
	size_t cursor;
	int32_t result;
} mdbx_cursor_renew_t;

void do_mdbx_cursor_renew(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_txn_t {
	size_t cursor;
	size_t txn;
} mdbx_cursor_txn_t;

void do_mdbx_cursor_txn(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_dbi_t {
	size_t cursor;
	uint32_t dbi;
} mdbx_cursor_dbi_t;

void do_mdbx_cursor_dbi(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_copy_t {
	size_t src;
	size_t dest;
	int32_t result;
} mdbx_cursor_copy_t;

void do_mdbx_cursor_copy(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_get_t {
	size_t cursor;
	size_t key;
	size_t data;
	uint32_t op;
	int32_t result;
} mdbx_cursor_get_t;

void do_mdbx_cursor_get(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_put_t {
	size_t cursor;
	size_t key;
	size_t data;
	uint32_t flags;
	int32_t result;
} mdbx_cursor_put_t;

void do_mdbx_cursor_put(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_del_t {
	size_t cursor;
	uint32_t flags;
	int32_t result;
} mdbx_cursor_del_t;

void do_mdbx_cursor_del(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_count_t {
	size_t cursor;
	size_t count;
	int32_t result;
} mdbx_cursor_count_t;

void do_mdbx_cursor_count(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_eof_t {
	size_t cursor;
	int32_t result;
} mdbx_cursor_eof_t;

void do_mdbx_cursor_eof(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_on_first_t {
	size_t cursor;
	int32_t result;
} mdbx_cursor_on_first_t;

void do_mdbx_cursor_on_first(size_t arg0, size_t arg1) ;

typedef struct mdbx_cursor_on_last_t {
	size_t cursor;
	int32_t result;
} mdbx_cursor_on_last_t;

void do_mdbx_cursor_on_last(size_t arg0, size_t arg1) ;

typedef struct mdbx_estimate_distance_t {
	size_t first;
	size_t last;
	int64_t distance_items;
	int32_t result;
} mdbx_estimate_distance_t;

void do_mdbx_estimate_distance(size_t arg0, size_t arg1) ;

#endif