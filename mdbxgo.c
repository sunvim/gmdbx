#include "mdbxgo.h"

int cmp_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (a->iov_len == b->iov_len)
    return a->iov_len ? memcmp(a->iov_base, b->iov_base, a->iov_len) : 0;

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base, b->iov_base, shortest) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u16(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 2 || b->iov_len < 2)) {
    return cmp_lexical(a, b);
  }
  uint16_t aa = *((uint16_t*)a->iov_base);
  uint16_t bb = *((uint16_t*)b->iov_base);
  return bb > aa ? -1 : aa > bb;
}

int mdbx_cmp_u32(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 4 || b->iov_len < 4)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)b->iov_base);
  return bb > aa ? -1 : aa > bb;
}

int mdbx_cmp_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 8 || b->iov_len < 8)) {
    return cmp_lexical(a, b);
  }
  uint64_t aa = *((uint64_t*)a->iov_base);
  uint64_t bb = *((uint64_t*)b->iov_base);
  return bb > aa ? -1 : aa > bb;
}

int mdbx_cmp_u16_prefix_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 2 || b->iov_len < 2)) {
    return cmp_lexical(a, b);
  }
  uint16_t aa = *((uint16_t*)a->iov_base);
  uint16_t bb = *((uint16_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  if (a->iov_len == b->iov_len)
    return a->iov_len ? memcmp(a->iov_base+2, b->iov_base+2, a->iov_len-2) : 0;

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base+2, b->iov_base+2, shortest-2) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u16_prefix_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 10 || b->iov_len < 10)) {
    return cmp_lexical(a, b);
  }
  uint16_t aa = *((uint16_t*)a->iov_base);
  uint16_t bb = *((uint16_t*)a->iov_base);
  if (aa < bb) return -1;
  if (aa > bb) return 1;
  uint64_t aa2 = *((uint64_t*)a->iov_base+2);
  uint64_t bb2 = *((uint64_t*)b->iov_base+2);
  if (aa2 < bb2) return -1;
  if (aa2 > bb2) return 1;
  return 0;
}

int mdbx_cmp_u32_prefix_u64_dup_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 12 || b->iov_len < 12)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  if (a->iov_len == b->iov_len) {
    int result = a->iov_len ? memcmp(a->iov_base+12, b->iov_base+12, a->iov_len-12) : 0;
	if (result != 0) return result;

	uint64_t aaa = *((uint64_t*)a->iov_base+4);
	uint64_t bbb = *((uint64_t*)a->iov_base+4);
	if (aaa < bbb) {
	  return -1;
	}
	if (aaa > bbb) {
	  return 1;
	}
	return 0;
  }

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base+12, b->iov_base+12, shortest-12) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u64_prefix_u64_dup_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 16 || b->iov_len < 16)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  if (a->iov_len == b->iov_len) {
    int result = a->iov_len ? memcmp(a->iov_base+16, b->iov_base+16, a->iov_len-16) : 0;
	if (result != 0) return result;

	uint64_t aaa = *((uint64_t*)a->iov_base+8);
	uint64_t bbb = *((uint64_t*)a->iov_base+8);
	if (aaa < bbb) {
	  return -1;
	}
	if (aaa > bbb) {
	  return 1;
	}
	return 0;
  }

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base+16, b->iov_base+16, shortest-16) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u32_prefix_u64_dup_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 20 || b->iov_len < 20)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  uint64_t av = *((uint64_t*)a->iov_base+12);
  uint64_t bv = *((uint64_t*)a->iov_base+12);
  if (av < bv) {
	return -1;
  }
  if (av > bv) {
    return 1;
  }
  av = *((uint64_t*)a->iov_base+4);
  bv = *((uint64_t*)a->iov_base+4);
  if (av < bv) {
    return -1;
  }
  if (av > bv) {
    return 1;
  }
  return 0;
}

int mdbx_cmp_u64_prefix_u64_dup_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 24 || b->iov_len < 24)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  uint64_t av = *((uint64_t*)a->iov_base+16);
  uint64_t bv = *((uint64_t*)a->iov_base+16);
  if (av < bv) {
	return -1;
  }
  if (av > bv) {
    return 1;
  }
  av = *((uint64_t*)a->iov_base+8);
  bv = *((uint64_t*)a->iov_base+8);
  if (av < bv) {
    return -1;
  }
  if (av > bv) {
    return 1;
  }
  return 0;
}

