package coinapi

type Api interface {
	GetDepth(cp CurrencyPair, size int) (*Depth, error)

	LimitBuy(amount, price string, cp CurrencyPair) (*Order, error)

	LimitSell(amount, price string, cp CurrencyPair) (*Order, error)

	MarketBuy(amount, price string, cp CurrencyPair) (*Order, error)

	MarketSell(amount, price string, cp CurrencyPair) (*Order, error)

	CancelOrder(orderId string, cp CurrencyPair) (bool, error)

	GetOneOrder(orderId string, cp CurrencyPair) (*Order, error)

	GetUnfinishedOrders(cp CurrencyPair) ([]Order, error)

	GetAccount() (*Account, error)

	GetTicker(cp CurrencyPair) (*Ticker, error)

	Withdraw(amount, currency, fees, receiveAddr, memo, safePwd string) (string, error)

	GetExchangeName() string

	GetKlineRecords(cp CurrencyPair, period string, size, since int) ([]Kline, error)

	GetOrderHistory(cp CurrencyPair, currentPage, pageSize int) ([]Order, error)

	//非个人，整个交易所的交易记录
	GetTrades(cp CurrencyPair, since int64) ([]Trade, error)
}
