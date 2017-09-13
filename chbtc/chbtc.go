package chbtc

import (
    "net/http"
    . "github.com/qct/crypto_coin_api"
    "fmt"
    "strconv"
    "log"
    "net/url"
    "time"
    "encoding/json"
    "strings"
    "errors"
    "sort"
)

const (
    MARKET_URL                = "http://api.c.com/data/v1/"
    TICKER_API                = "ticker?currency=%s"
    DEPTH_API                 = "depth?currency=%s&size=%d"
    TRADE_URL                 = "https://trade.c.com/api/"
    GET_ACCOUNT_API           = "getAccountInfo"
    GET_ORDER_API             = "getOrder"
    GET_UNFINISHED_ORDERS_API = "getUnfinishedOrdersIgnoreTradeType"
    CANCEL_ORDER_API          = "cancelOrder"
    PLACE_ORDER_API           = "order"
    WITHDRAW_API              = "withdraw"
    CANCEL_WITHDRAW_API       = "cancelWithdraw"
)

type ChbtcApi struct {
    httpClient *http.Client;
    accessKey,
    secretKey string
}

func NewApi(httpClient *http.Client, accessKey, secretKey string) *ChbtcApi {
    return &ChbtcApi{httpClient, accessKey, secretKey};
}

func (c *ChbtcApi) GetDepth(cp CurrencyPairV2, size int) (*Depth, error) {
    resp, err := HttpGet(c.httpClient, MARKET_URL+fmt.Sprintf(DEPTH_API, cp.CustomSymbol("_", true), size))
    if err != nil {
        return nil, err
    }
    asks := resp["asks"].([]interface{})
    bids := resp["bids"].([]interface{})

    depth := new(Depth)
    for _, e := range bids {
        var r DepthRecord
        ee := e.([]interface{})
        r.Amount = ee[1].(float64)
        r.Price = ee[0].(float64)
        depth.BidList = append(depth.BidList, r)
    }
    for _, e := range asks {
        var r DepthRecord
        ee := e.([]interface{})
        r.Amount = ee[1].(float64)
        r.Price = ee[0].(float64)
        depth.AskList = append(depth.AskList, r)
    }
    sort.Sort(depth.AskList)
    return depth, nil
}

func (c *ChbtcApi) LimitBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return c.placeOrder(amount, price, cp, 1)
}

func (c *ChbtcApi) LimitSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return c.placeOrder(amount, price, cp, 0)
}

func (c *ChbtcApi) MarketBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    panic("unsupported the market order")
}

func (c *ChbtcApi) MarketSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    panic("unsupported the market order")
}

func (c *ChbtcApi) CancelOrder(orderId string, cp CurrencyPairV2) (bool, error) {
    params := url.Values{}
    params.Set("method", "cancelOrder")
    params.Set("id", orderId)
    params.Set("currency", cp.CustomSymbol("_", true))
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+CANCEL_ORDER_API, params)
    if err != nil {
        log.Println(err)
        return false, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err)
        return false, err
    }
    code := respMap["code"].(float64)
    if code == 1000 {
        return true, nil
    }
    return false, errors.New(fmt.Sprintf("%.0f", code))
}

func (c *ChbtcApi) GetOneOrder(orderId string, cp CurrencyPairV2) (*OrderV2, error) {
    params := url.Values{}
    params.Set("method", "getOrder")
    params.Set("id", orderId)
    params.Set("currency", cp.CustomSymbol("_", true))
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+GET_ORDER_API, params)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    orderMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &orderMap)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    order := new(OrderV2)
    order.CurrencyPair = cp.CustomSymbol("_", true)
    parseOrder(order, orderMap)
    return order, nil
}

