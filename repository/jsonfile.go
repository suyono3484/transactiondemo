package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"io/fs"
	"os"
)

type handleState int

const (
	idle handleState = iota
	read
	write
)

type Handle struct {
	activeFileHandle *os.File
	fileHandleState  handleState
	module           *RepoModule
}

func (r *RepoModule) Open() types.RepoHandle {
	r.fileMtx.Lock()
	return &Handle{
		activeFileHandle: nil,
		fileHandleState:  idle,
		module:           r,
	}
}

func (h *Handle) Close() (err error) {
	defer h.module.fileMtx.Unlock()
	if h.activeFileHandle != nil {
		err = h.activeFileHandle.Close()
		h.activeFileHandle = nil
	}

	h.fileHandleState = idle

	return
}

func (h *Handle) ReadRecords(records []record.TransactionRecord) (int, error) {
	if h.module.config.SkipFile() || cap(records) == 0 {
		return 0, nil
	}

	if h.fileHandleState == write {
		panic("invalid file handle state")
	}

	var (
		err   error
		scan  *bufio.Scanner
		rec   record.TransactionRecord
		index int
	)

	if h.activeFileHandle == nil {
		h.activeFileHandle, err = os.Open(h.module.config.FilePath())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return 0, nil
			}
			return 0, err
		}
		h.fileHandleState = read
	}

	scan = bufio.NewScanner(h.activeFileHandle)
scanLoop:
	for scan.Scan() {
		if err = json.Unmarshal(scan.Bytes(), &rec); err != nil {
			return index, err
		}

		records[index] = rec
		index++
		if index == cap(records) {
			break scanLoop
		}
	}

	if err = scan.Err(); err != nil {
		return index, err
	}

	return index, err
}

func (h *Handle) AppendRecords(records []record.TransactionRecord) (int, error) {
	if h.module.config.SkipFile() || len(records) == 0 {
		return 0, nil
	}

	if h.fileHandleState == read {
		panic("invalid file handle state")
	}

	var (
		err   error
		index int
		rec   record.TransactionRecord
		bb    []byte
	)

	if h.activeFileHandle == nil {
		h.activeFileHandle, err = os.OpenFile(h.module.config.FilePath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return 0, err
		}
		h.fileHandleState = write
	}

	for index, rec = range records {
		if bb, err = json.Marshal(&rec); err != nil {
			return index, err
		}

		if _, err = fmt.Fprintln(h.activeFileHandle, string(bb)); err != nil {
			return index, err
		}
	}
	index = len(records)

	return index, err
}
