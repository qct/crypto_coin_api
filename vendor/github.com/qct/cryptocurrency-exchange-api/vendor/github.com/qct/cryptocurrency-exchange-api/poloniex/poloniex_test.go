package poloniex

import (
	"github.com/qct/cryptocurrency-exchange-api"
	"github.com/qct/cryptocurrency-exchange-api/builder"
)

func main() {
	api := builder.NewApiBuilder().Build(coinapi.POLONIEX)
	api.LimitBuy("0.2", "21.0", coinapi.NewCurrencyPair("abc", "def"))
}