func (c *ChbtcApi) GetUnfinishedOrders(cp CurrencyPairV2) ([]OrderV2, error) {
    params := url.Values{}
    params.Set("method", "getUnfinishedOrdersIgnoreTradeType")
    params.Set("currency", cp.CustomSymbol("_", true))
    params.Set("pageIndex", "1")
    params.Set("pageSize", "100")
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+GET_UNFINISHED_ORDERS_API, params)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    respStr := string(resp)
    if strings.Contains(respStr, "\"code\":3001") {
        log.Println(respStr)
        return nil, nil
    }

    var respArr []interface{}
    err = json.Unmarshal(resp, &respArr)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    var orders []OrderV2
    for _, v := range respArr {
        orderMap := v.(map[string]interface{})
        order := OrderV2{}
        order.CurrencyPair = cp.CustomSymbol("_", true)
        parseOrder(&order, orderMap)
        orders = append(orders, order)
    }
    return orders, nil
}

func (c *ChbtcApi) GetAccount() (*AccountV2, error) {
    params := url.Values{}
    params.Set("method", "getAccountInfo")
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+GET_ACCOUNT_API, params)
    if err != nil {
        return nil, err
    }

    var respMap map[string]interface{}
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println("json unmarshal error")
        return nil, err
    }
    if respMap["code"] != nil && respMap["code"].(float64) != 1000 {
        return nil, errors.New(string(resp))
    }

    acc := new(AccountV2)
    acc.Exchange = CHBTC
    acc.SubAccountsV2 = make(map[string]SubAccountV2)
    resultMap := respMap["result"].(map[string]interface{})
    balanceMap := resultMap["balance"].(map[string]interface{})
    frozenMap := resultMap["frozen"].(map[string]interface{})
    p2pMap := resultMap["p2p"].(map[string]interface{})
    acc.NetAsset = ToFloat64(resultMap["netAssets"])
    acc.Asset = ToFloat64(resultMap["totalAssets"])
    for t, v := range balanceMap {
        vv := v.(map[string]interface{})
        frozen := frozenMap["CNY"].(map[string]interface{})
        subAcc := SubAccountV2{}
        subAcc.Amount = ToFloat64(vv["amount"])
        subAcc.FrozenAmount = ToFloat64(frozen["amount"])
        subAcc.LoanAmount = ToFloat64(p2pMap[fmt.Sprintf("in%s", t)])
        subAcc.Currency = t
        acc.SubAccountsV2[subAcc.Currency] = subAcc
    }
    return acc, nil
}

func (c *ChbtcApi) GetTicker(cp CurrencyPairV2) (*Ticker, error) {
    resp, err := HttpGet(c.httpClient, MARKET_URL+fmt.Sprintf(TICKER_API, cp.CustomSymbol("_", true)))
    if err != nil {
        return nil, err
    }
    tickerMap := resp["ticker"].(map[string]interface{})
    ticker := new(Ticker)
    ticker.Date, _ = strconv.ParseUint(resp["date"].(string), 10, 64)
    ticker.Buy, _ = strconv.ParseFloat(tickerMap["buy"].(string), 64)
    ticker.Sell, _ = strconv.ParseFloat(tickerMap["sell"].(string), 64)
    ticker.Last, _ = strconv.ParseFloat(tickerMap["last"].(string), 64)
    ticker.High, _ = strconv.ParseFloat(tickerMap["high"].(string), 64)
    ticker.Low, _ = strconv.ParseFloat(tickerMap["low"].(string), 64)
    ticker.Vol, _ = strconv.ParseFloat(tickerMap["vol"].(string), 64)
    return ticker, nil
}

func (c *ChbtcApi) Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error) {
    params := url.Values{}
    params.Set("method", "withdraw")
    params.Set("currency", strings.ToLower(currency))
    params.Set("amount", amount)
    params.Set("fees", fees)
    params.Set("receiveAddr", receiveAddr)
    params.Set("safePwd", safePwd)
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+WITHDRAW_API, params)
    if err != nil {
        log.Println("withdraw failed.", err)
        return "", err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err, string(resp))
        return "", err
    }
    if respMap["code"].(float64) == 1000 {
        return respMap["id"].(string), nil
    }
    return "", errors.New(string(resp))
}

func (c *ChbtcApi) GetExchangeName() string {
    return CHBTC
}

