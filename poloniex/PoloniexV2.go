package poloniex

import (
    "encoding/json"
    "errors"
    "fmt"
    . "github.com/qct/crypto_coin_api"
    "log"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"
)

type PoloniexV2 struct {
    accessKey,
    secretKey string
    client *http.Client
}

func NewPoloniexV2(client *http.Client, accessKey, secretKey string) *PoloniexV2 {
    return &PoloniexV2{accessKey, secretKey, client}
}

func (poloniex *PoloniexV2) GetExchangeName() string {
    return EXCHANGE_NAME
}

func (poloniex *PoloniexV2) GetTicker(baseCurrency, counterCurrency string) (*Ticker, error) {
    respmap, err := HttpGet(poloniex.client, PUBLIC_URL+TICKER_API)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    tickermap := respmap[combineCurrencyPair(baseCurrency, counterCurrency)].(map[string]interface{})

    ticker := new(Ticker)
    ticker.High, _ = strconv.ParseFloat(tickermap["high24hr"].(string), 64)
    ticker.Low, _ = strconv.ParseFloat(tickermap["low24hr"].(string), 64)
    ticker.Last, _ = strconv.ParseFloat(tickermap["last"].(string), 64)
    ticker.Buy, _ = strconv.ParseFloat(tickermap["highestBid"].(string), 64)
    ticker.Sell, _ = strconv.ParseFloat(tickermap["lowestAsk"].(string), 64)
    ticker.Vol, _ = strconv.ParseFloat(tickermap["quoteVolume"].(string), 64)

    log.Println(tickermap)

    return ticker, nil
}

func (poloniex *PoloniexV2) GetDepth(size int, baseCurrency, counterCurrency string) (*Depth, error) {
    respmap, err := HttpGet(poloniex.client, PUBLIC_URL+fmt.Sprintf(ORDER_BOOK_API, combineCurrencyPair(baseCurrency, counterCurrency), size))
    if err != nil {
        log.Println(err)
        return nil, err
    }

    if respmap["asks"] == nil {
        log.Println(respmap)
        return nil, errors.New(fmt.Sprintf("%+v", respmap))
    }

    _, isOK := respmap["asks"].([]interface{})
    if !isOK {
        log.Println(respmap)
        return nil, errors.New(fmt.Sprintf("%+v", respmap))
    }

    var depth Depth

    for _, v := range respmap["asks"].([]interface{}) {
        var dr DepthRecord
        for i, vv := range v.([]interface{}) {
            switch i {
            case 0:
                dr.Price, _ = strconv.ParseFloat(vv.(string), 64)
            case 1:
                dr.Amount = vv.(float64)
            }
        }
        depth.AskList = append(depth.AskList, dr)
    }

    for _, v := range respmap["bids"].([]interface{}) {
        var dr DepthRecord
        for i, vv := range v.([]interface{}) {
            switch i {
            case 0:
                dr.Price, _ = strconv.ParseFloat(vv.(string), 64)
            case 1:
                dr.Amount = vv.(float64)
            }
        }
        depth.BidList = append(depth.BidList, dr)
    }

    return &depth, nil
}

func (Poloniex *PoloniexV2) GetKlineRecords(baseCurrency, counterCurrency, period string, size, since int) ([]Kline, error) {
    return nil, nil
}

func (poloniex *PoloniexV2) placeLimitOrder(command, amount, price string, baseCurrency, counterCurrency string) (*Order, error) {
    postData := url.Values{}
    postData.Set("command", command)
    postData.Set("currencyPair", combineCurrencyPair(baseCurrency, counterCurrency))
    postData.Set("rate", price)
    postData.Set("amount", amount)

    sign, _ := poloniex.buildPostForm(&postData)

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(poloniex.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    respmap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respmap)
    if err != nil || respmap["error"] != nil {
        log.Println(err, string(resp))
        return nil, err
    }

    orderNumber := respmap["orderNumber"].(string)
    order := new(Order)
    order.OrderTime = int(time.Now().Unix() * 1000)
    order.OrderID, _ = strconv.Atoi(orderNumber)
    order.Amount, _ = strconv.ParseFloat(amount, 64)
    order.Price, _ = strconv.ParseFloat(price, 64)
    order.Status = ORDER_UNFINISH
    order.CurrencyPair = combineCurrencyPair(baseCurrency, counterCurrency)

    switch command {
    case "sell":
        order.Side = SELL
    case "buy":
        order.Side = BUY
    }

    log.Println(string(resp))
    return order, nil
}

