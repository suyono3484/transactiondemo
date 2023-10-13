package transaction

import (
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"math"
	"strconv"
	"sync"
	"time"
)

type Config interface {
	Repo() types.RepoI
}

type TxModule struct {
	config   Config
	table    map[string]record.TransactionRecord
	tableMtx *sync.RWMutex
}

func New(config Config) *TxModule {
	return &TxModule{
		config:   config,
		table:    make(map[string]record.TransactionRecord),
		tableMtx: &sync.RWMutex{},
	}
}

func (t *TxModule) Load() error {
	var (
		n   int
		err error
		rec record.TransactionRecord
		h   types.RepoHandle
	)
	buf := make([]record.TransactionRecord, 10)

	t.tableMtx.Lock()
	defer t.tableMtx.Unlock()

	h = t.config.Repo().Open()
	defer func() {
		_ = h.Close()
	}()
readRecords:
	for {
		n, err = h.ReadRecords(buf)
		if err != nil {
			return err
		}

		if n == 0 {
			break readRecords
		}

		for _, rec = range buf[:n] {
			t.table[rec.ID] = rec
		}
	}
	return nil
}

func (t *TxModule) Add(description, date, amount string) error {
	if len(description) > 50 {
		return fmt.Errorf("%w: description is longer than 50 character", types.InvalidInputError)
	}

	var (
		tDate   time.Time
		err     error
		fAmount float64
		rec     record.TransactionRecord
		ok      bool
		h       types.RepoHandle
	)

	tDate, err = time.Parse(record.FiscalDateFormat, date)
	if err != nil {
		return fmt.Errorf("%w: invalid date %w", types.InvalidInputError, err)
	}

	fAmount, err = strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid amount %w", types.InvalidInputError, err)
	}

	rec = record.TransactionRecord{
		Description: description,
		Date:        record.FiscalDate(tDate),
		Amount:      fAmount,
	}
	rec.ID = fmt.Sprintf("%x",
		xxhash.Sum64String(
			fmt.Sprintf("%s%s%f", description, tDate.Format(time.RFC3339), fAmount)))

	t.tableMtx.Lock()
	defer t.tableMtx.Unlock()

	if _, ok = t.table[rec.ID]; !ok {
		h = t.config.Repo().Open()
		defer func() {
			_ = h.Close()
		}()

		_, err = h.AppendRecords([]record.TransactionRecord{rec})
		if err != nil {
			return fmt.Errorf("%w: %w", types.ServerError, err)
		}

		t.table[rec.ID] = rec
	}

	return nil
}

func (t *TxModule) List() []record.TransactionRecord {
	t.tableMtx.RLock()
	defer t.tableMtx.RUnlock()

	recs := make([]record.TransactionRecord, 0)

	for _, rec := range t.table {
		recs = append(recs, rec)
	}

	return recs
}

func (t *TxModule) Get(id, targetCurrency string) (outRec record.ConvertedTransaction, err error) {
	t.tableMtx.RLock()
	defer t.tableMtx.RUnlock()

	var (
		ok     bool
		rec    record.TransactionRecord
		txDate time.Time
		start  time.Time
		frecs  []record.FiscalRecord
	)

	rec, ok = t.table[id]
	if !ok {
		err = types.RecordNotFound
		return
	}

	outRec.TransactionRecord = rec
	if targetCurrency == types.DefaultCurrency {
		outRec.Rate = 1
		outRec.Converted = math.Round(rec.Amount*100) / 100
		return
	}

	txDate = time.Time(rec.Date)
	start = txDate.AddDate(0, -6, 0)
	_, outRec.Rate, err = t.config.Repo().CacheGetExchangeRate(targetCurrency, start, txDate)
	if err != nil {
		frecs, err = t.config.Repo().FetchFiscalData(targetCurrency, start, txDate)
		if err != nil {
			return
		}

		if len(frecs) == 0 {
			err = types.TargetCurrencyUnavailable
			return
		}

		t.config.Repo().CacheSetExchangeRate(targetCurrency, frecs[0].EffectiveDate, frecs[0].ExchangeRate.Float())
		outRec.Rate = frecs[0].ExchangeRate.Float()
		outRec.Converted = math.Round(outRec.Amount*outRec.Rate*100) / 100
		return
	}

	outRec.Converted = math.Round(outRec.Amount*outRec.Rate*100) / 100
	return
}
