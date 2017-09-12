package builder

import (
    "context"
    . "github.com/qct/crypto_coin_api"
    "github.com/qct/crypto_coin_api/chbtc"
    "github.com/qct/crypto_coin_api/coincheck"
    "github.com/qct/crypto_coin_api/huobi"
    "github.com/qct/crypto_coin_api/okcoin"
    "github.com/qct/crypto_coin_api/poloniex"
    "github.com/qct/crypto_coin_api/yunbi"
    "github.com/qct/crypto_coin_api/zaif"
    "log"
    "net"
    "net/http"
    "time"
)

type ApiBuilder struct {
    client      *http.Client
    httpTimeout time.Duration
    apiKey      string
    secretKey   string
}

func NewApiBuilder() (builder *ApiBuilder) {
    _client := http.DefaultClient
    transport := &http.Transport{
        MaxIdleConns:    10,
        IdleConnTimeout: 4 * time.Second,
    }
    _client.Transport = transport
    return &ApiBuilder{client: _client}
}

func (builder *ApiBuilder) ApiKey(key string) (_builder *ApiBuilder) {
    builder.apiKey = key
    return builder
}

func (builder *ApiBuilder) ApiSecretkey(key string) (_builder *ApiBuilder) {
    builder.secretKey = key
    return builder
}

func (builder *ApiBuilder) HttpTimeout(timeout time.Duration) (_builder *ApiBuilder) {
    builder.httpTimeout = timeout
    builder.client.Timeout = timeout
    transport := builder.client.Transport.(*http.Transport)
    if transport != nil {
        transport.ResponseHeaderTimeout = timeout
        transport.TLSHandshakeTimeout = timeout
        transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
            return net.DialTimeout(network, addr, timeout)
        }
    }
    return builder
}

func (builder *ApiBuilder) Build(exName string) (api Api) {
    switch exName {
    case OK_CN:
        api = okcoin.New(builder.client, builder.apiKey, builder.secretKey)
    case HUOBI:
        api = huobi.New(builder.client, builder.apiKey, builder.secretKey)
    case CHBTC:
        api = chbtc.New(builder.client, builder.apiKey, builder.secretKey)
    case YUNBI:
        api = yunbi.New(builder.client, builder.apiKey, builder.secretKey)
    case POLONIEX:
        api = poloniex.NewPoloApi(builder.client, builder.apiKey, builder.secretKey)
    case OK_COM:
        api = okcoin.NewCOM(builder.client, builder.apiKey, builder.secretKey)
    case COIN_CHECK:
        api = coincheck.New(builder.client, builder.apiKey, builder.secretKey)
    case ZAIF:
        api = zaif.New(builder.client, builder.apiKey, builder.secretKey)
    default:
        log.Println("error")
    }
    return api
}