func (poloniex *PoloniexV2) LimitBuy(amount, price string, baseCurrency, counterCurrency string) (*Order, error) {
    return poloniex.placeLimitOrder("buy", amount, price, baseCurrency, counterCurrency)
}

func (poloniex *PoloniexV2) LimitSell(amount, price string, baseCurrency, counterCurrency string) (*Order, error) {
    return poloniex.placeLimitOrder("sell", amount, price, baseCurrency, counterCurrency)
}

func (poloniex *PoloniexV2) CancelOrder(orderId string) (bool, error) {
    postData := url.Values{}
    postData.Set("command", "cancelOrder")
    postData.Set("orderNumber", orderId)

    sign, err := poloniex.buildPostForm(&postData)
    if err != nil {
        log.Println(err)
        return false, err
    }

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(poloniex.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return false, err
    }

    //log.Println(string(resp));

    respmap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respmap)
    if err != nil || respmap["error"] != nil {
        log.Println(err, string(resp))
        return false, err
    }

    success := int(respmap["success"].(float64))
    if success != 1 {
        log.Println(respmap)
        return false, nil
    }

    return true, nil
}

func (poloniex *PoloniexV2) GetOneOrder(orderId string, baseCurrency, counterCurrency string) (*Order, error) {
    postData := url.Values{}
    postData.Set("command", "returnOrderTrades")
    postData.Set("orderNumber", orderId)

    sign, _ := poloniex.buildPostForm(&postData)

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(poloniex.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    //println(string(resp))
    if strings.Contains(string(resp), "error") {
        ords, err1 := poloniex.GetUnfinishOrders(baseCurrency, counterCurrency)
        if err1 != nil {
            log.Println(err1)
        } else {
            _ordId, _ := strconv.Atoi(orderId)

            for _, ord := range ords {
                if ord.OrderID == _ordId {
                    return &ord, nil
                }
            }
        }
        //log.Println(string(resp))
        return nil, errors.New(string(resp))
    }

    respmap := make([]interface{}, 0)
    err = json.Unmarshal(resp, &respmap)
    if err != nil {
        log.Println(err, string(resp))
        return nil, err
    }

    order := new(Order)
    order.OrderID, _ = strconv.Atoi(orderId)
    order.CurrencyPair = combineCurrencyPair(baseCurrency, counterCurrency)

    total := 0.0

    for _, v := range respmap {
        vv := v.(map[string]interface{})
        _amount, _ := strconv.ParseFloat(vv["amount"].(string), 64)
        _rate, _ := strconv.ParseFloat(vv["rate"].(string), 64)
        _fee, _ := strconv.ParseFloat(vv["fee"].(string), 64)

        order.DealAmount += _amount
        total += (_amount * _rate)
        order.Fee = _fee

        if strings.Compare("sell", vv["type"].(string)) == 0 {
            order.Side = TradeSide(SELL)
        } else {
            order.Side = TradeSide(BUY)
        }
    }

    order.AvgPrice = total / order.DealAmount

    return order, nil
}

func (poloniex *PoloniexV2) GetUnfinishOrders(baseCurrency, counterCurrency string) ([]Order, error) {
    postData := url.Values{}
    postData.Set("command", "returnOpenOrders")
    postData.Set("currencyPair", combineCurrencyPair(baseCurrency, counterCurrency))

    sign, err := poloniex.buildPostForm(&postData)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(poloniex.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    orderAr := make([]interface{}, 1)
    err = json.Unmarshal(resp, &orderAr)
    if err != nil {
        log.Println(err, string(resp))
        return nil, err
    }

    orders := make([]Order, 0)
    for _, v := range orderAr {
        vv := v.(map[string]interface{})
        order := Order{}
        order.CurrencyPair = combineCurrencyPair(baseCurrency, counterCurrency)
        order.OrderID, _ = strconv.Atoi(vv["orderNumber"].(string))
        order.Amount, _ = strconv.ParseFloat(vv["amount"].(string), 64)
        order.Price, _ = strconv.ParseFloat(vv["rate"].(string), 64)
        order.Status = ORDER_UNFINISH

        side := vv["type"].(string)
        switch side {
        case "buy":
            order.Side = TradeSide(BUY)
        case "sell":
            order.Side = TradeSide(SELL)
        }

        orders = append(orders, order)
    }

    //log.Println(orders)
    return orders, nil
}

func (Poloniex *PoloniexV2) GetOrderHistorys(baseCurrency, counterCurrency string, currentPage, pageSize int) ([]Order, error) {
    return nil, nil
}

func (poloniex *PoloniexV2) GetAccount() (*Account, error) {
    postData := url.Values{}
    postData.Add("command", "returnCompleteBalances")
    sign, err := poloniex.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(poloniex.client, TRADE_API, postData, headers)

    if err != nil {
        log.Println(err)
        return nil, err
    }

    respmap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respmap)

    if err != nil || respmap["error"] != nil {
        log.Println(err)
        return nil, err
    }

    acc := new(Account)
    acc.Exchange = EXCHANGE_NAME
    acc.SubAccountsV2 = make(map[string]SubAccount)

    for k, v := range respmap {
        vv := v.(map[string]interface{})
        subAcc := SubAccount{}
        subAcc.CurrencyStr = k
        subAcc.Amount, _ = strconv.ParseFloat(vv["available"].(string), 64)
        subAcc.ForzenAmount, _ = strconv.ParseFloat(vv["onOrders"].(string), 64)
        acc.SubAccountsV2[subAcc.CurrencyStr] = subAcc

        //var currency Currency
        //
        //switch k {
        //case "BTC":
        //    currency = BTC
        //case "LTC":
        //    currency = LTC
        //case "ETH":
        //    currency = ETH
        //case "ETC":
        //    currency = ETC
        //case "USD":
        //    currency = USD
        //case "BTS":
        //    currency = BTS
        //default:
        //    currency = -1
        //}
        //
        //if currency > 0 {
        //    vv := v.(map[string]interface{})
        //    subAcc := SubAccount{}
        //    subAcc.Currency = currency
        //    subAcc.Amount, _ = strconv.ParseFloat(vv["available"].(string), 64)
        //    subAcc.ForzenAmount, _ = strconv.ParseFloat(vv["onOrders"].(string), 64)
        //    acc.SubAccounts[subAcc.Currency] = subAcc
        //}
    }

    return acc, nil
}

func (p *PoloniexV2) Withdraw(amount string, currency string, fees, receiveAddr, safePwd string) (string, error) {
    params := url.Values{}
    params.Add("command", "withdraw")
    params.Add("address", receiveAddr)
    params.Add("amount", amount)
    params.Add("currency", strings.ToUpper(currency));
    sign, err := p.buildPostForm(&params)
    if err != nil {
        return "", err
    }

    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(p.client, TRADE_API, params, headers)

    if err != nil {
        log.Println(err)
        return "", err
    }
    println(string(resp))

    respMap := make(map[string]interface{})

    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err)
        return "", err
    }

    if respMap["error"] == nil {
        return string(resp), nil
    }

    return "", errors.New(string(resp))
}

