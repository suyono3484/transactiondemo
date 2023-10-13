package types

import (
	"github.com/suyono3484/transactiondemo/transaction/record"
	"time"
)

type TxI interface {
	Load() error
	Add(description, date, amount string) error
	List() []record.TransactionRecord
	Get(id, targetCurrency string) (outRec record.ConvertedTransaction, err error)
}

type RepoI interface {
	Open() RepoHandle
	CacheGetExchangeRate(cDesc string, start, txDate time.Time) (date record.FiscalDate, rate float64, err error)
	CacheSetExchangeRate(cDesc string, date record.FiscalDate, rate float64)
	FetchFiscalData(cDesc string, start, txDate time.Time) ([]record.FiscalRecord, error)
}

type RepoHandle interface {
	AppendRecords(records []record.TransactionRecord) (int, error)
	ReadRecords(records []record.TransactionRecord) (int, error)
	Close() error
}
