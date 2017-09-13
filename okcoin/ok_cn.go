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
    client    *http.Client
    apiKey    string
    secretKey string
    baseUrl   string
}

func NewOkCNApi(client *http.Client, apiKey, secretKey string) *OkCNApi {
    return &OkCNApi{client, apiKey, secretKey, URL_BASE}
}

func (o *OkCNApi) GetDepth(cp CurrencyPairV2, size int) (*Depth, error) {
    var depth Depth
    url := o.baseUrl + URL_DEPTH + "?symbol=" + cp.CustomSymbol("_", true) + "&size=" + strconv.Itoa(size)
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

func (o *OkCNApi) LimitBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return o.placeOrder(BUY, amount, price, cp)
}

func (o *OkCNApi) LimitSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return o.placeOrder(SELL, amount, price, cp)
}

func (o *OkCNApi) MarketBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return o.placeOrder(BUY_MARKET, amount, price, cp)
}

func (o *OkCNApi) MarketSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    return o.placeOrder(SELL_MARKET, amount, price, cp)
}

func (o *OkCNApi) CancelOrder(orderId string, cp CurrencyPairV2) (bool, error) {
    postData := url.Values{}
    postData.Set("order_id", orderId)
    postData.Set("symbol", cp.CustomSymbol("_", true))
    o.buildPostForm(&postData)

    body, err := HttpPostForm(o.client, o.baseUrl+URL_CANCEL_ORDER, postData)
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

func (o *OkCNApi) GetOneOrder(orderId string, cp CurrencyPairV2) (*OrderV2, error) {
    orderAr, err := o.getOrders(orderId, cp)
    if err != nil {
        return nil, err
    }
    if len(orderAr) == 0 {
        return nil, nil
    }
    return &orderAr[0], nil
}

func (o *OkCNApi) GetUnfinishedOrders(cp CurrencyPairV2) ([]OrderV2, error) {
    return o.getOrders("-1", cp)
}

func (o *OkCNApi) GetAccount() (*AccountV2, error) {
    postData := url.Values{}
    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }
    body, err := HttpPostForm(o.client, o.baseUrl+URL_USERINFO, postData)
    if err != nil {
        return nil, err
    }

    var respMap map[string]interface{}
    err = json.Unmarshal(body, &respMap)
    if err != nil {
        return nil, err
    }
    if !respMap["result"].(bool) {
        errCode := strconv.FormatFloat(respMap["error_code"].(float64), 'f', 0, 64)
        return nil, errors.New(errCode)
    }

    info, ok := respMap["info"].(map[string]interface{})
    if !ok {
        return nil, errors.New(string(body))
    }

    funds := info["funds"].(map[string]interface{})
    asset := funds["asset"].(map[string]interface{})
    free := funds["free"].(map[string]interface{})
    freezed := funds["freezed"].(map[string]interface{})

    account := new(AccountV2)
    account.Exchange = o.GetExchangeName()
    account.Asset, _ = strconv.ParseFloat(asset["total"].(string), 64)
    account.NetAsset, _ = strconv.ParseFloat(asset["net"].(string), 64)

    var btcSubAccount SubAccountV2
    var ltcSubAccount SubAccountV2
    var cnySubAccount SubAccountV2
    var ethSubAccount SubAccountV2
    var etcSubAccount SubAccountV2
    var bccSubAccount SubAccountV2

    btcSubAccount.Currency = "BTC"
    btcSubAccount.Amount, _ = strconv.ParseFloat(free["btc"].(string), 64)
    btcSubAccount.LoanAmount = 0
    btcSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["btc"].(string), 64)

    ltcSubAccount.Currency = "LTC"
    ltcSubAccount.Amount, _ = strconv.ParseFloat(free["ltc"].(string), 64)
    ltcSubAccount.LoanAmount = 0
    ltcSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["ltc"].(string), 64)

    ethSubAccount.Currency = "ETH"
    ethSubAccount.Amount, _ = strconv.ParseFloat(free["eth"].(string), 64)
    ethSubAccount.LoanAmount = 0
    ethSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["eth"].(string), 64)

    etcSubAccount.Currency = "ETC"
    etcSubAccount.Amount = ToFloat64(free["etc"])
    etcSubAccount.LoanAmount = 0
    etcSubAccount.FrozenAmount = ToFloat64(freezed["etc"])

    bccSubAccount.Currency = "BCC"
    bccSubAccount.Amount = ToFloat64(free["bcc"])
    bccSubAccount.LoanAmount = 0
    bccSubAccount.FrozenAmount = ToFloat64(freezed["bcc"])

    cnySubAccount.Currency = "CNY"
    cnySubAccount.Amount, _ = strconv.ParseFloat(free["cny"].(string), 64)
    cnySubAccount.LoanAmount = 0
    cnySubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["cny"].(string), 64)

    account.SubAccountsV2 = make(map[string]SubAccountV2, 6)
    account.SubAccountsV2["BTC"] = btcSubAccount
    account.SubAccountsV2["LTC"] = ltcSubAccount
    account.SubAccountsV2["ETH"] = ethSubAccount
    account.SubAccountsV2["ETC"] = etcSubAccount
    account.SubAccountsV2["BCC"] = bccSubAccount
    account.SubAccountsV2["CNY"] = cnySubAccount
    return account, nil
}

