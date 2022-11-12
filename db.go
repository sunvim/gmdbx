package gmdbx

import (
	"errors"
)

type DB struct {
	env  *Env
	opts *Option
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

func (d *DB) CloseDBI(dbi DBI) error {
	if err := d.env.CloseDBI(dbi); err != ErrSuccess {
		return errors.New(err.Error())
	}
	return nil
}

func (d *DB) Close() error {
	if err := d.env.Close(false); err != ErrSuccess {
		return errors.New(err.Error())
	}
	return nil
}
