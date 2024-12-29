package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sunvim/gmdbx"
)

func main() {
	fmt.Println("test mdbx go")
	testRead()
	// testWrite()
}
func testRead() {
	env, err := gmdbx.NewEnv()
	if !errors.Is(err, gmdbx.ErrSuccess) {
		log.Fatal("open env: ", err)
	}
	if err = env.SetMaxDBS(1); err != gmdbx.ErrSuccess {
		log.Fatal("set max dbs: ", err)
	}

	env.SetGeometry(defaultGeometry)

	path := "tmp/test.db"

	baseDir := filepath.Dir(path)
	if baseDir != "" {
		os.MkdirAll(baseDir, 0755)
	}

	err = env.Open(path, gmdbx.EnvNoMetaSync|gmdbx.EnvSyncDurable|gmdbx.EnvNoSubDir, 0600)
	if !errors.Is(err, gmdbx.ErrSuccess) {
		log.Fatal("open db failed: ", err)
	}
	defer env.Close(false)

	tx := &gmdbx.Tx{}
	if err = env.Begin(tx, gmdbx.TxReadWrite); !errors.Is(err, gmdbx.ErrSuccess) {
		log.Fatal("open tx failed: ", err)
	}
	defer tx.Commit()
	dbi, _ := tx.OpenDBI("default", gmdbx.DBCreate)
	defer env.CloseDBI(dbi)
	stx := time.Now()
	var i uint64
	var cnt uint64
	for i = 0; i < 100000; i++ {
		kb := I2b(i)
		k := gmdbx.Bytes(&kb)
		vb := gmdbx.Val{}
		gerr := tx.Get(dbi, &k, &vb)
		if gerr != gmdbx.ErrSuccess {
			fmt.Printf("key: %v get failed: %v\n", k.Bytes(), gerr)
			return
		}
		if vb.Len != 0 {
			fmt.Printf("key: %x, value: %s\n", k.Bytes(), vb.String())
			cnt++
		}
	}

	fmt.Printf("elapse: %v ,count: %d \n", time.Since(stx), cnt)

}

func testWrite() {
	env, err := gmdbx.NewEnv()
	if !errors.Is(err, gmdbx.ErrSuccess) {
		log.Fatal("open env: ", err)
	}

	if err = env.SetMaxDBS(1); err != gmdbx.ErrSuccess {
		log.Fatal("set max dbs: ", err)
	}

	err = env.SetGeometry(defaultGeometry)
	if err != gmdbx.ErrSuccess {
		log.Fatal("set geometry failed")
	}
	err = env.SetOption(gmdbx.OptTxnDpLimit, 65535)
	if err != gmdbx.ErrSuccess {
		log.Fatal("set tx dp limit failed")
	}

	path := "tmp/test.db"

	baseDir := filepath.Dir(path)
	if baseDir != "" {
		os.MkdirAll(baseDir, 0755)
	}
	err = env.Open(path, gmdbx.EnvNoMetaSync|gmdbx.EnvSyncDurable|gmdbx.EnvNoSubDir, 0600)
	if !errors.Is(err, gmdbx.ErrSuccess) {
		log.Fatal("open db failed: ", err)
	}
	defer env.Close(false)

	tx := &gmdbx.Tx{}
	if err = env.Begin(tx, gmdbx.TxReadWrite); err != gmdbx.ErrSuccess {
		log.Fatal("open tx failed: ", err)
	}
	defer tx.Commit()

	dbi, err := tx.OpenDBI("default", gmdbx.DBCreate)
	if err != gmdbx.ErrSuccess {
		log.Fatal("open dbi failed: ", err)
	}
	defer env.CloseDBI(dbi)

	const prikey = "user/"
	stx := time.Now()
	var i uint64

	// if there are amounts of data to write, please lock the OS thread frist
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for i = 0; i < 100000; i++ {
		kb := I2b(i)
		vb := randomString(64)
		k := gmdbx.Bytes(&kb)
		v := gmdbx.Bytes(&vb)
		if err = tx.Put(dbi, &k, &v, gmdbx.PutUpsert); err != gmdbx.ErrSuccess {
			println("put failed: ", err)
			return
		}

	}

	fmt.Printf("elapse: %v \n", time.Since(stx))

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

const (
	defaultFlags = gmdbx.EnvSyncDurable |
		gmdbx.EnvNoTLS |
		gmdbx.EnvWriteMap |
		gmdbx.EnvLIFOReclaim |
		gmdbx.EnvNoMemInit |
		gmdbx.EnvCoalesce
)

var (
	defaultGeometry = gmdbx.Geometry{
		SizeLower:       1 << 30,
		SizeNow:         1 << 30,
		SizeUpper:       1 << 30,
		GrowthStep:      1 << 30,
		ShrinkThreshold: 1 << 63,
		PageSize:        1 << 16,
	}
)
