package gmdbx

import (
	"errors"
	"reflect"
	"runtime"
)

type DB struct {
	env  *Env
	opts *Option
	dbi  DBI
}

// New create new database
func New(path string) (*DB, error) {
	env, err := NewEnv()
	if err != ErrSuccess {
		return nil, errors.New(err.Error())
	}
	opts := &DefaultOption
	opts.Path = path
	return &DB{
		env:  env,
		opts: opts,
	}, nil
}

func (d *DB) SetEnvOption(opt Opt, value uint64) error {

	err := d.env.SetOption(opt, value)

	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	return nil
}

// SetOption  set database option
func (d *DB) SetOption(opts *Option) {
	d.opts = opts
}

func (d *DB) Open() error {
	err := d.env.SetGeometry(d.opts.Geometry)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	err = d.env.SetMaxDBS(d.opts.MaxDBS)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	err = d.env.SetOption(OptTxnDpLimit, uint64(d.opts.TxnDpLimit))
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	err = d.env.Open(d.opts.Path, d.opts.Flags, 0664)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}

	txn := NewTransaction(d.env)
	err = d.env.Begin(txn, TxReadWrite)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	defer txn.Commit()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	d.dbi, err = txn.OpenDBI("default", DBCreate)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}

	return nil
}

func (d *DB) Update(fn func(tx *Tx) error) error {
	txn := NewTransaction(d.env)
	err := d.env.Begin(txn, TxReadWrite)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	defer txn.Commit()

	return fn(txn)
}

func (d *DB) View(fn func(tx *Tx) error) error {
	txn := NewTransaction(d.env)
	err := d.env.Begin(txn, TxReadOnly)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	defer txn.Commit()

	return fn(txn)
}

func (d *DB) Put(k, v []byte) error {
	txn := NewTransaction(d.env)
	err := d.env.Begin(txn, TxReadWrite)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	defer txn.Commit()

	var ki, vi Val
	if k != nil && !reflect.DeepEqual(k, []byte{}) {
		ki = Bytes(&k)
	}
	if v != nil && !reflect.DeepEqual(v, []byte{}) {
		vi = Bytes(&v)
	}

	err = txn.Put(d.dbi, &ki, &vi, PutUpsert)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	return nil
}

func (d *DB) Del(k []byte) error {

	txn := NewTransaction(d.env)
	err := d.env.Begin(txn, TxReadWrite)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	defer txn.Commit()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var ki Val
	if k != nil && !reflect.DeepEqual(k, []byte{}) {
		ki = Bytes(&k)
	}
	err = txn.Delete(d.dbi, &ki, nil)
	if err != ErrSuccess {
		return errors.New(err.Error())
	}
	return nil
}

func (d *DB) Get(k []byte) ([]byte, error) {
	txn := NewTransaction(d.env)
	err := d.env.Begin(txn, TxReadOnly)
	if err != ErrSuccess {
		return nil, errors.New(err.Error())
	}
	defer txn.Commit()
	var ki Val
	if k != nil && !reflect.DeepEqual(k, []byte{}) {
		ki = Bytes(&k)
	}
	vi := Val{}
	err = txn.Get(d.dbi, &ki, &vi)
	if err != ErrSuccess && err != ErrNotFound {
		return nil, errors.New(err.Error())
	}
	if err == ErrNotFound {
		return nil, NotFound
	}
	return vi.Bytes(), nil
}

func (d *DB) Close() error {
	d.env.CloseDBI(d.dbi)
	d.env.Close(false)
	return nil
}
