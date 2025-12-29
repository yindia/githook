package internal

import "expvar"

var (
	requestsTotal = expvar.NewMap("githooks_requests_total")
	parseErrors   = expvar.NewMap("githooks_parse_errors_total")
	publishErrors = expvar.NewMap("githooks_publish_errors_total")
)

func IncRequest(provider string) {
	requestsTotal.Add(provider, 1)
}

func IncParseError(provider string) {
	parseErrors.Add(provider, 1)
}

func IncPublishError(driver string) {
	publishErrors.Add(driver, 1)
}
