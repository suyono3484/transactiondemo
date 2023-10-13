package transactiondemo

import (
	"github.com/suyono3484/transactiondemo/types"
)

type App struct {
	AppSkipFile        bool
	AppFilePath        string
	AppRepo            types.RepoI
	AppExchangeRateURL string
	AppTransaction     types.TxI
}

func (a *App) SkipFile() bool {
	return a.AppSkipFile
}

func (a *App) FilePath() string {
	return a.AppFilePath
}

func (a *App) Repo() types.RepoI {
	return a.AppRepo
}

func (a *App) ExchangeRateURL() string {
	return a.AppExchangeRateURL
}

func (a *App) Transaction() types.TxI {
	return a.AppTransaction
}
