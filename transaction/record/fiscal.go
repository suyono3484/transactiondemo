package record

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ExchangeRate float64

type FiscalRecord struct {
	RecordDate          FiscalDate   `json:"record_date"`
	Country             string       `json:"country"`
	Currency            string       `json:"currency"`
	CountryCurrencyDesc string       `json:"country_currency_desc"`
	ExchangeRate        ExchangeRate `json:"exchange_rate"`
	EffectiveDate       FiscalDate   `json:"effective_date"`
}

func (e *ExchangeRate) UnmarshalJSON(b []byte) error {
	var (
		s   string
		err error
		f   float64
	)
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	if f, err = strconv.ParseFloat(s, 64); err != nil {
		return err
	}

	*e = ExchangeRate(f)
	return nil
}

func (e *ExchangeRate) MarshalJSON() ([]byte, error) {
	f := float64(*e)
	s := fmt.Sprintf("%f", f)

	return json.Marshal(s)
}

func (e ExchangeRate) Float() float64 {
	return float64(e)
}
