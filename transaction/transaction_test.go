package transaction_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suyono3484/transactiondemo"
	repoModule "github.com/suyono3484/transactiondemo/repository"
	tx "github.com/suyono3484/transactiondemo/transaction"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"github.com/suyono3484/transactiondemo/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTxModule_Get(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	fileName := filepath.Join(dir, "data.json")

	app := &transactiondemo.App{
		AppExchangeRateURL: types.ExchangeRateURL,
		AppFilePath:        fileName,
		AppSkipFile:        false,
	}

	repo := repoModule.New(app)
	app.AppRepo = repo

	transaction := tx.New(app)

	date1 := time.Date(2023, time.October, 10, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2023, time.August, 7, 0, 0, 0, 0, time.UTC)

	amount1 := "12.15"
	amount2 := "34.75"

	if err = transaction.Add("transaction 1", date1.Format(record.FiscalDateFormat), amount1); err != nil {
		t.Fatal(err)
	}

	if err = transaction.Add("transaction 2", date2.Format(record.FiscalDateFormat), amount2); err != nil {
		t.Fatal(err)
	}

	list := transaction.List()
	var or record.ConvertedTransaction
	for _, l := range list {
		or, err = transaction.Get(l.ID, "Canada-Dollar")
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%+v\n", or)
	}
}

func TestTxModule_List(t *testing.T) {
	var (
		transaction *tx.TxModule
		repo        *repoModule.RepoModule
		app         *transactiondemo.App
	)

	app = &transactiondemo.App{
		AppSkipFile: true,
	}

	repo = repoModule.New(app)
	app.AppRepo = repo

	transaction = tx.New(app)

	_ = transaction.Add("transaction 1", time.Now().Format(record.FiscalDateFormat), "12.15")
	_ = transaction.Add("transaction 1", time.Now().Format(record.FiscalDateFormat), "12.17")

	assert.Equal(t, 2, len(transaction.List()))
}

func TestTxModule_Load(t *testing.T) {
	var fileName string

	createTempFile(&fileName)
	defer func() {
		_ = os.Remove(fileName)
	}()

	transaction := writeAndReread(fileName)
	if err := transaction.Load(); err != nil {
		t.Fatal(err)
	}

	list := transaction.List()
	assert.Equal(t, 2, len(list))
}

func createTempFile(name *string) {
	f, err := os.CreateTemp("", "data.*.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()

	*name = f.Name()
}

func writeAndReread(name string) *tx.TxModule {
	app := &transactiondemo.App{
		AppSkipFile: false,
		AppFilePath: name,
	}

	repo := repoModule.New(app)
	app.AppRepo = repo

	transaction := tx.New(app)
	_ = transaction.Add("transaction 1", time.Now().Format(record.FiscalDateFormat), "12.15")
	_ = transaction.Add("transaction 2", time.Now().Format(record.FiscalDateFormat), "12.17")

	app = &transactiondemo.App{
		AppSkipFile: false,
		AppFilePath: name,
	}

	repo = repoModule.New(app)
	app.AppRepo = repo

	transaction = tx.New(app)
	if err := transaction.Load(); err != nil {
		panic(err)
	}

	return transaction
}
