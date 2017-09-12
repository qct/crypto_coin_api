package okcoin

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"
    "strings"

    . "github.com/qct/crypto_coin_api"
    "sort"
)

const (
    EXCHANGE_NAME_CN = "okcoin.cn"
    URL_BASE         = "https://www.okcoin.cn/api/v1/"
    URL_TICKER       = "ticker.do"
    URL_DEPTH        = "depth.do"
    URL_TRADES       = "trades.do"
    URL_KLINE        = "kline.do?symbol=%s&type=%s&size=%d&since=%d"

    URL_USERINFO      = "userinfo.do"
    URL_TRADE         = "trade.do"
    URL_CANCEL_ORDER  = "cancel_order.do"
    URL_ORDER_INFO    = "order_info.do"
    URL_ORDERS_INFO   = "orders_info.do"
    ORDER_HISTORY_URI = "order_history.do"
    TRADE_URI         = "trade_history.do"
    WITHDRAW          = "withdraw.do"
)

type OkCNApi struct {
    client       *http.Client
    api_key      string
    secret_key   string
    api_base_url string
}

func New(client *http.Client, apiKey, secretKey string) *OkCNApi {
    return &OkCNApi{client, apiKey, secretKey,URL_BASE}
}

func (o *OkCNApi) GetDepth(cp CurrencyPairV2, size int) (*Depth, error) {
    var depth Depth

    url := o.api_base_url + URL_DEPTH + "?symbol=" + currencyPair2String(currency) + "&size=" + strconv.Itoa(size)
    bodyDataMap, err := HttpGet(o.client, url)
    if err != nil {
        return nil, err
    }

    if bodyDataMap["result"] != nil && !bodyDataMap["result"].(bool) {
        return nil, errors.New(fmt.Sprintf("%.0f", bodyDataMap["error_code"].(float64)))
    }

    for _, v := range bodyDataMap["asks"].([]interface{}) {
        var dr DepthRecord
        for i, vv := range v.([]interface{}) {
            switch i {
            case 0:
                dr.Price = vv.(float64)
            case 1:
                dr.Amount = vv.(float64)
            }
        }
        depth.AskList = append(depth.AskList, dr)
    }

    for _, v := range bodyDataMap["bids"].([]interface{}) {
        var dr DepthRecord
        for i, vv := range v.([]interface{}) {
            switch i {
            case 0:
                dr.Price = vv.(float64)
            case 1:
                dr.Amount = vv.(float64)
            }
        }
        depth.BidList = append(depth.BidList, dr)
    }

    sort.Sort(depth.AskList)
    return &depth, nil
}

func (o *OkCNApi) LimitBuy(amount, price string, currency CurrencyPair) (*Order, error) {
    return o.placeOrder("buy", amount, price, currency)
}

func (o *OkCNApi) LimitSell(amount, price string, currency CurrencyPair) (*Order, error) {
    return o.placeOrder("sell", amount, price, currency)
}

func (o *OkCNApi) MarketBuy(amount, price string, currency CurrencyPair) (*Order, error) {
    return o.placeOrder("buy_market", amount, price, currency)
}

func (o *OkCNApi) MarketSell(amount, price string, currency CurrencyPair) (*Order, error) {
    return o.placeOrder("sell_market", amount, price, currency)
}

func (o *OkCNApi) CancelOrder(orderId string, currency CurrencyPair) (bool, error) {
    postData := url.Values{}
    postData.Set("order_id", orderId)
    postData.Set("symbol", currencyPair2String(currency))

    o.buildPostForm(&postData)

    body, err := HttpPostForm(o.client, o.api_base_url+URL_CANCEL_ORDER, postData)

    if err != nil {
        return false, err
    }

    var respMap map[string]interface{}

    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return false, err
    }

    if !respMap["result"].(bool) {
        return false, errors.New(string(body))
    }

    return true, nil
}

