package poloniex

import (
    "errors"
    "fmt"
    . "github.com/qct/crypto_coin_api"
    "log"
    "net/http"
    "strconv"
    "strings"
    "net/url"
    "encoding/json"
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

func (p *PoloniexV2) LimitBuy(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error) {
    return p.placeLimitOrder("buy", amount, price, baseCurrency, counterCurrency)
}

func (p *PoloniexV2) LimitSell(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error) {
    return p.placeLimitOrder("sell", amount, price, baseCurrency, counterCurrency)
}

func (p *PoloniexV2) MarketBuy(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error) {
    return &OrderV2{}, nil
}

func (p *PoloniexV2) MarketSell(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error) {
    return &OrderV2{}, nil
}

func (p *PoloniexV2) CancelOrder(orderId, baseCurrency, counterCurrency string) (bool, error) {
    postData := url.Values{}
    postData.Set("command", "cancelOrder")
    postData.Set("orderNumber", orderId)

    sign, err := p.buildPostForm(&postData)
    if err != nil {
        log.Println(err)
        return false, err
    }

    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(p.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return false, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil || respMap["error"] != nil {
        log.Println(err, string(resp))
        return false, err
    }

    success := int(respMap["success"].(float64))
    if success != 1 {
        log.Println(respMap)
        return false, nil
    }
    return true, nil
}

func (p *PoloniexV2) GetOneOrder(orderId, baseCurrency, counterCurrency string) (*OrderV2, error) {
    postData := url.Values{}
    postData.Set("command", "returnOrderTrades")
    postData.Set("orderNumber", orderId)
    sign, _ := p.buildPostForm(&postData)
    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(p.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    if strings.Contains(string(resp), "error") {
        orders, err1 := p.GetUnfinishedOrders(baseCurrency, counterCurrency)
        if err1 != nil {
            log.Println(err1)
        } else {
            _ordId, _ := strconv.Atoi(orderId)
            for _, ord := range orders {
                if ord.OrderID == _ordId {
                    return &ord, nil
                }
            }
        }
        return nil, errors.New(string(resp))
    }

    respMap := make([]interface{}, 0)
    err = json.Unmarshal(resp, &respMap)
    if err != nil {
        log.Println(err, string(resp))
        return nil, err
    }

    order := new(OrderV2)
    order.OrderID, _ = strconv.Atoi(orderId)
    order.CurrencyPair = combineCurrencyPair(baseCurrency, counterCurrency)

    total := 0.0
    for _, v := range respMap {
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

func (p *PoloniexV2) GetUnfinishedOrders(baseCurrency, counterCurrency string) ([]OrderV2, error) {
    postData := url.Values{}
    postData.Set("command", "returnOpenOrders")
    postData.Set("currencyPair", combineCurrencyPair(baseCurrency, counterCurrency))

    sign, err := p.buildPostForm(&postData)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(p.client, TRADE_API, postData, headers)
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

    orders := make([]OrderV2, 0)
    for _, v := range orderAr {
        vv := v.(map[string]interface{})
        order := OrderV2{}
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
    return orders, nil
}

func (p *PoloniexV2) GetOrderHistory(baseCurrency, counterCurrency string, currentPage, pageSize int) ([]OrderV2, error) {
    return []OrderV2{}, nil
}

func (p *PoloniexV2) GetAccount() (*AccountV2, error) {
    postData := url.Values{}
    postData.Add("command", "returnCompleteBalances")
    sign, err := p.buildPostForm(&postData)
    if err != nil {
        return nil, err
    }
    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}
    resp, err := HttpPostForm2(p.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)

    if err != nil || respMap["error"] != nil {
        log.Println(err)
        return nil, err
    }

    acc := new(AccountV2)
    acc.Exchange = EXCHANGE_NAME
    acc.SubAccountsV2 = make(map[string]SubAccountV2)

    for k, v := range respMap {
        vv := v.(map[string]interface{})
        subAcc := SubAccountV2{}
        subAcc.Currency = k
        subAcc.Amount, _ = strconv.ParseFloat(vv["available"].(string), 64)
        subAcc.FrozenAmount, _ = strconv.ParseFloat(vv["onOrders"].(string), 64)
        acc.SubAccountsV2[subAcc.Currency] = subAcc
    }
    return acc, nil
}

func (p *PoloniexV2) GetTicker(baseCurrency, counterCurrency string) (*Ticker, error) {
    resp, err := HttpGet(p.client, PUBLIC_URL+TICKER_API)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    tickerMap := resp[combineCurrencyPair(baseCurrency, counterCurrency)].(map[string]interface{})
    ticker := new(Ticker)
    ticker.High, _ = strconv.ParseFloat(tickerMap["high24hr"].(string), 64)
    ticker.Low, _ = strconv.ParseFloat(tickerMap["low24hr"].(string), 64)
    ticker.Last, _ = strconv.ParseFloat(tickerMap["last"].(string), 64)
    ticker.Buy, _ = strconv.ParseFloat(tickerMap["highestBid"].(string), 64)
    ticker.Sell, _ = strconv.ParseFloat(tickerMap["lowestAsk"].(string), 64)
    ticker.Vol, _ = strconv.ParseFloat(tickerMap["quoteVolume"].(string), 64)
    return ticker, nil
}

func (p *PoloniexV2) GetDepth(size int, baseCurrency, counterCurrency string) (*Depth, error) {
    resp, err := HttpGet(p.client, PUBLIC_URL+fmt.Sprintf(ORDER_BOOK_API, combineCurrencyPair(baseCurrency, counterCurrency), size))
    if err != nil {
        log.Println(err)
        return nil, err
    }
    if resp["asks"] == nil {
        log.Println(resp)
        return nil, errors.New(fmt.Sprintf("%+v", resp))
    }
    if _, ok := resp["asks"].([]interface{}); !ok {
        log.Println(resp)
        return nil, errors.New(fmt.Sprintf("%+v", resp))
    }

    var depth Depth
    for _, v := range resp["asks"].([]interface{}) {
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
    for _, v := range resp["bids"].([]interface{}) {
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

func (p *PoloniexV2) GetKlineRecords(baseCurrency, counterCurrency, period string, size, since int) ([]Kline, error) {
    return nil, nil
}

func (p *PoloniexV2) GetTrades(baseCurrency, counterCurrency string, since int64) ([]Trade, error) {
    return []Trade{}, nil
}

func (p *PoloniexV2) GetExchangeName() string {
    return EXCHANGE_NAME
}

func (p *PoloniexV2) Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error) {
    params := url.Values{}
    params.Add("command", "withdraw")
    params.Add("address", receiveAddr)
    params.Add("amount", amount)
    params.Add("currency", strings.ToUpper(currency));
    if memo != "" {
        params.Add("paymentId", memo);
    }
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

//-------------------------
func (p *PoloniexV2) GetDepositsWithdrawals(start, end string) (*PoloniexDepositsWithdrawals, error) {
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

    sign, err := p.buildPostForm(&params)
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(p.client, TRADE_API, params, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    println(string(resp))

    records := new(PoloniexDepositsWithdrawals)
    err = json.Unmarshal(resp, records)

    return records, err
}

func (p *PoloniexV2) GetCurrency(currency string) (*PoloniexCurrency, error) {
    resp, err := HttpGet(p.client, PUBLIC_URL+CURRENCIES_API)

    if err != nil || resp["error"] != nil {
        log.Println(err)
        return nil, err
    }

    currencyMap := resp[strings.ToUpper(currency)].(map[string]interface{})

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

func (p *PoloniexV2) GetAllCurrencies() (map[string]*PoloniexCurrency, error) {
    respmap, err := HttpGet(p.client, PUBLIC_URL+CURRENCIES_API)

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
//-------------------------

func combineCurrencyPair(baseCurrency, counterCurrency string) string {
    if baseCurrency == "" {
        baseCurrency = "BTC"
    }
    return strings.ToUpper(baseCurrency) + "_" + strings.ToUpper(counterCurrency)
}

func (p *PoloniexV2) placeLimitOrder(command, amount, price string, baseCurrency, counterCurrency string) (*OrderV2, error) {
    postData := url.Values{}
    postData.Set("command", command)
    postData.Set("currencyPair", combineCurrencyPair(baseCurrency, counterCurrency))
    postData.Set("rate", price)
    postData.Set("amount", amount)
    sign, _ := p.buildPostForm(&postData)
    headers := map[string]string{
        "Key":  p.accessKey,
        "Sign": sign}

    resp, err := HttpPostForm2(p.client, TRADE_API, postData, headers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    respMap := make(map[string]interface{})
    err = json.Unmarshal(resp, &respMap)
    if err != nil || respMap["error"] != nil {
        log.Println(err, string(resp))
        return nil, err
    }

    orderNumber := respMap["orderNumber"].(string)
    order := new(OrderV2)
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

func (p *PoloniexV2) buildPostForm(postForm *url.Values) (string, error) {
    postForm.Add("nonce", fmt.Sprintf("%d", time.Now().UnixNano()+500000000000))
    payload := postForm.Encode()
    sign, err := GetParamHmacSHA512Sign(p.secretKey, payload)
    if err != nil {
        return "", err
    }
    return sign, nil
}