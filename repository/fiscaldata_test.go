package repository_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo"
	"github.com/suyono3484/transactiondemo/repository"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRepoModule_FetchFiscalData(t *testing.T) {
	setUp, cleanUp, setData := prepareTestServer()
	tsURL := setUp(t)
	defer cleanUp(t)

	app := &transactiondemo.App{
		AppExchangeRateURL: tsURL,
	}
	if os.Getenv("TEST_USING_ACTUAL_URL") != "" {
		app.AppExchangeRateURL = types.ExchangeRateURL
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
		fiscals      []record.FiscalRecord
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
			fiscals: []record.FiscalRecord{
				{
					RecordDate: record.FiscalDate(
						time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)),
					Country:             "Canada",
					Currency:            "Dollar",
					CountryCurrencyDesc: "Canada-Dollar",
					ExchangeRate:        record.ExchangeRate(1.75),
					EffectiveDate: record.FiscalDate(
						time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)),
				},
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
			fiscals: []record.FiscalRecord{},
			wantTrueFunc: func(records []record.FiscalRecord) bool {
				return len(records) == 0
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setData(t, tt.fiscals)
			got, err := repo.FetchFiscalData(tt.args.cDesc, tt.args.start, tt.args.txDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFiscalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.True(t, tt.wantTrueFunc(got))
		})
	}
}

func prepareTestServer() (setUp func(t *testing.T) string, cleanUp func(t *testing.T), setData func(*testing.T, []record.FiscalRecord)) {
	var (
		ts      *httptest.Server
		fiscals []record.FiscalRecord
	)

	setUp = func(t *testing.T) string {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rc := repository.RecordContainer{
				Data: fiscals,
			}

			b, err := json.Marshal(&rc)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
		}))
		return ts.URL
	}

	cleanUp = func(t *testing.T) {
		ts.Close()
	}

	setData = func(t *testing.T, records []record.FiscalRecord) {
		fiscals = records
	}

	return
}

func TestFiscalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	app := &transactiondemo.App{
		AppExchangeRateURL: ts.URL,
	}

	repo := repository.New(app)
	time1 := time.Now()
	_, err := repo.FetchFiscalData("Canada-Dollar", time1.AddDate(0, -6, 0), time1)
	assert.Error(t, err)
}