int mdbx_cmp_u32_prefix_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 4 || b->iov_len < 4)) {
    return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  if (a->iov_len == b->iov_len)
    return a->iov_len ? memcmp(a->iov_base+4, b->iov_base+4, a->iov_len-4) : 0;

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base+4, b->iov_base+4, shortest-4) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u32_prefix_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 12 || b->iov_len < 12)) {
   return cmp_lexical(a, b);
  }
  uint32_t aa = *((uint32_t*)a->iov_base);
  uint32_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) return -1;
  if (aa > bb) return 1;
  uint64_t aa2 = *((uint64_t*)a->iov_base+4);
  uint64_t bb2 = *((uint64_t*)b->iov_base+4);
  if (aa2 < bb2) return -1;
  if (aa2 > bb2) return 1;
  return 0;
}

int mdbx_cmp_u64_prefix_lexical(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 8 || b->iov_len < 8)) {
    return cmp_lexical(a, b);
  }
  uint64_t aa = *((uint64_t*)a->iov_base);
  uint64_t bb = *((uint64_t*)a->iov_base);
  if (aa < bb) {
	return -1;
  }
  if (aa > bb) {
    return 1;
  }
  if (a->iov_len == b->iov_len)
    return a->iov_len ? memcmp(a->iov_base+8, b->iov_base+8, a->iov_len-8) : 0;

  const int diff_len = (a->iov_len < b->iov_len) ? -1 : 1;
  const size_t shortest = (a->iov_len < b->iov_len) ? a->iov_len : b->iov_len;
  int diff_data = shortest ? memcmp(a->iov_base+8, b->iov_base+8, shortest-8) : 0;
  return likely(diff_data) ? diff_data : diff_len;
}

int mdbx_cmp_u64_prefix_u64(const MDBX_val *a, const MDBX_val *b) {
  if (unlikely(a->iov_len < 16 || b->iov_len < 16)) {
   return cmp_lexical(a, b);
  }
  uint64_t aa = *((uint32_t*)a->iov_base);
  uint64_t bb = *((uint32_t*)a->iov_base);
  if (aa < bb) return -1;
  if (aa > bb) return 1;
  aa = *((uint64_t*)a->iov_base+8);
  bb = *((uint64_t*)b->iov_base+8);
  if (aa < bb) return -1;
  if (aa > bb) return 1;
  return 0;
}


void do_mdbx_strerror(size_t arg0, size_t arg1) {
	mdbx_strerror_t* args = (mdbx_strerror_t*)(void*)arg0;
	args->result = (size_t)(void*)mdbx_strerror((int)args->code);
}


void do_mdbx_env_set_geometry(size_t arg0, size_t arg1) {
	mdbx_env_set_geometry_t* args = (mdbx_env_set_geometry_t*)(void*)arg0;
	args->result = (int32_t)mdbx_env_set_geometry(
		(MDBX_env*)(void*)args->env,
		args->size_lower,
		args->size_now,
		args->size_upper,
		args->growth_step,
		args->shrink_threshold,
		args->page_size
	);
}


void do_mdbx_env_info_ex(size_t arg0, size_t arg1) {
	mdbx_env_info_t* args = (mdbx_env_info_t*)(void*)arg0;
	args->result = (int32_t)mdbx_env_info_ex(
		(MDBX_env*)(void*)args->env,
		(MDBX_txn*)(void*)args->txn,
		(MDBX_envinfo*)(void*)args->info,
		args->size
	);
}

void do_mdbx_txn_info(size_t arg0, size_t arg1) {
	mdbx_txn_info_t* args = (mdbx_txn_info_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_info(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_txn_info*)(void*)args->info,
		args->scan_rlt >= 0
	);
}


