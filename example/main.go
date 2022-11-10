package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
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

	env.SetGeometry(DefaultStableGeometry)

	err = env.Open("tmp.db", gmdbx.EnvNoMetaSync|gmdbx.EnvSyncDurable, 0755)
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
	const prikey = "user/"
	stx := time.Now()
	var i uint64
	var vb = gmdbx.Val{}
	var cnt uint64
	for i = 0; i < 100000; i++ {
		kb := append([]byte(prikey), I2b(i)...)
		k := gmdbx.Bytes(&kb)
		tx.Get(dbi, &k, &vb)
		if vb.Len != 0 {
			println("idx: ", i, " content: ", vb.String())
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

	err = env.SetGeometry(DefaultStableGeometry)
	if err != gmdbx.ErrSuccess {
		log.Fatal("set geometry failed")
	}
	err = env.SetOption(gmdbx.OptTxnDpLimit, 65535)
	if err != gmdbx.ErrSuccess {
		log.Fatal("set tx dp limit failed")
	}

	err = env.Open("tmp.db", gmdbx.EnvNoMetaSync|gmdbx.EnvSyncDurable, 0755)
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for i = 0; i < 100000; i++ {
		kb := append([]byte(prikey), I2b(i)...)
		vb := randomString(4096)
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
	DefaultLogFlags = gmdbx.EnvNoMetaSync |
		gmdbx.EnvNoTLS |
		gmdbx.EnvWriteMap |
		gmdbx.EnvLIFOReclaim |
		gmdbx.EnvNoMemInit |
		gmdbx.EnvCoalesce

	DefaultStableFlags = gmdbx.EnvSyncDurable |
		gmdbx.EnvNoTLS |
		gmdbx.EnvWriteMap |
		gmdbx.EnvLIFOReclaim |
		gmdbx.EnvNoMemInit |
		gmdbx.EnvCoalesce

	Kilobyte = 1024
	Megabyte = 1024 * 1024
	Gigabyte = Megabyte * 1024
	Terabyte = Gigabyte * 1024
)

var (
	DefaultStableGeometry = gmdbx.Geometry{
		SizeLower:       1 << 30,
		SizeNow:         1 << 30,
		SizeUpper:       1 << 34,
		GrowthStep:      1 << 30,
		ShrinkThreshold: 1 << 63,
		PageSize:        1 << 16,
	}
	DefaultLogGeometry = gmdbx.Geometry{
		SizeLower:       1 * Megabyte,
		SizeNow:         1 * Megabyte,
		SizeUpper:       4 * Gigabyte,
		GrowthStep:      16 * Megabyte,
		ShrinkThreshold: 8 * Megabyte,
		PageSize:        8 * Kilobyte,
	}
)
