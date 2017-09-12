package builder

import (
    "github.com/stretchr/testify/assert"
    "testing"
    "github.com/qct/crypto_coin_api"
)

var b = NewApiBuilder()

func TestApiBuilder_Build(t *testing.T) {
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.OK_CN).GetExchangeName(), "okcoin.cn")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.OK_COM).GetExchangeName(), "okcoin.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.HUOBI).GetExchangeName(), "huobi.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.CHBTC).GetExchangeName(), "chbtc.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.YUNBI).GetExchangeName(), "yunbi.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.POLONIEX).GetExchangeName(), "poloniex.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.COIN_CHECK).GetExchangeName(), "coincheck.com")
    assert.Equal(t, b.ApiKey("").ApiSecretkey("").Build(coinapi.ZAIF).GetExchangeName(), "zaif.jp")
}
