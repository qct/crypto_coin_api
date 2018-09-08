package builder

import (
	"context"
	. "github.com/qct/cryptocurrency-exchange-api"
	"github.com/qct/cryptocurrency-exchange-api/chbtc"
	"github.com/qct/cryptocurrency-exchange-api/okcoin"
	"github.com/qct/cryptocurrency-exchange-api/poloniex"
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

func NewApiBuilder() *ApiBuilder {
	client := http.DefaultClient
	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 4 * time.Second,
	}
	client.Transport = transport
	return &ApiBuilder{client: client}
}

func (b *ApiBuilder) ApiKey(key string) *ApiBuilder {
	b.apiKey = key
	return b
}

func (b *ApiBuilder) ApiSecretKey(key string) *ApiBuilder {
	b.secretKey = key
	return b
}

func (b *ApiBuilder) HttpTimeout(timeout time.Duration) *ApiBuilder {
	b.httpTimeout = timeout
	b.client.Timeout = timeout
	transport := b.client.Transport.(*http.Transport)
	if transport != nil {
		transport.ResponseHeaderTimeout = timeout
		transport.TLSHandshakeTimeout = timeout
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		}
	}
	return b
}

func (b *ApiBuilder) Build(exName string) (api Api) {
	switch exName {
	case OK_CN:
		api = okcoin.NewOkCNApi(b.client, b.apiKey, b.secretKey)
	case CHBTC:
		api = chbtc.NewApi(b.client, b.apiKey, b.secretKey)
	case POLONIEX:
		api = poloniex.New(b.client, b.apiKey, b.secretKey)
		//case OK_COM:
		//    api = okcoin.NewCOM(b.client, b.apiKey, b.secretKey)
		//case COIN_CHECK:
		//    api = coincheck.New(b.client, b.apiKey, b.secretKey)
		//case ZAIF:
		//    api = zaif.New(b.client, b.apiKey, b.secretKey)
		//    api = yunbi.New(b.client, b.apiKey, b.secretKey)
		//case YUNBI:
		//    api = huobi.New(b.client, b.apiKey, b.secretKey)
		//case HUOBI:
	default:
		log.Println("error")
	}
	return api
}

func (b *ApiBuilder) BuildFutureApi(exName string) (api FutureApi) {
	switch exName {
	case OK_EX:
		api = okcoin.NewOkExApi(b.client, b.apiKey, b.secretKey)
	default:
		log.Println("error")
	}
	return api
}
