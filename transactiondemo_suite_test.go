package transactiondemo_test

import (
	"github.com/suyono3484/transactiondemo"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	repoModule "github.com/suyono3484/transactiondemo/repository"
	tx "github.com/suyono3484/transactiondemo/transaction"
)

func TestTransactionDemo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transaction Demo Suite")
}

var _ = Describe("Transaction", func() {
	var (
		transaction *tx.TxModule
		repo        *repoModule.RepoModule
		app         *transactiondemo.App
	)
	BeforeEach(func() {
		app = &transactiondemo.App{
			AppSkipFile: true,
		}

		repo = repoModule.New(app)
		app.AppRepo = repo

		transaction = tx.New(app)
	})

	When("add a valid new transaction", func() {
		It("accepts a description, a transaction date, and a transaction amount", func() {
			Expect(transaction.Add("description", time.Now().Format(record.FiscalDateFormat), "12.15")).NotTo(HaveOccurred())
		})

		Context("working with file", func() {
			var (
				fileName string
			)

			BeforeEach(func() {
				createTempFile(&fileName)
			})

			AfterEach(func() {
				_ = os.Remove(fileName)
			})

			It("stores the transaction into a file", func() {
				transaction = writeAndReread(fileName)

				list := transaction.List()
				Expect(list).To(HaveLen(2))
			})
		})

		It("assigns a unique identifier for each transaction", func() {
			_ = transaction.Add("transaction 1", time.Now().Format(record.FiscalDateFormat), "12.15")
			_ = transaction.Add("transaction 2", time.Now().Format(record.FiscalDateFormat), "12.17")

			list := transaction.List()
			GinkgoWriter.Printf("data: %+v\n", list)

			m := make(map[string]any)
			for _, r := range list {
				m[r.ID] = nil
			}

			Expect(m).To(HaveLen(2))
		})
	})

	When("add an invalid new transaction", func() {
		validDescription := "transaction 1"
		validTime := time.Now().Format(record.FiscalDateFormat)
		validAmount := "12.15"

		It("rejects a description with length over 50 characters", func() {
			longText := `Lorem ipsum dolor sit amet, consectetur adipiscing e`
			Expect(transaction.Add(longText, validTime, validAmount)).To(HaveOccurred())
		})

		It("rejects an invalid/malformed date", func() {
			Expect(transaction.Add(validDescription, "invalid date", validAmount)).To(HaveOccurred())
		})

		It("rejects an invalid amount", func() {
			Expect(transaction.Add(validDescription, validTime, "three dollars and fifty cents")).To(HaveOccurred())
		})
	})
})

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
