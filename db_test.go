package gmdbx

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestDb() (*DB, error) {
	path := "testmdbx"
	os.RemoveAll(path)

	db, err := New(path)
	if err != nil {
		return nil, err
	}
	if err = db.Open(); err != nil {
		return nil, err
	}
	return db, nil

}

func TestUpdateAndView(t *testing.T) {
	db, err := newTestDb()
	if err != nil {
		t.Fatal("open db failed: ", err)
	}
	defer db.Close()

	key := []byte("hello")
	val := []byte("world")

	var dbi DBI

	err = db.Update(func(tx *Tx) error {
		dbi, err = tx.OpenDBI("test", DBCreate)
		if err != ErrSuccess {
			return errors.New(err.Error())
		}
		ki, vi := Bytes(&key), Bytes(&val)
		err = tx.Put(dbi, &ki, &vi, PutUpsert)

		if err != ErrSuccess {
			return errors.New(err.Error())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	vi := Val{}

	err = db.View(func(tx *Tx) error {
		ki := Bytes(&key)
		err = tx.Get(dbi, &ki, &vi)
		if err != ErrSuccess {
			return errors.New(err.Error())
		}

		return nil
	})
	if err != nil {
		t.Fatal("get failed: ", err)
	}

	assert.Equal(t, val, vi.Bytes(), "update and view")

}

func TestPut(b *testing.T) {
	env, err := NewEnv()
	if !errors.Is(err, ErrSuccess) {
		b.Fatal("open env: ", err)
	}

	if err = env.SetMaxDBS(1); err != ErrSuccess {
		b.Fatal("set max dbs: ", err)
	}

	err = env.SetGeometry(DefaultGeometry)
	if err != ErrSuccess {
		b.Fatal("set geometry failed")
	}
	err = env.SetOption(OptTxnDpLimit, 65535)
	if err != ErrSuccess {
		b.Fatal("set tx dp limit failed")
	}

	err = env.Open("tmp.db", DefaultFlags, 0755)
	if err != ErrSuccess {
		b.Fatal("open db failed: ", err)
	}

	tx := &Tx{}
	if err = env.Begin(tx, TxReadWrite); err != ErrSuccess {
		b.Fatal("open tx failed: ", err)
	}
	dbi, err := tx.OpenDBI("default", DBCreate)
	if err != ErrSuccess {
		b.Fatal("open dbi failed: ", err)
	}
	defer func() {
		env.CloseDBI(dbi)
		env.Close(false)
	}()
	const prikey = "user/"

	for i := 0; i < 100; i++ {
		k := ToVal(i)
		err = tx.Put(dbi, &k, &k, PutUpsert)
		if err != ErrSuccess {
			b.Fatalf("err: %v \n", err)
		}
	}
	tx.Commit()

	// cursor get
	rtx := &Tx{}
	if err = env.Begin(rtx, TxReadOnly); err != ErrSuccess {
		b.Fatal("open tx failed: ", err)
	}

	c, _ := rtx.OpenCursor(dbi)
	k := Val{}
	v := Val{}
	for {
		err := c.Get(&k, &v, CursorNext)
		if err != ErrSuccess {
			break
		}
	}

	// get
	for i := 0; i < 100; i++ {
		k := ToVal(i)
		rtx.Get(dbi, &k, &v)
	}

}

func BenchmarkDBPut(b *testing.B) {
	db, err := newTestDb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	k, v := "hello", "world"
	ki, vi := String(&k), String(&v)
	b.ResetTimer()
	db.Update(func(tx *Tx) error {

		dbi, _ := tx.OpenDBI("default", DBCreate)
		defer db.CloseDBI(dbi)

		for i := 0; i < b.N; i++ {
			tx.Put(dbi, &ki, &vi, PutUpsert)
		}
		return nil
	})
}

func I2b(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

var defaultLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString returns a random string with a fixed length
func randomString(n int) []byte {

	letters := defaultLetters

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return b
}
