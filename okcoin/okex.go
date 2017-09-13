package okcoin

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/qct/crypto_coin_api"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	FUTURE_EXCHANGE_NAME   = "okex.com"
	FUTURE_API_BASE_URL    = "https://www.okex.com/api/v1/"
	FUTURE_TICKER_URI      = "future_ticker.do?symbol=%s&contract_type=%s"
	FUTURE_DEPTH_URI       = "future_depth.do?symbol=%s&contract_type=%s&size=%d"
	FUTURE_USERINFO_URI    = "future_userinfo.do"
	FUTURE_CANCEL_URI      = "future_cancel.do"
	FUTURE_ORDER_INFO_URI  = "future_order_info.do"
	FUTURE_ORDERS_INFO_URI = "future_orders_info.do"
	FUTURE_POSITION_URI    = "future_position.do"
	FUTURE_TRADE_URI       = "future_trade.do"
	FUTURE_ESTIMATED_PRICE = "future_estimated_price.do?symbol=%s"
	FUTURE_GET_KLINE_URI   = "future_kline.do"
	EXCHANGE_RATE_URI      = "exchange_rate.do"
)

type futureUserInfoResponse struct {
	Info struct {
		Btc map[string]float64 `json:btc`
		Ltc map[string]float64 `json:ltc`
	} `json:info`
	Result bool `json:"result,bool"`
}

type OkExApi struct {
	apiKey       string
	apiSecretKey string
	client       *http.Client
}

func NewOkExApi(client *http.Client, apiKey, secretKey string) *OkExApi {
	return &OkExApi{apiKey: apiKey, apiSecretKey: secretKey, client: client}
}

func (o *OkExApi) GetFutureEstimatedPrice(cp CurrencyPair) (float64, error) {
	resp, err := o.client.Get(fmt.Sprintf(FUTURE_API_BASE_URL+FUTURE_ESTIMATED_PRICE, cp.CustomSymbol("_", true)))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return 0, err
	}
	return bodyMap["forecast_price"].(float64), nil
}

func (o *OkExApi) GetFutureTicker(cp CurrencyPair, contractType string) (*Ticker, error) {
	url := FUTURE_API_BASE_URL + FUTURE_TICKER_URI
	resp, err := o.client.Get(fmt.Sprintf(url, cp.CustomSymbol("_", true), contractType))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, err
	}

	if bodyMap["result"] != nil && !bodyMap["result"].(bool) {
		return nil, errors.New(string(body))
	}
	tickerMap := bodyMap["ticker"].(map[string]interface{})
	ticker := new(Ticker)
	ticker.Date, _ = strconv.ParseUint(bodyMap["date"].(string), 10, 64)
	ticker.Buy = tickerMap["buy"].(float64)
	ticker.Sell = tickerMap["sell"].(float64)
	ticker.Last = tickerMap["last"].(float64)
	ticker.High = tickerMap["high"].(float64)
	ticker.Low = tickerMap["low"].(float64)
	ticker.Vol = tickerMap["vol"].(float64)
	return ticker, nil
}

func (o *OkExApi) GetFutureDepth(cp CurrencyPair, contractType string, size int) (*Depth, error) {
	url := FUTURE_API_BASE_URL + FUTURE_DEPTH_URI
	resp, err := o.client.Get(fmt.Sprintf(url, cp.CustomSymbol("_", true), contractType, size))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyMap := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, err
	}
	if bodyMap["error_code"] != nil {
		log.Println(bodyMap)
		return nil, errors.New(string(body))
	}

	depth := new(Depth)
	for _, v := range bodyMap["asks"].([]interface{}) {
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
	for _, v := range bodyMap["bids"].([]interface{}) {
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
	return depth, nil
}

func (o *OkExApi) GetFutureIndex(cp CurrencyPair) (float64, error) {
	return 0, nil
}

func (o *OkExApi) GetFutureUserInfo() (*FutureAccount, error) {
	userInfoUrl := FUTURE_API_BASE_URL + FUTURE_USERINFO_URI
	postData := url.Values{}
	o.buildPostForm(&postData)
	body, err := HttpPostForm(o.client, userInfoUrl, postData)
	if err != nil {
		return nil, err
	}
	resp := futureUserInfoResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(string(body))
	}

	account := new(FutureAccount)
	account.FutureSubAccounts = make(map[string]FutureSubAccount, 2)
	btcMap := resp.Info.Btc
	ltcMap := resp.Info.Ltc
	account.FutureSubAccounts["BTC"] = FutureSubAccount{Currency: "BTC",
		AccountRights: btcMap["account_rights"],
		KeepDeposit:   btcMap["keep_deposit"],
		ProfitReal:    btcMap["profit_real"],
		ProfitUnreal:  btcMap["profit_unreal"],
		RiskRate:      btcMap["risk_rate"],
	}
	account.FutureSubAccounts["LTC"] = FutureSubAccount{Currency: "LTC",
		AccountRights: ltcMap["account_rights"],
		KeepDeposit:   ltcMap["keep_deposit"],
		ProfitReal:    ltcMap["profit_real"],
		ProfitUnreal:  ltcMap["profit_unreal"],
		RiskRate:      ltcMap["risk_rate"],
	}
	return account, nil
}

