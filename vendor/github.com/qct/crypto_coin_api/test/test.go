package main

import (
	"../chbtc"
	"../okcoin"
	"fmt"
	"github.com/qct/crypto_coin_api"
	"net/http"
)

func main() {
	futureApi := okcoin.NewFuture(http.DefaultClient, "", "")
	depth, _ := futureApi.GetFutureDepth(coinapi.BTC_USD, "quarter", 50)
	fmt.Println(depth.AskList.Len())

	chbtcApi := chbtc.New(http.DefaultClient, "", "")
	chbtcDepth, _ := chbtcApi.GetDepth(50, coinapi.BTC_CNY)
	fmt.Println(chbtcDepth.AskList.Len())
}
