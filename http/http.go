package http

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/suyono3484/transactiondemo/types"
	"net/http"
)

type Config interface {
	Transaction() types.TxI
}

type Module struct {
	config Config
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func New(config Config) *Module {
	return &Module{
		config: config,
	}
}

func (h *Module) Router() http.Handler {
	router := httprouter.New()

	router.POST("/add", h.AddEndpoint)
	router.GET("/get/:id", h.GetEndpoint)

	return router
}

func (h *Module) AddEndpoint(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	description := r.PostFormValue("description")
	date := r.PostFormValue("date")
	amount := r.PostFormValue("amount")

	err := h.config.Transaction().Add(description, date, amount)
	if err != nil {
		if errors.Is(err, types.InvalidInputError) {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Module) GetEndpoint(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	v := r.URL.Query()
	inTarget, ok := v["target"]
	var target string
	if ok && len(inTarget) > 0 {
		target = inTarget[0]
	} else {
		target = types.DefaultCurrency
	}

	outRec, err := h.config.Transaction().Get(params.ByName("id"), target)
	if err != nil {
		if errors.Is(err, types.RecordNotFound) {
			writeErrorResponse(w, http.StatusNotFound, types.RecordNotFound.Error())
			return
		}

		if errors.Is(err, types.TargetCurrencyUnavailable) {
			writeErrorResponse(w, http.StatusInternalServerError, "the transaction cannot be converted to the target currency")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var b []byte
	if b, err = json.Marshal(&outRec); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func writeErrorResponse(w http.ResponseWriter, responseStatusCode int, message string) {
	var (
		b   []byte
		err error
	)

	if b, err = json.Marshal(&ErrorMessage{
		Error: message,
	}); err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseStatusCode)
	_, _ = w.Write(b)
}
