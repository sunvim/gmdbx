package capture

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
	"syscall"
)

var lockStdFileDescriptorsSwapping sync.Mutex

// Capture captures stderr and stdout of a given function call.
func Capture(call func()) (output []byte, err error) {
	originalStdErr, originalStdOut := os.Stderr, os.Stdout
	defer func() {
		lockStdFileDescriptorsSwapping.Lock()

		os.Stderr, os.Stdout = originalStdErr, originalStdOut

		lockStdFileDescriptorsSwapping.Unlock()
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer func() {
		e := r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()

	os.Stderr, os.Stdout = w, w

	lockStdFileDescriptorsSwapping.Unlock()

	out := make(chan []byte)
	go func() {
		defer func() {
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()

		var b bytes.Buffer

		_, err := io.Copy(&b, r)
		if err != nil {
			panic(err)
		}

		out <- b.Bytes()
	}()

	call()

	err = w.Close()
	if err != nil {
		return nil, err
	}
	w = nil

	return <-out, err
}

// CaptureWithCGo captures stderr and stdout as well as stderr and stdout of C of a given function call.
func CaptureWithCGo(call func()) (output []byte, err error) {
	lockStdFileDescriptorsSwapping.Lock()

	originalStdout, e := syscall.Dup(syscall.Stdout)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	originalStderr, e := syscall.Dup(syscall.Stderr)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	defer func() {
		lockStdFileDescriptorsSwapping.Lock()

		if e := syscall.Dup2(originalStdout, syscall.Stdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Close(originalStdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Dup2(originalStderr, syscall.Stderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Close(originalStderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}

		lockStdFileDescriptorsSwapping.Unlock()
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer func() {
		e := r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()

	if e := syscall.Dup2(int(w.Fd()), syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}
	if e := syscall.Dup2(int(w.Fd()), syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	out := make(chan []byte)
	go func() {
		defer func() {
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()

		var b bytes.Buffer

		_, err := io.Copy(&b, r)
		if err != nil {
			panic(err)
		}

		out <- b.Bytes()
	}()

	call()

	lockStdFileDescriptorsSwapping.Lock()

	C.fflush(C.stdout)

	err = w.Close()
	if err != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, err
	}
	w = nil

	if e := syscall.Close(syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}
	if e := syscall.Close(syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return nil, e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	return <-out, err
}

type Msg struct {
	Line string
}

// CaptureWithCGo captures stderr and stdout as well as stderr and stdout of C of a given function call.
func CaptureWithCGoChan(ch chan string, call func()) (err error) {
	lockStdFileDescriptorsSwapping.Lock()

	originalStdout, e := syscall.Dup(syscall.Stdout)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}

	originalStderr, e := syscall.Dup(syscall.Stderr)
	if e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	defer func() {
		lockStdFileDescriptorsSwapping.Lock()

		if e := syscall.Dup2(originalStdout, syscall.Stdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Close(originalStdout); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Dup2(originalStderr, syscall.Stderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}
		if e := syscall.Close(originalStderr); e != nil {
			lockStdFileDescriptorsSwapping.Unlock()

			err = e
		}

		lockStdFileDescriptorsSwapping.Unlock()
	}()

	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	defer func() {
		e := r.Close()
		if e != nil {
			err = e
		}
		if w != nil {
			e = w.Close()
			if err != nil {
				err = e
			}
		}
	}()

	lockStdFileDescriptorsSwapping.Lock()

	if e := syscall.Dup2(int(w.Fd()), syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}
	if e := syscall.Dup2(int(w.Fd()), syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			// If there is a panic in the function call, copying from "r" does not work anymore.
			_ = recover()
		}()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			ch <- text
		}
	}()

	call()

	lockStdFileDescriptorsSwapping.Lock()

	C.fflush(C.stdout)

	err = w.Close()
	if err != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return err
	}
	w = nil

	if e := syscall.Close(syscall.Stdout); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}
	if e := syscall.Close(syscall.Stderr); e != nil {
		lockStdFileDescriptorsSwapping.Unlock()

		return e
	}

	lockStdFileDescriptorsSwapping.Unlock()

	wg.Wait()
	return err
}

type Buffer struct {
	rd *io.PipeReader
	wr *io.PipeWriter
}

func NewBuffer() *Buffer {
	rd, wr := io.Pipe()
	return &Buffer{rd, wr}
}

func (b *Buffer) Close() error {
	if b.rd == nil {
		return nil
	}
	_ = b.rd.Close()
	_ = b.wr.Close()
	b.rd = nil
	b.wr = nil
	return nil
}

func (b *Buffer) Read(d []byte) (int, error) {
	return b.rd.Read(d)
}

func (b *Buffer) Write(d []byte) (int, error) {
	return b.wr.Write(d)
}
