package repository

import (
	"github.com/suyono3484/transactiondemo/transaction/record"
	"sync"
	"time"
)

type Config interface {
	FilePath() string
	SkipFile() bool
	ExchangeRateURL() string
}

type RepoModule struct {
	config      Config
	fileMtx     *sync.Mutex
	fiscalCache *fiscalCache
}

func New(config Config) *RepoModule {
	return &RepoModule{
		config:  config,
		fileMtx: &sync.Mutex{},
		fiscalCache: &fiscalCache{
			createdAt: time.Now(),
			table:     make(map[string]map[record.FiscalDate]float64),
			mtx:       &sync.RWMutex{},
		},
	}
}
