package types

import "errors"

var (
	InvalidInputError         = errors.New("invalid input")
	ServerError               = errors.New("server")
	CacheNoDataError          = errors.New("cache no data or too old")
	RecordNotFound            = errors.New("record not found")
	TargetCurrencyUnavailable = errors.New("target currency unavailable")
)