func (o *OkExApi) PlaceFutureOrder(cp CurrencyPair, contractType, price, amount string, openType, matchPrice, leverRate int) (string, error) {
	postData := url.Values{}
	postData.Set("symbol", cp.CustomSymbol("_", true))
	postData.Set("price", price)
	postData.Set("contract_type", contractType)
	postData.Set("amount", amount)
	postData.Set("type", strconv.Itoa(openType))
	postData.Set("lever_rate", strconv.Itoa(leverRate))
	postData.Set("match_price", strconv.Itoa(matchPrice))
	o.buildPostForm(&postData)
	placeOrderUrl := FUTURE_API_BASE_URL + FUTURE_TRADE_URI
	body, err := HttpPostForm(o.client, placeOrderUrl, postData)
	if err != nil {
		return "", err
	}
	respMap := make(map[string]interface{})
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return "", err
	}
	if !respMap["result"].(bool) {
		return "", errors.New(string(body))
	}
	return fmt.Sprintf("%.0f", respMap["order_id"].(float64)), nil
}

func (o *OkExApi) FutureCancelOrder(cp CurrencyPair, contractType, orderId string) (bool, error) {
	postData := url.Values{}
	postData.Set("symbol", cp.CustomSymbol("_", true))
	postData.Set("order_id", orderId)
	postData.Set("contract_type", contractType)
	o.buildPostForm(&postData)
	cancelUrl := FUTURE_API_BASE_URL + FUTURE_CANCEL_URI
	body, err := HttpPostForm(o.client, cancelUrl, postData)
	if err != nil {
		return false, err
	}
	respMap := make(map[string]interface{})
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return false, err
	}
	if respMap["result"] != nil && !respMap["result"].(bool) {
		return false, errors.New(string(body))
	}
	return true, nil
}

func (o *OkExApi) GetFuturePosition(cp CurrencyPair, contractType string) ([]FuturePosition, error) {
	positionUrl := FUTURE_API_BASE_URL + FUTURE_POSITION_URI
	postData := url.Values{}
	postData.Set("contract_type", contractType)
	postData.Set("symbol", cp.CustomSymbol("_", true))
	o.buildPostForm(&postData)
	body, err := HttpPostForm(o.client, positionUrl, postData)
	if err != nil {
		return nil, err
	}
	respMap := make(map[string]interface{})
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return nil, err
	}
	if !respMap["result"].(bool) {
		return nil, errors.New(string(body))
	}

	var posAr []FuturePosition
	forceLiquPriceStr := respMap["force_liqu_price"].(string)
	forceLiquPriceStr = strings.Replace(forceLiquPriceStr, ",", "", 1)
	forceLiquPrice, err := strconv.ParseFloat(forceLiquPriceStr, 64)
	holdings := respMap["holding"].([]interface{})
	for _, v := range holdings {
		holdingMap := v.(map[string]interface{})
		pos := FuturePosition{}
		pos.ForceLiquPrice = forceLiquPrice
		pos.LeverRate = int(holdingMap["lever_rate"].(float64))
		pos.ContractType = holdingMap["contract_type"].(string)
		pos.ContractId = int64(holdingMap["contract_id"].(float64))
		pos.BuyAmount = holdingMap["buy_amount"].(float64)
		pos.BuyAvailable = holdingMap["buy_available"].(float64)
		pos.BuyPriceAvg = holdingMap["buy_price_avg"].(float64)
		pos.BuyPriceCost = holdingMap["buy_price_cost"].(float64)
		pos.BuyProfitReal = holdingMap["buy_profit_real"].(float64)
		pos.SellAmount = holdingMap["sell_amount"].(float64)
		pos.SellAvailable = holdingMap["sell_available"].(float64)
		pos.SellPriceAvg = holdingMap["sell_price_avg"].(float64)
		pos.SellPriceCost = holdingMap["sell_price_cost"].(float64)
		pos.SellProfitReal = holdingMap["sell_profit_real"].(float64)
		pos.CreateDate = int64(holdingMap["create_date"].(float64))
		pos.Symbol = cp.CustomSymbol("_", true)
		posAr = append(posAr, pos)
	}
	return posAr, nil
}

func (o *OkExApi) GetFutureOrders(orderIds []string, cp CurrencyPair, contractType string) ([]FutureOrder, error) {
	postData := url.Values{}
	postData.Set("order_id", strings.Join(orderIds, ","))
	postData.Set("contract_type", contractType)
	postData.Set("symbol", cp.CustomSymbol("_", true))
	o.buildPostForm(&postData)
	body, err := HttpPostForm(o.client, FUTURE_API_BASE_URL+FUTURE_ORDERS_INFO_URI, postData)
	if err != nil {
		return nil, err
	}
	return o.parseOrders(body, cp)
}

