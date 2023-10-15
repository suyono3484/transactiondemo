package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"testing"
	"time"
)

type testConfigRepo struct {
	filePath        string
	skipFile        bool
	exchangeRateURL string
}

func (tr *testConfigRepo) FilePath() string {
	return tr.filePath
}

func (tr *testConfigRepo) SkipFile() bool {
	return tr.skipFile
}

func (tr *testConfigRepo) ExchangeRateURL() string {
	return tr.exchangeRateURL
}

func TestCacheExpiry(t *testing.T) {
	rm := New(&testConfigRepo{})

	date1 := record.FiscalDate(time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC))
	date2 := record.FiscalDate(time.Date(2023, time.June, 30, 0, 0, 0, 0, time.UTC))
	cDesc := "Singapore-Dollar"
	rate1 := 1.57
	rate2 := 1.79
	rm.CacheSetExchangeRate(cDesc, date1, rate1)
	rm.CacheSetExchangeRate(cDesc, date2, rate2)

	txDate1 := time.Date(2023, time.October, 3, 0, 0, 0, 0, time.UTC)
	resultDate, resultRate, err := rm.CacheGetExchangeRate(cDesc, txDate1.AddDate(0, -6, 0), txDate1)
	assert.NoError(t, err)
	assert.Equal(t, rate1, resultRate)
	assert.Equal(t, date1, resultDate)

	txDate2 := time.Date(2023, time.August, 3, 0, 0, 0, 0, time.UTC)
	resultDate, resultRate, err = rm.CacheGetExchangeRate(cDesc, txDate2.AddDate(0, -6, 0), txDate2)
	assert.NoError(t, err)
	assert.Equal(t, rate2, resultRate)
	assert.Equal(t, date2, resultDate)

	rm.fiscalCache.createdAt = rm.fiscalCache.createdAt.Add(-13 * time.Hour)
	_, _, err = rm.CacheGetExchangeRate(cDesc, txDate2.AddDate(0, -6, 0), txDate2)
	assert.Error(t, err)
	rm.CacheSetExchangeRate(cDesc, date1, rate1)

	_, _, err = rm.CacheGetExchangeRate("Thailand-Baht", txDate2.AddDate(0, -6, 0), txDate2)
	assert.Error(t, err)
}
