package transactiondemo_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	repoModule "github.com/suyono3484/transactiondemo/repository"
	tx "github.com/suyono3484/transactiondemo/transaction"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/suyono3484/transactiondemo"
)

var _ = Describe("retrieve a transaction using exchange rate", func() {
	It("calculates the converted rate using exchange rate", func() {
		testExchange := 1.75
		currDesc := "Canada-Dollar"
		amount := 12.15
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rc := repoModule.RecordContainer{
				Data: []record.FiscalRecord{
					{
						RecordDate: record.FiscalDate(
							time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)),
						Country:             "Canada",
						Currency:            "Dollar",
						CountryCurrencyDesc: currDesc,
						ExchangeRate:        record.ExchangeRate(testExchange),
						EffectiveDate: record.FiscalDate(
							time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)),
					},
				},
			}

			b, err := json.Marshal(&rc)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			GinkgoWriter.Println("mock data:", string(b))

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
		}))
		defer ts.Close()

		dir, err := os.MkdirTemp("", "test-*")
		Expect(err).ToNot(HaveOccurred())

		fileName := filepath.Join(dir, "data.json")
		defer func() {
			_ = os.RemoveAll(dir)
		}()

		app := &transactiondemo.App{
			AppExchangeRateURL: ts.URL,
			AppFilePath:        fileName,
			AppSkipFile:        false,
		}
		repo := repoModule.New(app)
		app.AppRepo = repo
		transaction := tx.New(app)

		err = transaction.Add("transaction 1",
			time.Now().Format(record.FiscalDateFormat),
			fmt.Sprintf("%f", amount))
		Expect(err).ToNot(HaveOccurred())

		list := transaction.List()
		Expect(list).To(HaveLen(1))

		var outRec record.ConvertedTransaction
		outRec, err = transaction.Get(list[0].ID, currDesc)
		Expect(err).ToNot(HaveOccurred())
		Expect(outRec.Rate).To(Equal(testExchange))
		Expect(outRec.Converted).To(Equal(math.Round(amount*testExchange*100) / 100))
	})
})
