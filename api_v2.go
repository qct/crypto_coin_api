package coinapi

type ApiV2 interface {
    LimitBuy(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error)

    LimitSell(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error)

    MarketBuy(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error)

    MarketSell(amount, price, baseCurrency, counterCurrency string) (*OrderV2, error)

    CancelOrder(orderId, baseCurrency, counterCurrency string) (bool, error)

    GetOneOrder(orderId, baseCurrency, counterCurrency string) (*OrderV2, error)

    GetUnfinishedOrders(baseCurrency, counterCurrency string) ([]OrderV2, error)

    GetOrderHistory(baseCurrency, counterCurrency string, currentPage, pageSize int) ([]OrderV2, error)

    GetAccount() (*AccountV2, error)

    GetTicker(baseCurrency, counterCurrency string) (*Ticker, error)

    GetDepth(size int, baseCurrency, counterCurrency string) (*Depth, error)

    GetKlineRecords(baseCurrency, counterCurrency, period string, size, since int) ([]Kline, error)

    //非个人，整个交易所的交易记录
    GetTrades(baseCurrency, counterCurrency string, since int64) ([]Trade, error)

    GetExchangeName() string

    Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error)
}
