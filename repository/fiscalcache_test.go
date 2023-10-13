package repository_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo"
	"github.com/suyono3484/transactiondemo/repository"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	app := &transactiondemo.App{}
	repo := repository.New(app)

	date1, _ := time.Parse(record.FiscalDateFormat, time.Now().AddDate(0, -1, 0).Format(record.FiscalDateFormat))
	date2, _ := time.Parse(record.FiscalDateFormat, time.Now().AddDate(0, 0, -7).Format(record.FiscalDateFormat))
	rate1 := 20.5
	rate2 := 15.25
	cDesc := "Canada-Dollar"
	repo.CacheSetExchangeRate(cDesc, record.FiscalDate(date1), rate1)
	repo.CacheSetExchangeRate(cDesc, record.FiscalDate(date2), rate2)

	d, r, e := repo.CacheGetExchangeRate(cDesc, time.Now().AddDate(0, -6, 0), time.Now())
	if e != nil {
		t.Fatal(e)
	}

	assert.Equal(t, record.FiscalDate(date2), d)
	assert.Equal(t, rate2, r)
}
