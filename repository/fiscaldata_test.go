package repository_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo"
	"github.com/suyono3484/transactiondemo/repository"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"testing"
	"time"
)

func TestRepoModule_FetchFiscalData(t *testing.T) {
	app := &transactiondemo.App{
		AppExchangeRateURL: types.ExchangeRateURL,
	}

	repo := repository.New(app)

	type args struct {
		cDesc  string
		start  time.Time
		txDate time.Time
	}

	tests := []struct {
		name         string
		args         args
		wantTrueFunc func([]record.FiscalRecord) bool
		wantErr      bool
	}{
		{
			name: "positive; expect some data",
			args: args{
				cDesc:  "Canada-Dollar",
				start:  time.Now().AddDate(0, -6, 0),
				txDate: time.Now(),
			},
			wantTrueFunc: func(records []record.FiscalRecord) bool {
				return len(records) >= 1 && len(records) <= 2
			},
			wantErr: false,
		},
		{
			name: "negative; expect no data",
			args: args{
				cDesc:  "NoCountry-NoCurrency",
				start:  time.Now().AddDate(0, -6, 0),
				txDate: time.Now(),
			},
			wantTrueFunc: func(records []record.FiscalRecord) bool {
				return len(records) == 0
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FetchFiscalData(tt.args.cDesc, tt.args.start, tt.args.txDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFiscalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.True(t, tt.wantTrueFunc(got))
		})
	}
}
