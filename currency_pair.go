package coinapi

import "strings"

// default: upper case
type CurrencyV2 string

type CurrencyPairV2 struct {
    BaseCurrency    CurrencyV2
    CounterCurrency CurrencyV2
}

func NewCurrencyPairV2(base CurrencyV2, counter CurrencyV2) CurrencyPairV2 {
    return CurrencyPairV2{base, counter}
}

func (cp CurrencyPairV2) Symbol() string {
    return cp.CustomSymbol("_", false)
}

func (cp CurrencyPairV2) CustomSymbol(c string, lower bool) string {
    base := string(cp.BaseCurrency)
    counter := string(cp.CounterCurrency)
    if lower {
        base = strings.ToLower(base)
        counter = strings.ToLower(counter)
    }
    return strings.Join([]string{base, counter}, c)
}
