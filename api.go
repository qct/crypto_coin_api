package coinapi

type Api interface {
    GetDepth(cp CurrencyPairV2, size int) (*Depth, error)

    LimitBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error)

    LimitSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error)

    CancelOrder(orderId string, cp CurrencyPairV2) (bool, error)

    GetOneOrder(orderId string, cp CurrencyPairV2) (*OrderV2, error)

    GetUnfinishedOrders(cp CurrencyPairV2) ([]OrderV2, error)

    GetAccount() (*AccountV2, error)

    GetTicker(cp CurrencyPairV2) (*Ticker, error)

    Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error)

    GetExchangeName() string

    GetKlineRecords(cp CurrencyPairV2, period string, size, since int) ([]Kline, error)

    GetOrderHistory(cp CurrencyPairV2, currentPage, pageSize int) ([]OrderV2, error)

    //非个人，整个交易所的交易记录
    GetTrades(cp CurrencyPairV2, since int64) ([]Trade, error)

    MarketBuy(amount, price string, cp CurrencyPairV2) (*OrderV2, error)

    MarketSell(amount, price string, cp CurrencyPairV2) (*OrderV2, error)
}