func (p *PoloniexV2) WithdrawWithMemo(amount string, currency string, paymentId, receiveAddr, safePwd string) (string, error) {
    params := url.Values{}
    params.Add("command", "withdraw")
    params.Add("address", receiveAddr)
    params.Add("amount", amount)
    params.Add("currency", strings.ToUpper(currency));
    params.Add("paymentId", paymentId);
    sign, err := p.buildPostForm(&params)
    if err != nil {
        return "", err
    }

    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(p.client, TRADE_API, params, headers)

    if err != nil {
        log.Println(err)
        return "", err
    }
    println(string(resp))

    respMap := make(map[string]interface{})

    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err)
        return "", err
    }

    if respMap["error"] == nil {
        return string(resp), nil
    }

    return "", errors.New(string(resp))
}

func (poloniex *PoloniexV2) GetDepositsWithdrawals(start, end string) (*PoloniexDepositsWithdrawals, error) {
    params := url.Values{}
    params.Set("command", "returnDepositsWithdrawals")
    println(start)
    if start != "" {
        params.Set("start", start)
    } else {
        params.Set("start", "0")
    }

    if end != "" {
        params.Set("end", end)
    } else {
        params.Set("end", strconv.FormatInt(time.Now().Unix(), 10))
    }

    sign, err := poloniex.buildPostForm(&params)
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Key":  poloniex.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(poloniex.client, TRADE_API, params, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    println(string(resp))

    records := new(PoloniexDepositsWithdrawals)
    err = json.Unmarshal(resp, records)

    return records, err
}