func (c *ChbtcApi) GetKlineRecords(cp CurrencyPairV2, period string, size, since int) ([]Kline, error) {
    return nil, nil
}

func (c *ChbtcApi) GetOrderHistory(cp CurrencyPairV2, currentPage, pageSize int) ([]OrderV2, error) {
    return nil, nil
}

func (c *ChbtcApi) GetTrades(cp CurrencyPairV2, since int64) ([]Trade, error) {
    panic("unsupported")
}

func (c *ChbtcApi) CancelWithdraw(id, currency, safePwd string) (bool, error) {
    params := url.Values{}
    params.Set("method", "cancelWithdraw")
    params.Set("currency", strings.ToLower(currency))
    params.Set("downloadId", id)
    params.Set("safePwd", safePwd)
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+CANCEL_WITHDRAW_API, params)
    if err != nil {
        log.Println("cancel withdraw fail.", err)
        return false, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err, string(resp))
        return false, err
    }

    if respMap["code"].(float64) == 1000 {
        return true, nil
    }
    return false, errors.New(string(resp))
}

func (c *ChbtcApi) buildPostForm(postForm *url.Values) error {
    postForm.Set("accesskey", c.accessKey)
    payload := postForm.Encode()
    secretKeySha, _ := GetSHA(c.secretKey)
    sign, err := GetParamHmacMD5Sign(secretKeySha, payload)
    if err != nil {
        return err
    }
    postForm.Set("sign", sign)
    postForm.Set("reqTime", fmt.Sprintf("%d", time.Now().UnixNano()/1000000))
    return nil
}

func parseOrder(order *OrderV2, orderMap map[string]interface{}) {
    order.OrderID, _ = strconv.Atoi(orderMap["id"].(string))
    order.Amount = orderMap["total_amount"].(float64)
    order.DealAmount = orderMap["trade_amount"].(float64)
    order.Price = orderMap["price"].(float64)
    order.Fee = orderMap["fees"].(float64)
    if order.DealAmount > 0 {
        order.AvgPrice = orderMap["trade_money"].(float64) / order.DealAmount
    } else {
        order.AvgPrice = 0
    }

    order.OrderTime = int(orderMap["trade_date"].(float64))
    orType := orderMap["type"].(float64)
    switch orType {
    case 0:
        order.Side = SELL
    case 1:
        order.Side = BUY
    default:
        log.Printf("unknown order type %f", orType)
    }

    status := TradeStatus(orderMap["status"].(float64))
    switch status {
    case 0:
        order.Status = ORDER_UNFINISHED
    case 1:
        order.Status = ORDER_CANCEL
    case 2:
        order.Status = ORDER_FINISH
    case 3:
        order.Status = ORDER_PART_FINISH
    }
}

func (c *ChbtcApi) placeOrder(amount, price string, cp CurrencyPairV2, tradeType int) (*OrderV2, error) {
    params := url.Values{}
    params.Set("method", "order")
    params.Set("price", price)
    params.Set("amount", amount)
    params.Set("currency", cp.CustomSymbol("_", true))
    params.Set("tradeType", fmt.Sprintf("%d", tradeType))
    c.buildPostForm(&params)
    resp, err := HttpPostForm(c.httpClient, TRADE_URL+PLACE_ORDER_API, params)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    code := respMap["code"].(float64)
    if code != 1000 {
        log.Println(string(resp))
        return nil, errors.New(fmt.Sprintf("%.0f", code))
    }

    id := respMap["id"].(string)
    order := new(OrderV2)
    order.Amount, _ = strconv.ParseFloat(amount, 64)
    order.Price, _ = strconv.ParseFloat(price, 64)
    order.Status = ORDER_UNFINISHED
    order.CurrencyPair = cp.CustomSymbol("_", true)
    order.OrderTime = int(time.Now().UnixNano() / 1000000)
    order.OrderID, _ = strconv.Atoi(id)
    switch tradeType {
    case 0:
        order.Side = SELL
    case 1:
        order.Side = BUY
    }
    return order, nil
}
