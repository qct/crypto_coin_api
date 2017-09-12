package coinapi

import "strings"

type CurrencyV2 string

type CurrencyPairV2 struct {
    BaseCurrency    CurrencyV2
    CounterCurrency CurrencyV2
}

func NewCurrencyPairV2(base CurrencyV2, counter CurrencyV2) CurrencyPairV2 {
    return CurrencyPairV2{base, counter}
}

func (cp CurrencyPairV2) Symbol() string {
    return cp.CustomSymbol("_")
}

func (cp CurrencyPairV2) CustomSymbol(c string) string {
    return strings.Join([]string{string(cp.BaseCurrency), string(cp.CounterCurrency)}, c)
}
