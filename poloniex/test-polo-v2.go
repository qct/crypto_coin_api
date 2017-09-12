package poloniex

import (
    "github.com/qct/crypto_coin_api/builder"
    "github.com/qct/crypto_coin_api"
)

func main() {
    api := builder.NewApiBuilder().Build(coinapi.POLONIEX)
    api.LimitBuy("0.2", "21.0", coinapi.NewCurrencyPairV2("abc", "def"))
}