func (o *OkCNApi) getOrders(orderId string, currency CurrencyPair) ([]Order, error) {
    postData := url.Values{}
    postData.Set("order_id", orderId)
    postData.Set("symbol", currencyPair2String(currency))

    o.buildPostForm(&postData)

    body, err := HttpPostForm(o.client, o.api_base_url+URL_ORDER_INFO, postData)
    //println(string(body))
    if err != nil {
        return nil, err
    }

    var respMap map[string]interface{}

    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return nil, err
    }

    if !respMap["result"].(bool) {
        return nil, errors.New(string(body))
    }

    orders := respMap["orders"].([]interface{})

    var orderAr []Order
    for _, v := range orders {
        orderMap := v.(map[string]interface{})

        var order Order
        order.Currency = currency
        order.OrderID = int(orderMap["order_id"].(float64))
        order.Amount = orderMap["amount"].(float64)
        order.Price = orderMap["price"].(float64)
        order.DealAmount = orderMap["deal_amount"].(float64)
        order.AvgPrice = orderMap["avg_price"].(float64)
        order.OrderTime = int(orderMap["create_date"].(float64))

        //status:-1:已撤销  0:未成交  1:部分成交  2:完全成交 4:撤单处理中
        switch int(orderMap["status"].(float64)) {
        case -1:
            order.Status = ORDER_CANCEL
        case 0:
            order.Status = ORDER_UNFINISH
        case 1:
            order.Status = ORDER_PART_FINISH
        case 2:
            order.Status = ORDER_FINISH
        case 4:
            order.Status = ORDER_CANCEL_ING
        }

        switch orderMap["type"].(string) {
        case "buy":
            order.Side = BUY
        case "sell":
            order.Side = SELL
        case "buy_market":
            order.Side = BUY_MARKET
        case "sell_market":
            order.Side = SELL_MARKET
        }

        orderAr = append(orderAr, order)
    }

    //fmt.Println(orders);
    return orderAr, nil
}

func (o *OkCNApi) GetOneOrder(orderId string, currency CurrencyPair) (*Order, error) {
    orderAr, err := o.getOrders(orderId, currency)
    if err != nil {
        return nil, err
    }

    if len(orderAr) == 0 {
        return nil, nil
    }

    return &orderAr[0], nil
}

func (o *OkCNApi) GetUnfinishOrders(currency CurrencyPair) ([]Order, error) {
    return o.getOrders("-1", currency)
}

func (o *OkCNApi) GetAccount() (*Account, error) {
    postData := url.Values{}
    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    body, err := HttpPostForm(o.client, o.api_base_url+URL_USERINFO, postData)
    if err != nil {
        return nil, err
    }

    var respMap map[string]interface{}

    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return nil, err
    }

    if !respMap["result"].(bool) {
        errcode := strconv.FormatFloat(respMap["error_code"].(float64), 'f', 0, 64)
        return nil, errors.New(errcode)
    }

    info, ok := respMap["info"].(map[string]interface{})
    if !ok {
        return nil, errors.New(string(body))
    }

    funds := info["funds"].(map[string]interface{})
    asset := funds["asset"].(map[string]interface{})
    free := funds["free"].(map[string]interface{})
    freezed := funds["freezed"].(map[string]interface{})

    account := new(Account)
    account.Exchange = o.GetExchangeName()
    account.Asset, _ = strconv.ParseFloat(asset["total"].(string), 64)
    account.NetAsset, _ = strconv.ParseFloat(asset["net"].(string), 64)

    var btcSubAccount SubAccount
    var ltcSubAccount SubAccount
    var cnySubAccount SubAccount
    var ethSubAccount SubAccount
    var etcSubAccount SubAccount
    var bccSubAccount SubAccount

    btcSubAccount.Currency = BTC
    btcSubAccount.Amount, _ = strconv.ParseFloat(free["btc"].(string), 64)
    btcSubAccount.LoanAmount = 0
    btcSubAccount.ForzenAmount, _ = strconv.ParseFloat(freezed["btc"].(string), 64)

    ltcSubAccount.Currency = LTC
    ltcSubAccount.Amount, _ = strconv.ParseFloat(free["ltc"].(string), 64)
    ltcSubAccount.LoanAmount = 0
    ltcSubAccount.ForzenAmount, _ = strconv.ParseFloat(freezed["ltc"].(string), 64)

    ethSubAccount.Currency = ETH
    ethSubAccount.Amount, _ = strconv.ParseFloat(free["eth"].(string), 64)
    ethSubAccount.LoanAmount = 0
    ethSubAccount.ForzenAmount, _ = strconv.ParseFloat(freezed["eth"].(string), 64)

    etcSubAccount.Currency = ETC
    etcSubAccount.Amount = ToFloat64(free["etc"])
    etcSubAccount.LoanAmount = 0
    etcSubAccount.ForzenAmount = ToFloat64(freezed["etc"])

    bccSubAccount.Currency = BCC
    bccSubAccount.Amount = ToFloat64(free["bcc"])
    bccSubAccount.LoanAmount = 0
    bccSubAccount.ForzenAmount = ToFloat64(freezed["bcc"])

    cnySubAccount.Currency = CNY
    cnySubAccount.Amount, _ = strconv.ParseFloat(free["cny"].(string), 64)
    cnySubAccount.LoanAmount = 0
    cnySubAccount.ForzenAmount, _ = strconv.ParseFloat(freezed["cny"].(string), 64)

    account.SubAccounts = make(map[Currency]SubAccount, 3)
    account.SubAccounts[BTC] = btcSubAccount
    account.SubAccounts[LTC] = ltcSubAccount
    account.SubAccounts[CNY] = cnySubAccount
    account.SubAccounts[ETH] = ethSubAccount
    account.SubAccounts[ETC] = etcSubAccount
    account.SubAccounts[BCC] = bccSubAccount

    return account, nil
}

