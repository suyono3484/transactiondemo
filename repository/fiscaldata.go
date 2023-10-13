package repository

import (
	"encoding/json"
	"fmt"
	"github.com/suyono3484/transactiondemo/transaction/record"
	"io"
	"net/http"
	"time"
)

type RecordContainer struct {
	Data []record.FiscalRecord `json:"data"`
}

func (r *RepoModule) FetchFiscalData(cDesc string, start, txDate time.Time) ([]record.FiscalRecord, error) {
	sortParam := "-record_date"
	fieldsParam := "record_date,country,currency,country_currency_desc,exchange_rate,effective_date"
	dateRangeFilter := fmt.Sprintf("effective_date:gte:%s,effective_date:lte:%s",
		start.Format(record.FiscalDateFormat), txDate.Format(record.FiscalDateFormat))
	currencyFilter := fmt.Sprintf("country_currency_desc:in:(%s)", cDesc)
	filterParam := fmt.Sprintf("%s,%s", currencyFilter, dateRangeFilter)

	var container RecordContainer
	url := fmt.Sprintf("%s?sort=%s&fields=%s&filter=%s", r.config.ExchangeRateURL(), sortParam, fieldsParam, filterParam)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return container.Data, err
	}

	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return container.Data, err
	}

	if resp.StatusCode != http.StatusOK {
		return container.Data, fmt.Errorf("HTTP Status not OK: %d", resp.StatusCode)
	}

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return container.Data, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err = json.Unmarshal(b, &container); err != nil {
		return container.Data, err
	}

	return container.Data, nil
}
