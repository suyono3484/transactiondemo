package repository

import (
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"sync"
	"time"
)

type fiscalCache struct {
	createdAt time.Time
	table     map[string]map[record.FiscalDate]float64
	mtx       *sync.RWMutex
}

func (r *RepoModule) CacheGetExchangeRate(cDesc string, start, txDate time.Time) (date record.FiscalDate, rate float64, err error) {
	r.fiscalCache.mtx.RLock()
	defer r.fiscalCache.mtx.RUnlock()

	var (
		ok       bool
		currency map[record.FiscalDate]float64
	)

	if time.Now().Add(-12 * time.Hour).After(r.fiscalCache.createdAt) {
		err = types.CacheNoDataError
		return
	}

	currency, ok = r.fiscalCache.table[cDesc]
	if !ok {
		err = types.CacheNoDataError
		return
	}

	date = record.FiscalDate(start)
	for fd, c := range currency {
		t := time.Time(fd)
		if start.Before(t) && txDate.After(t) {
			start = t
			date = fd
			rate = c
		}
	}

	return
}

func (r *RepoModule) CacheSetExchangeRate(cDesc string, date record.FiscalDate, rate float64) {
	r.fiscalCache.mtx.Lock()
	defer r.fiscalCache.mtx.Unlock()

	if time.Now().Add(-12 * time.Hour).After(r.fiscalCache.createdAt) {
		r.fiscalCache.table = make(map[string]map[record.FiscalDate]float64)
		r.fiscalCache.createdAt = time.Now()
	}

	var (
		ok       bool
		currency map[record.FiscalDate]float64
	)

	currency, ok = r.fiscalCache.table[cDesc]
	if !ok {
		currency = make(map[record.FiscalDate]float64)
	}
	currency[date] = rate
	r.fiscalCache.table[cDesc] = currency
}
