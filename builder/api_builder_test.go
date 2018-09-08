package builder

import (
	"cryptocurrency-exchange-api"
	"github.com/stretchr/testify/assert"
	"testing"
)

var b = NewApiBuilder()

func TestApiBuilder_Build(t *testing.T) {
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.OK_CN).GetExchangeName(), "okcoin.cn")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.OK_COM).GetExchangeName(), "okcoin.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.HUOBI).GetExchangeName(), "huobi.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.CHBTC).GetExchangeName(), "chbtc.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.YUNBI).GetExchangeName(), "yunbi.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.POLONIEX).GetExchangeName(), "poloniex.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.COIN_CHECK).GetExchangeName(), "coincheck.com")
	assert.Equal(t, b.ApiKey("").ApiSecretKey("").Build(coinapi.ZAIF).GetExchangeName(), "zaif.jp")
}
