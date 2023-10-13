package record

import (
	"encoding/json"
	"time"
)

const FiscalDateFormat = "2006-01-02"

type FiscalDate time.Time

type TransactionRecord struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Date        FiscalDate `json:"date"`
	Amount      float64    `json:"amount"`
}

type ConvertedTransaction struct {
	TransactionRecord
	Rate      float64 `json:"rate"`
	Converted float64 `json:"converted"`
}

func (f *FiscalDate) UnmarshalJSON(b []byte) error {
	var (
		s   string
		err error
		t   time.Time
	)
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	if t, err = time.Parse(FiscalDateFormat, s); err != nil {
		return err
	}

	*f = FiscalDate(t)
	return nil
}

func (f *FiscalDate) MarshalJSON() ([]byte, error) {
	t := time.Time(*f)
	s := t.Format(FiscalDateFormat)

	return json.Marshal(s)
}

func (f FiscalDate) Date() time.Time {
	return time.Time(f)
}

func (f FiscalDate) String() string {
	return f.Date().String()
}