func (o *OkCNApi) GetTicker(currency CurrencyPair) (*Ticker, error) {
    var tickerMap map[string]interface{}
    var ticker Ticker

    url := o.api_base_url + URL_TICKER + "?symbol=" + currencyPair2String(currency)
    bodyDataMap, err := HttpGet(o.client, url)
    if err != nil {
        return nil, err
    }

    tickerMap = bodyDataMap["ticker"].(map[string]interface{})
    ticker.Date, _ = strconv.ParseUint(bodyDataMap["date"].(string), 10, 64)
    ticker.Last, _ = strconv.ParseFloat(tickerMap["last"].(string), 64)
    ticker.Buy, _ = strconv.ParseFloat(tickerMap["buy"].(string), 64)
    ticker.Sell, _ = strconv.ParseFloat(tickerMap["sell"].(string), 64)
    ticker.Low, _ = strconv.ParseFloat(tickerMap["low"].(string), 64)
    ticker.High, _ = strconv.ParseFloat(tickerMap["high"].(string), 64)
    ticker.Vol, _ = strconv.ParseFloat(tickerMap["vol"].(string), 64)

    return &ticker, nil
}

func (o *OkCNApi) GetExchangeName() string {
    return EXCHANGE_NAME_CN
}

func (o *OkCNApi) GetKlineRecords(currency CurrencyPair, period string, size, since int) ([]Kline, error) {
    klineUrl := o.api_base_url + fmt.Sprintf(URL_KLINE, currencyPair2String(currency), period, size, since)

    resp, err := http.Get(klineUrl)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    var klines [][]interface{}

    err = json.Unmarshal(body, &klines)
    if err != nil {
        return nil, err
    }

    var klineRecords []Kline

    for _, record := range klines {
        r := Kline{}
        for i, e := range record {
            switch i {
            case 0:
                r.Timestamp = int64(e.(float64)) / 1000 //to unix timestramp
            case 1:
                r.Open = e.(float64)
            case 2:
                r.High = e.(float64)
            case 3:
                r.Low = e.(float64)
            case 4:
                r.Close = e.(float64)
            case 5:
                r.Vol = e.(float64)
            }
        }
        klineRecords = append(klineRecords, r)
    }

    return klineRecords, nil
}