func (poloniex *PoloniexV2) buildPostForm(postForm *url.Values) (string, error) {
    postForm.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()+500000000000))
    payload := postForm.Encode()
    //println(payload)
    sign, err := GetParamHmacSHA512Sign(poloniex.secretKey, payload)
    if err != nil {
        return "", err
    }
    //log.Println(sign)
    return sign, nil
}

func (poloniex *PoloniexV2) GetTrades(baseCurrency, counterCurrency string, since int64) ([]Trade, error) {
    panic("unimplements")
}

func (poloniex *PoloniexV2) MarketBuy(amount, price string, baseCurrency, counterCurrency string) (*Order, error) {
    panic("unsupport the market order")
}

func (poloniex *PoloniexV2) MarketSell(amount, price string, baseCurrency, counterCurrency string) (*Order, error) {
    panic("unsupport the market order")
}

func (poloniex *PoloniexV2) GetCurrency(currency string) (*PoloniexCurrency, error) {
    respmap, err := HttpGet(poloniex.client, PUBLIC_URL+CURRENCIES_API)

    if err != nil || respmap["error"] != nil {
        log.Println(err)
        return nil, err
    }

    currencyMap := respmap[strings.ToUpper(currency)].(map[string]interface{})

    poloniexCurrency := new(PoloniexCurrency)
    poloniexCurrency.ID = int(currencyMap["id"].(float64))
    poloniexCurrency.Name, _ = currencyMap["name"].(string)
    poloniexCurrency.TxFee, _ = strconv.ParseFloat(currencyMap["txFee"].(string), 64)
    poloniexCurrency.MinConf = int(currencyMap["minConf"].(float64))
    poloniexCurrency.DepositAddress, _ = currencyMap["depositAddress"].(string)
    poloniexCurrency.Disabled = int(currencyMap["disabled"].(float64))
    poloniexCurrency.Delisted = int(currencyMap["delisted"].(float64))
    poloniexCurrency.Frozen = int(currencyMap["frozen"].(float64))

    return poloniexCurrency, nil
}

func (poloniex *PoloniexV2) GetAllCurrencies() (map[string]*PoloniexCurrency, error) {
    respmap, err := HttpGet(poloniex.client, PUBLIC_URL+CURRENCIES_API)

    if err != nil || respmap["error"] != nil {
        log.Println(err)
        return nil, err
    }

    result := map[string]*PoloniexCurrency{}
    for k, v := range respmap {
        currencyMap := v.(map[string]interface{})
        poloniexCurrency := new(PoloniexCurrency)
        poloniexCurrency.ID = int(currencyMap["id"].(float64))
        poloniexCurrency.Name, _ = currencyMap["name"].(string)
        poloniexCurrency.TxFee, _ = strconv.ParseFloat(currencyMap["txFee"].(string), 64)
        poloniexCurrency.MinConf = int(currencyMap["minConf"].(float64))
        poloniexCurrency.DepositAddress, _ = currencyMap["depositAddress"].(string)
        poloniexCurrency.Disabled = int(currencyMap["disabled"].(float64))
        poloniexCurrency.Delisted = int(currencyMap["delisted"].(float64))
        poloniexCurrency.Frozen = int(currencyMap["frozen"].(float64))

        result[k] = poloniexCurrency
    }
    return result, nil
}

func combineCurrencyPair(baseCurrency, counterCurrency string) string {
    if baseCurrency == "" {
        baseCurrency = "BTC"
    }
    return strings.ToUpper(baseCurrency) + "_" + strings.ToUpper(counterCurrency)
}