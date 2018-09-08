package poloniex

import (
	"github.com/qct/crypto_coin_api"
	"github.com/qct/crypto_coin_api/builder"
)

func main() {
	api := builder.NewApiBuilder().Build(coinapi.POLONIEX)
	api.LimitBuy("0.2", "21.0", coinapi.NewCurrencyPair("abc", "def"))
}