func (o *OkCNApi) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
    orderHistoryUrl := o.api_base_url + ORDER_HISTORY_URI

    postData := url.Values{}
    postData.Set("status", "1")
    postData.Set("symbol", CurrencyPairSymbol[currency])
    postData.Set("current_page", fmt.Sprintf("%d", currentPage))
    postData.Set("page_length", fmt.Sprintf("%d", pageSize))

    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    body, err := HttpPostForm(o.client, orderHistoryUrl, postData)
    if err != nil {
        return nil, err
    }

    var respMap map[string]interface{}

    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return nil, err
    }

    if !respMap["result"].(bool) {
        return nil, errors.New(string(body))
    }

    orders := respMap["orders"].([]interface{})

    var orderAr []Order
    for _, v := range orders {
        orderMap := v.(map[string]interface{})

        var order Order
        order.Currency = currency
        order.OrderID = int(orderMap["order_id"].(float64))
        order.Amount = orderMap["amount"].(float64)
        order.Price = orderMap["price"].(float64)
        order.DealAmount = orderMap["deal_amount"].(float64)
        order.AvgPrice = orderMap["avg_price"].(float64)
        order.OrderTime = int(orderMap["create_date"].(float64))

        //status:-1:已撤销  0:未成交  1:部分成交  2:完全成交 4:撤单处理中
        switch int(orderMap["status"].(float64)) {
        case -1:
            order.Status = ORDER_CANCEL
        case 0:
            order.Status = ORDER_UNFINISH
        case 1:
            order.Status = ORDER_PART_FINISH
        case 2:
            order.Status = ORDER_FINISH
        case 4:
            order.Status = ORDER_CANCEL_ING
        }

        switch orderMap["type"].(string) {
        case "buy":
            order.Side = BUY
        case "sell":
            order.Side = SELL
        }

        orderAr = append(orderAr, order)
    }

    return orderAr, nil
}

func (o *OkCNApi) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
    tradeUrl := ok.api_base_url + TRADE_URI
    postData := url.Values{}
    postData.Set("symbol", CurrencyPairSymbol[currencyPair])
    postData.Set("since", fmt.Sprintf("%d", since))

    err := ok.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    body, err := HttpPostForm(ok.client, tradeUrl, postData)
    if err != nil {
        return nil, err
    }
    //println(string(body))

    var trades []Trade
    err = json.Unmarshal(body, &trades)
    if err != nil {
        return nil, err
    }

    return trades, nil
}

func (o *OkCNApi) Withdraw(amount string, currency CurrencyPair, fees, receiveAddr, safePwd string) (string, error) {
    tradeUrl := ok.api_base_url + WITHDRAW
    postData := url.Values{}
    postData.Set("symbol", strings.ToLower(currency.String()))
    postData.Set("withdraw_amount", amount);
    postData.Set("chargefee", fees);
    postData.Set("withdraw_address", receiveAddr);
    postData.Set("trade_pwd", safePwd);
    err := ok.buildPostForm(&postData)
    if err != nil {
        return "", err
    }

    body, err := HttpPostForm(ok.client, tradeUrl, postData)
    if err != nil {
        fmt.Println("WITHDRAW fail.", err)
        return "", err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(body, &respMap)
    if err != nil {
        fmt.Println(err, string(body))
        return "", err
    }

    if respMap["result"].(bool) {
        return fmt.Sprintf("%.6f", respMap["withdraw_id"].(float64)), nil;
    }

    return "", errors.New(string(body))
}

func (o *OkCNApi) buildPostForm(postForm *url.Values) error {
    postForm.Set("api_key", o.api_key)
    payload := postForm.Encode()
    payload = payload + "&secret_key=" + o.secret_key
    sign, err := GetParamMD5Sign(o.secret_key, payload)
    if err != nil {
        return err
    }
    postForm.Set("sign", strings.ToUpper(sign))
    return nil
}

func (o *OkCNApi) placeOrder(side, amount, price string, currency CurrencyPair) (*Order, error) {
    postData := url.Values{}
    postData.Set("type", side)

    if side != "buy_market" {
        postData.Set("amount", amount)
    }
    if side != "sell_market" {
        postData.Set("price", price)
    }
    postData.Set("symbol", currencyPair2String(currency))

    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    body, err := HttpPostForm(o.client, o.api_base_url+URL_TRADE, postData)
    if err != nil {
        return nil, err
    }

    //println(string(body));

    var respMap map[string]interface{}

    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return nil, err
    }

    if !respMap["result"].(bool) {
        return nil, errors.New(string(body))
    }

    order := new(Order)
    order.OrderID = int(respMap["order_id"].(float64))
    order.Price, _ = strconv.ParseFloat(price, 64)
    order.Amount, _ = strconv.ParseFloat(amount, 64)
    order.Currency = currency
    order.Status = ORDER_UNFINISH

    switch side {
    case "buy":
        order.Side = BUY
    case "sell":
        order.Side = SELL
    }

    return order, nil
}