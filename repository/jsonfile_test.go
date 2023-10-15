package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	config := &testConfigRepo{}
	rm := New(config)

	setUp := func(t *testing.T) {
		dir, err := os.MkdirTemp("", "test-*")
		if err != nil {
			t.Fatal(err)
		}
		config.filePath = filepath.Join(dir, "data.json")
	}

	cleanUp := func(t *testing.T) {
		if config.filePath != "" {
			_ = os.Remove(config.filePath)
		}
	}

	t.Run("open and close", func(t *testing.T) {
		setUp(t)
		defer cleanUp(t)

		handle := rm.Open()
		err := handle.Close()
		assert.NoError(t, err)
	})

	t.Run("write and re-read", func(t *testing.T) {
		setUp(t)
		defer cleanUp(t)

		var (
			handle types.RepoHandle
			err    error
			nrows  int
		)
		func() {
			handle = rm.Open()
			defer func() {
				err = handle.Close()
				assert.NoError(t, err)
			}()

			nrows, err = handle.AppendRecords([]record.TransactionRecord{
				{
					ID:          "id 1",
					Description: "description 1",
					Date:        record.FiscalDate(time.Now()),
					Amount:      0.5,
				},
				{
					ID:          "id 2",
					Description: "description 2",
					Date:        record.FiscalDate(time.Now()),
					Amount:      1.5,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, 2, nrows)
		}()

		func() {
			handle = rm.Open()
			defer func() {
				err = handle.Close()
				assert.NoError(t, err)
			}()

			recs := make([]record.TransactionRecord, 10)
			nrows, err = handle.ReadRecords(recs)
			assert.NoError(t, err)
			assert.Equal(t, 2, nrows)
		}()

	})

	t.Run("read without prior write (no file)", func(t *testing.T) {
		setUp(t)
		defer cleanUp(t)

		var err error
		handle := rm.Open()
		defer func() {
			err = handle.Close()
			assert.NoError(t, err)
		}()

		recs := make([]record.TransactionRecord, 10)
		_, err = handle.ReadRecords(recs)
		assert.NoError(t, err)
	})

	t.Run("read with corrupt file", func(t *testing.T) {
		setUp(t)
		defer cleanUp(t)

		var err error
		func() {
			var f *os.File
			if f, err = os.OpenFile(config.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = f.Close()
			}()

			if _, err = f.Write([]byte("some garbage")); err != nil {
				t.Fatal(err)
			}
		}()

		handle := rm.Open()
		defer func() {
			err = handle.Close()
			assert.NoError(t, err)
		}()

		recs := make([]record.TransactionRecord, 10)
		_, err = handle.ReadRecords(recs)
		assert.Error(t, err)
	})

	t.Run("read while write", func(t *testing.T) {
		setUp(t)
		defer cleanUp(t)

		assert.Panics(t, func() {
			var err error
			handle := rm.Open()
			defer func() {
				err = handle.Close()
				assert.NoError(t, err)
			}()

			_, err = handle.AppendRecords([]record.TransactionRecord{
				{
					ID:          "some id",
					Description: "some description",
					Date:        record.FiscalDate(time.Now()),
					Amount:      0.5,
				},
			})
			assert.NoError(t, err)

			recs := make([]record.TransactionRecord, 10)
			_, _ = handle.ReadRecords(recs)
		})
	})
}