void do_mdbx_txn_flags(size_t arg0, size_t arg1) {
	mdbx_txn_flags_t* args = (mdbx_txn_flags_t*)(void*)arg0;
	args->flags = (int32_t)mdbx_txn_flags(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_txn_id(size_t arg0, size_t arg1) {
	mdbx_txn_id_t* args = (mdbx_txn_id_t*)(void*)arg0;
	args->id = mdbx_txn_id(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_txn_commit_ex(size_t arg0, size_t arg1) {
	mdbx_txn_commit_ex_t* args = (mdbx_txn_commit_ex_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_commit_ex(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_commit_latency*)(void*)args->latency
	);
}

void do_mdbx_txn_abort(size_t arg0, size_t arg1) {
	mdbx_txn_result_t* args = (mdbx_txn_result_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_abort(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_txn_break(size_t arg0, size_t arg1) {
	mdbx_txn_result_t* args = (mdbx_txn_result_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_break(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_txn_reset(size_t arg0, size_t arg1) {
	mdbx_txn_result_t* args = (mdbx_txn_result_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_reset(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_txn_renew(size_t arg0, size_t arg1) {
	mdbx_txn_result_t* args = (mdbx_txn_result_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_renew(
		(MDBX_txn*)(void*)args->txn
	);
}

void do_mdbx_canary_put(size_t arg0, size_t arg1) {
	mdbx_txn_canary_t* args = (mdbx_txn_canary_t*)(void*)arg0;
	args->result = (int32_t)mdbx_canary_put(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_canary*)(void*)args->canary
	);
}

void do_mdbx_canary_get(size_t arg0, size_t arg1) {
	mdbx_txn_canary_t* args = (mdbx_txn_canary_t*)(void*)arg0;
	args->result = (int32_t)mdbx_canary_get(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_canary*)(void*)args->canary
	);
}

void do_mdbx_dbi_stat(size_t arg0, size_t arg1) {
	mdbx_dbi_stat_t* args = (mdbx_dbi_stat_t*)(void*)arg0;
	args->result = (int32_t)mdbx_dbi_stat(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_stat*)(void*)args->stat,
		args->size
	);
}

void do_mdbx_dbi_flags_ex(size_t arg0, size_t arg1) {
	mdbx_dbi_flags_t* args = (mdbx_dbi_flags_t*)(void*)arg0;
	args->result = (int32_t)mdbx_dbi_flags_ex(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(unsigned*)(void*)args->flags,
		(unsigned*)(void*)args->state
	);
}

void do_mdbx_drop(size_t arg0, size_t arg1) {
	mdbx_drop_t* args = (mdbx_drop_t*)(void*)arg0;
	args->result = (int32_t)mdbx_drop(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		args->del > 0
	);
}

void do_mdbx_get(size_t arg0, size_t arg1) {
	mdbx_get_t* args = (mdbx_get_t*)(void*)arg0;
	args->result = (int32_t)mdbx_get(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data
	);
}

void do_mdbx_get_equal_or_great(size_t arg0, size_t arg1) {
	mdbx_get_t* args = (mdbx_get_t*)(void*)arg0;
	args->result = (int32_t)mdbx_get_equal_or_great(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data
	);
}

void do_mdbx_get_ex(size_t arg0, size_t arg1) {
	mdbx_get_ex_t* args = (mdbx_get_ex_t*)(void*)arg0;
	args->result = (int32_t)mdbx_get_ex(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data,
		(size_t*)(void*)args->values_count
	);
}

void do_mdbx_put(size_t arg0, size_t arg1) {
	mdbx_put_t* args = (mdbx_put_t*)(void*)arg0;
	args->result = (int32_t)mdbx_put(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data,
		(MDBX_put_flags_t)args->flags
	);
}

void do_mdbx_replace(size_t arg0, size_t arg1) {
	mdbx_replace_t* args = (mdbx_replace_t*)(void*)arg0;
	args->result = (int32_t)mdbx_replace(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data,
		(MDBX_val*)(void*)args->old_data,
		(MDBX_put_flags_t)args->flags
	);
}

void do_mdbx_del(size_t arg0, size_t arg1) {
	mdbx_del_t* args = (mdbx_del_t*)(void*)arg0;
	args->result = (int32_t)mdbx_del(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data
	);
}


void do_mdbx_txn_begin_ex(size_t arg0, size_t arg1) {
	mdbx_txn_begin_t* args = (mdbx_txn_begin_t*)(void*)arg0;
	args->result = (int32_t)mdbx_txn_begin_ex(
		(MDBX_env*)(void*)args->env,
		//(MDBX_txn*)(void*)args->parent,
		NULL,
		(MDBX_txn_flags_t)args->flags,
		(MDBX_txn**)(void*)args->txn,
		(void*)args->context
	);
}


void do_mdbx_cursor_create(size_t arg0, size_t arg1) {
	mdbx_cursor_create_t* args = (mdbx_cursor_create_t*)(void*)arg0;
	args->cursor = (size_t)mdbx_cursor_create(
		(void*)args->context
	);
}


void do_mdbx_cursor_bind(size_t arg0, size_t arg1) {
	mdbx_cursor_bind_t* args = (mdbx_cursor_bind_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_bind(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_cursor*)(void*)args->cursor,
		(MDBX_dbi)args->dbi
	);
}


void do_mdbx_cursor_open(size_t arg0, size_t arg1) {
	mdbx_cursor_open_t* args = (mdbx_cursor_open_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_open(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_dbi)args->dbi,
		(MDBX_cursor**)(void*)args->cursor
	);
}

void do_mdbx_cursor_close(size_t arg0, size_t arg1) {
	mdbx_cursor_close((MDBX_cursor*)(void*)arg0);
}


void do_mdbx_cursor_renew(size_t arg0, size_t arg1) {
	mdbx_cursor_renew_t* args = (mdbx_cursor_renew_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_renew(
		(MDBX_txn*)(void*)args->txn,
		(MDBX_cursor*)(void*)args->cursor
	);
}


void do_mdbx_cursor_txn(size_t arg0, size_t arg1) {
	mdbx_cursor_txn_t* args = (mdbx_cursor_txn_t*)(void*)arg0;
	args->txn = (size_t)mdbx_cursor_txn(
		(MDBX_cursor*)(void*)args->cursor
	);
}

void do_mdbx_cursor_dbi(size_t arg0, size_t arg1) {
	mdbx_cursor_dbi_t* args = (mdbx_cursor_dbi_t*)(void*)arg0;
	args->dbi = (uint32_t)mdbx_cursor_dbi(
		(MDBX_cursor*)(void*)args->cursor
	);
}

void do_mdbx_cursor_copy(size_t arg0, size_t arg1) {
	mdbx_cursor_copy_t* args = (mdbx_cursor_copy_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_copy(
		(MDBX_cursor*)(void*)args->src,
		(MDBX_cursor*)(void*)args->dest
	);
}

void do_mdbx_cursor_get(size_t arg0, size_t arg1) {
	mdbx_cursor_get_t* args = (mdbx_cursor_get_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_get(
		(MDBX_cursor*)(void*)args->cursor,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data,
		(MDBX_cursor_op)args->op
	);
}

void do_mdbx_cursor_put(size_t arg0, size_t arg1) {
	mdbx_cursor_put_t* args = (mdbx_cursor_put_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_put(
		(MDBX_cursor*)(void*)args->cursor,
		(MDBX_val*)(void*)args->key,
		(MDBX_val*)(void*)args->data,
		(MDBX_put_flags_t)args->flags
	);
}

void do_mdbx_cursor_del(size_t arg0, size_t arg1) {
	mdbx_cursor_del_t* args = (mdbx_cursor_del_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_del(
		(MDBX_cursor*)(void*)args->cursor,
		(MDBX_put_flags_t)args->flags
	);
}

void do_mdbx_cursor_count(size_t arg0, size_t arg1) {
	mdbx_cursor_count_t* args = (mdbx_cursor_count_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_count(
		(MDBX_cursor*)(void*)args->cursor,
		(size_t*)(void*)args->count
	);
}

void do_mdbx_cursor_eof(size_t arg0, size_t arg1) {
	mdbx_cursor_eof_t* args = (mdbx_cursor_eof_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_eof(
		(MDBX_cursor*)(void*)args->cursor
	);
}

void do_mdbx_cursor_on_first(size_t arg0, size_t arg1) {
	mdbx_cursor_on_first_t* args = (mdbx_cursor_on_first_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_on_first(
		(MDBX_cursor*)(void*)args->cursor
	);
}

void do_mdbx_cursor_on_last(size_t arg0, size_t arg1) {
	mdbx_cursor_on_last_t* args = (mdbx_cursor_on_last_t*)(void*)arg0;
	args->result = (int32_t)mdbx_cursor_on_last(
		(MDBX_cursor*)(void*)args->cursor
	);
}

void do_mdbx_estimate_distance(size_t arg0, size_t arg1) {
	mdbx_estimate_distance_t* args = (mdbx_estimate_distance_t*)(void*)arg0;
	args->result = (int32_t)mdbx_estimate_distance(
		(MDBX_cursor*)(void*)args->first,
		(MDBX_cursor*)(void*)args->last,
		(ptrdiff_t*)(void*)args->distance_items
	);
}