func (o *OkCNApi) GetTicker(cp CurrencyPairV2) (*Ticker, error) {
    url := o.baseUrl + URL_TICKER + "?symbol=" + cp.CustomSymbol("_", true)
    bodyDataMap, err := HttpGet(o.client, url)
    if err != nil {
        return nil, err
    }

    var tickerMap map[string]interface{}
    var ticker Ticker
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

func (o *OkCNApi) Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error) {
    tradeUrl := o.baseUrl + WITHDRAW
    postData := url.Values{}
    postData.Set("symbol", strings.ToLower(currency))
    postData.Set("withdraw_amount", amount);
    postData.Set("chargefee", fees);
    postData.Set("withdraw_address", receiveAddr);
    postData.Set("trade_pwd", safePwd);
    err := o.buildPostForm(&postData)
    if err != nil {
        return "", err
    }
    body, err := HttpPostForm(o.client, tradeUrl, postData)
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

func (o *OkCNApi) GetExchangeName() string {
    return EXCHANGE_NAME_CN
}

func (o *OkCNApi) GetKlineRecords(cp CurrencyPairV2, period string, size, since int) ([]Kline, error) {
    klineUrl := o.baseUrl + fmt.Sprintf(URL_KLINE, cp.CustomSymbol("_", true), period, size, since)
    resp, err := http.Get(klineUrl)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    var kLines [][]interface{}
    err = json.Unmarshal(body, &kLines)
    if err != nil {
        return nil, err
    }

    var klineRecords []Kline
    for _, record := range kLines {
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

func (o *OkCNApi) GetOrderHistory(cp CurrencyPairV2, currentPage, pageSize int) ([]OrderV2, error) {
    orderHistoryUrl := o.baseUrl + ORDER_HISTORY_URI
    postData := url.Values{}
    postData.Set("status", "1")
    postData.Set("symbol", cp.CustomSymbol("_", true))
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
    var orderAr []OrderV2
    for _, v := range orders {
        orderMap := v.(map[string]interface{})
        var order OrderV2
        order.CurrencyPair = cp.CustomSymbol("_", true)
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
            order.Status = ORDER_UNFINISHED
        case 1:
            order.Status = ORDER_PART_FINISH
        case 2:
            order.Status = ORDER_FINISH
        case 4:
            order.Status = ORDER_CANCELING
        }
        order.Side = StringToTradeSide(strings.ToUpper(orderMap["type"].(string)))
        orderAr = append(orderAr, order)
    }
    return orderAr, nil
}

func (o *OkCNApi) GetTrades(cp CurrencyPairV2, since int64) ([]Trade, error) {
    tradeUrl := o.baseUrl + TRADE_URI
    postData := url.Values{}
    postData.Set("symbol", cp.CustomSymbol("_", true))
    postData.Set("since", fmt.Sprintf("%d", since))
    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }
    body, err := HttpPostForm(o.client, tradeUrl, postData)
    if err != nil {
        return nil, err
    }

    var trades []Trade
    err = json.Unmarshal(body, &trades)
    if err != nil {
        return nil, err
    }
    return trades, nil
}

func (o *OkCNApi) getOrders(orderId string, cp CurrencyPairV2) ([]OrderV2, error) {
    postData := url.Values{}
    postData.Set("order_id", orderId)
    postData.Set("symbol", cp.CustomSymbol("_", true))
    o.buildPostForm(&postData)

    body, err := HttpPostForm(o.client, o.baseUrl+URL_ORDER_INFO, postData)
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
    var orderAr []OrderV2
    for _, v := range orders {
        orderMap := v.(map[string]interface{})
        var order OrderV2
        order.CurrencyPair = cp.CustomSymbol("_", true)
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
            order.Status = ORDER_UNFINISHED
        case 1:
            order.Status = ORDER_PART_FINISH
        case 2:
            order.Status = ORDER_FINISH
        case 4:
            order.Status = ORDER_CANCELING
        }
        order.Side = StringToTradeSide(strings.ToUpper(orderMap["type"].(string)))
        orderAr = append(orderAr, order)
    }
    return orderAr, nil
}

func (o *OkCNApi) buildPostForm(postForm *url.Values) error {
    postForm.Set("apiKey", o.apiKey)
    payload := postForm.Encode()
    payload = payload + "&secretKey=" + o.secretKey
    sign, err := GetParamMD5Sign(o.secretKey, payload)
    if err != nil {
        return err
    }
    postForm.Set("sign", strings.ToUpper(sign))
    return nil
}

func (o *OkCNApi) placeOrder(side TradeSide, amount, price string, cp CurrencyPairV2) (*OrderV2, error) {
    postData := url.Values{}
    postData.Set("type", strings.ToLower(side.String()))
    postData.Set("symbol", cp.CustomSymbol("_", true))
    if side != BUY_MARKET {
        postData.Set("amount", amount)
    }
    if side != SELL_MARKET {
        postData.Set("price", price)
    }
    err := o.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    body, err := HttpPostForm(o.client, o.baseUrl+URL_TRADE, postData)
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

    order := new(OrderV2)
    order.OrderID = int(respMap["order_id"].(float64))
    order.Price, _ = strconv.ParseFloat(price, 64)
    order.Amount, _ = strconv.ParseFloat(amount, 64)
    order.CurrencyPair = cp.CustomSymbol("_", true)
    order.Status = ORDER_UNFINISHED
    order.Side = side
    return order, nil
}