func (o *OkExApi) GetUnfinishedFutureOrders(cp CurrencyPair, contractType string) ([]FutureOrder, error) {
	postData := url.Values{}
	postData.Set("order_id", "-1")
	postData.Set("contract_type", contractType)
	postData.Set("symbol", cp.CustomSymbol("_", true))
	postData.Set("status", "1")
	postData.Set("current_page", "1")
	postData.Set("page_length", "50")
	o.buildPostForm(&postData)
	body, err := HttpPostForm(o.client, FUTURE_API_BASE_URL+FUTURE_ORDER_INFO_URI, postData)
	if err != nil {
		return nil, err
	}
	return o.parseOrders(body, cp)
}

func (o *OkExApi) GetFee() (float64, error) {
	return 0.03, nil //期货固定0.03%手续费
}

func (o *OkExApi) GetExchangeRate() (float64, error) {
	respMap, err := HttpGet(o.client, FUTURE_API_BASE_URL+EXCHANGE_RATE_URI)
	if err != nil {
		log.Println(respMap)
		return -1, err
	}
	if respMap["rate"] == nil {
		log.Println(respMap)
		return -1, errors.New("error")
	}
	return respMap["rate"].(float64), nil
}

func (o *OkExApi) GetContractValue(cp CurrencyPair) (float64, error) {
	switch cp.CustomSymbol("_", true) {
	case "btc_usd":
		return 100, nil
	case "ltc_usd":
		return 10, nil
	}
	return -1, errors.New("error")
}

func (o *OkExApi) GetDeliveryTime() (int, int, int, int) {
	return 4, 16, 0, 0 //星期五，下午4点交割
}

func (o *OkExApi) GetKlineRecords(contract_type string, cp CurrencyPair, period string, size, since int) ([]FutureKline, error) {
	params := url.Values{}
	params.Set("symbol", cp.CustomSymbol("_", true))
	params.Set("type", period)
	params.Set("contract_type", contract_type)
	params.Set("size", fmt.Sprintf("%d", size))
	params.Set("since", fmt.Sprintf("%d", since))
	resp, err := o.client.Get(FUTURE_API_BASE_URL + FUTURE_GET_KLINE_URI + "?" + params.Encode())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var kLines [][]interface{}
	err = json.Unmarshal(body, &kLines)
	if err != nil {
		log.Println(string(body))
		return nil, err
	}
	var klineRecords []FutureKline
	for _, record := range kLines {
		r := FutureKline{}
		r.Kline = new(Kline)
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
			case 6:
				r.Vol2 = e.(float64)
			}
		}
		klineRecords = append(klineRecords, r)
	}
	return klineRecords, nil
}

func (o *OkExApi) GetExchangeName() string {
	return FUTURE_EXCHANGE_NAME
}

func (o *OkExApi) GetTrades(cp CurrencyPair, since int64) ([]Trade, error) {
	panic("unsupported")
}

func (o *OkExApi) parseOrders(body []byte, cp CurrencyPair) ([]FutureOrder, error) {
	respMap := make(map[string]interface{})
	err := json.Unmarshal(body, &respMap)
	if err != nil {
		return nil, err
	}
	if !respMap["result"].(bool) {
		return nil, errors.New(string(body))
	}

	var orders []interface{}
	orders = respMap["orders"].([]interface{})
	var futureOrders []FutureOrder
	for _, v := range orders {
		vv := v.(map[string]interface{})
		futureOrder := FutureOrder{}
		futureOrder.OrderID = int64(vv["order_id"].(float64))
		futureOrder.Amount = vv["amount"].(float64)
		futureOrder.Price = vv["price"].(float64)
		futureOrder.AvgPrice = vv["price_avg"].(float64)
		futureOrder.DealAmount = vv["deal_amount"].(float64)
		futureOrder.Fee = vv["fee"].(float64)
		futureOrder.OType = int(vv["type"].(float64))
		futureOrder.OrderTime = int64(vv["create_date"].(float64))
		futureOrder.LeverRate = int(vv["lever_rate"].(float64))
		futureOrder.ContractName = vv["contract_name"].(string)
		futureOrder.Currency = cp.CustomSymbol("_", true)
		st := int(vv["status"].(float64))
		switch st {
		case 0:
			futureOrder.Status = ORDER_UNFINISHED
		case 1:
			futureOrder.Status = ORDER_PART_FINISH
		case 2:
			futureOrder.Status = ORDER_FINISH
		case 4:
			futureOrder.Status = ORDER_CANCELING
		case -1:
			futureOrder.Status = ORDER_CANCEL
		}
		futureOrders = append(futureOrders, futureOrder)
	}
	return futureOrders, nil
}

func (o *OkExApi) buildPostForm(postForm *url.Values) error {
	postForm.Set("apiKey", o.apiKey)
	//postForm.Set("secretKey", ctx.secretKey)

	payload := postForm.Encode()
	payload = payload + "&secretKey=" + o.apiSecretKey
	payload2, _ := url.QueryUnescape(payload) // can't escape for sign
	sign, err := GetParamMD5Sign(o.apiSecretKey, payload2)
	if err != nil {
		return err
	}

	postForm.Set("sign", strings.ToUpper(sign))
	//postForm.Del("secretKey")
	//fmt.Println(postForm)
	return nil
}
