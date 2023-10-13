package main

import (
	"context"
	"errors"
	"github.com/suyono3484/transactiondemo"
	hm "github.com/suyono3484/transactiondemo/http"
	repoModule "github.com/suyono3484/transactiondemo/repository"
	"github.com/suyono3484/transactiondemo/transaction"
	"github.com/suyono3484/transactiondemo/types"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	app := &transactiondemo.App{
		AppFilePath:        "data.json",
		AppSkipFile:        false,
		AppExchangeRateURL: types.ExchangeRateURL,
	}
	repo := repoModule.New(app)
	app.AppRepo = repo
	tx := transaction.New(app)
	err := tx.Load()
	if err != nil {
		log.Fatal("loading persistence data:", err)
	}
	app.AppTransaction = tx
	httpModule := hm.New(app)

	srv := &http.Server{
		Handler: httpModule.Router(),
	}
	idleConnClosed := make(chan any)

	var l net.Listener
	l, err = net.Listen("tcp", "")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutting down: %v\n", err)
		}
		close(idleConnClosed)
	}()

	log.Printf("HTTP server is listening on %v", l.Addr())
	if err = srv.Serve(l); errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-idleConnClosed
}
