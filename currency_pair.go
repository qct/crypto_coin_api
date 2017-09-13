package coinapi

import "strings"

// default: upper case
type Currency string

type CurrencyPair struct {
	BaseCurrency    Currency
	CounterCurrency Currency
}

func NewCurrencyPair(base Currency, counter Currency) CurrencyPair {
	return CurrencyPair{base, counter}
}

func (cp CurrencyPair) Symbol() string {
	return cp.CustomSymbol("_", false)
}

func (cp CurrencyPair) CustomSymbol(c string, lower bool) string {
	base := string(cp.BaseCurrency)
	counter := string(cp.CounterCurrency)
	if lower {
		base = strings.ToLower(base)
		counter = strings.ToLower(counter)
	}
	return strings.Join([]string{base, counter}, c)
}
