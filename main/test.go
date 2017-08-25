package main

import (
    "github.com/qct/crypto_coin_api/poloniex"
    "net/http"
    "fmt"
)

func main() {
    poloV2 := poloniex.NewPoloniexV2(http.DefaultClient, "", "")
    if depth, err := poloV2.GetDepth(20, "", ""); err != nil {
        fmt.Println(err)
    }else {
        fmt.Println(depth)
    }
}