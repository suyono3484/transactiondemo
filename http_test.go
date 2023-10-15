package transactiondemo_test

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suyono3484/transactiondemo"
	hm "github.com/suyono3484/transactiondemo/http"
	repoModule "github.com/suyono3484/transactiondemo/repository"
	tx "github.com/suyono3484/transactiondemo/transaction"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	goUrl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrInvalidContentType = errors.New("invalid Content-Type")
var _ = Describe("Run HTTP Server", func() {
	var (
		ts          *httptest.Server
		as          *httptest.Server
		dir         string
		err         error
		fileName    string
		app         *transactiondemo.App
		repo        *repoModule.RepoModule
		transaction *tx.TxModule
		httpModule  *hm.Module
		list        []record.TransactionRecord
		outRec      record.ConvertedTransaction
		respCode    int
		respString  string
		fiscals     []record.FiscalRecord
	)

	testExchange := 1.75
	currDesc := "Canada-Dollar"
	amount := 12.15

	BeforeEach(func() {
		fiscals = []record.FiscalRecord{
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
		}

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rc := repoModule.RecordContainer{
				Data: fiscals,
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

		dir, err = os.MkdirTemp("", "test-*")
		Expect(err).ToNot(HaveOccurred())

		fileName = filepath.Join(dir, "data.json")

		app = &transactiondemo.App{
			AppExchangeRateURL: ts.URL,
			AppFilePath:        fileName,
			AppSkipFile:        false,
		}
		repo = repoModule.New(app)
		app.AppRepo = repo
		transaction = tx.New(app)
		app.AppTransaction = transaction
		httpModule = hm.New(app)

		as = httptest.NewServer(httpModule.Router())
	})

	AfterEach(func() {
		as.Close()
		_ = os.RemoveAll(dir)
		ts.Close()
	})

	It("serves http request for adding and getting transaction", func() {
		respCode, respString, err = sendAddRequest(as.URL, "transaction 1",
			time.Now().Format(record.FiscalDateFormat),
			fmt.Sprintf("%f", amount))
		Expect(err).ToNot(HaveOccurred())
		Expect(respCode).To(Equal(http.StatusOK))

		list = transaction.List()
		Expect(list).To(HaveLen(1))

		outRec, respString, err = sendGetRequest(as.URL, list[0].ID, currDesc)
		Expect(err).ToNot(HaveOccurred())
		Expect(outRec.Rate).To(Equal(testExchange))
		Expect(outRec.Converted).To(Equal(math.Round(amount*testExchange*100) / 100))

		outRec, respString, err = sendGetRequest(as.URL, list[0].ID, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(outRec.Rate).To(Equal(float64(1)))
		Expect(outRec.Converted).To(Equal(amount))
	})

	It("returns error response for invalid input when adding", func() {
		respCode, respString, err = sendAddRequest(as.URL, "transaction 1",
			"invalid date",
			fmt.Sprintf("%f", amount))
		GinkgoWriter.Println("response string", respString)
		Expect(err).ToNot(HaveOccurred())
		Expect(respCode).ToNot(Equal(http.StatusOK))
		Expect(errors.Is(err, ErrInvalidContentType)).To(BeFalse())
	})

	It("returns appropriate error message when no record found", func() {
		outRec, respString, err = sendGetRequest(as.URL, "random ID", currDesc)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrInvalidContentType)).To(BeFalse())
	})

	When("no record returned from fiscal data server", func() {
		It("returns appropriate error message", func() {
			fiscals = []record.FiscalRecord{}
			respCode, respString, err = sendAddRequest(as.URL, "transaction 1",
				time.Now().Format(record.FiscalDateFormat),
				fmt.Sprintf("%f", amount))
			Expect(err).ToNot(HaveOccurred())
			Expect(respCode).To(Equal(http.StatusOK))

			list = transaction.List()
			Expect(list).To(HaveLen(1))

			outRec, respString, err = sendGetRequest(as.URL, list[0].ID, currDesc)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrInvalidContentType)).To(BeFalse())
		})
	})
})

func sendAddRequest(url, description, date, amount string) (int, string, error) {
	url = fmt.Sprintf("%s/add", url)
	val := goUrl.Values{}
	val.Set("description", description)
	val.Set("date", date)
	val.Set("amount", amount)

	client := &http.Client{}
	resp, err := client.PostForm(url, val)
	if err != nil {
		return 0, "", err
	}

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if err = checkContentType(resp); err != nil {
			return 0, "", err
		}
	}
	return resp.StatusCode, string(b), nil
}

func sendGetRequest(url, id, targetCurrency string) (outRec record.ConvertedTransaction, respStr string, err error) {
	if targetCurrency == "" {
		url = fmt.Sprintf("%s/get/%s", url, id)
	} else {
		url = fmt.Sprintf("%s/get/%s?target=%s", url, id, targetCurrency)

	}
	client := &http.Client{}

	var resp *http.Response
	resp, err = client.Get(url)
	if err != nil {
		return
	}

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	GinkgoWriter.Println("response data:", string(b))
	if resp.StatusCode != http.StatusOK {
		if err = checkContentType(resp); err != nil {
			err = fmt.Errorf("http response not OK: %d and %w", resp.StatusCode, err)
		} else {
			err = fmt.Errorf("http response not OK: %d", resp.StatusCode)
		}
		respStr = string(b)
		return
	}

	if err = checkContentType(resp); err != nil {
		return
	}

	if err = json.Unmarshal(b, &outRec); err != nil {
		return
	}
	return
}

func checkContentType(resp *http.Response) (err error) {
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		err = fmt.Errorf("unexpected http response Content-Type header: %s; %w", ct, ErrInvalidContentType)
		return
	}
	return nil
}
