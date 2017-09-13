package coinapi

type FutureApi interface {
	//获取交割预估价
	GetFutureEstimatedPrice(cp CurrencyPair) (float64, error)

	//期货行情; btc_usd:比特币 ltc_usd:莱特币; 合约类型: this_week:当周 next_week:下周 month:当月 quarter:季度
	GetFutureTicker(cp CurrencyPair, contractType string) (*Ticker, error)

	//期货深度; btc_usd:比特币 ltc_usd:莱特币; 合约类型: this_week:当周 next_week:下周 month:当月 quarter:季度
	GetFutureDepth(cp CurrencyPair, contractType string, size int) (*Depth, error)

	//期货指数; btc_usd: 比特币 ltc_usd: 莱特币
	GetFutureIndex(cp CurrencyPair) (float64, error)

	//全仓账户
	GetFutureUserInfo() (*FutureAccount, error)

	//期货下单; openType 1:开多 2:开空 3:平多 4:平空; 是否为对手价 0:不是 1:是, 当取值为1时, price无效
	PlaceFutureOrder(cp CurrencyPair, contractType, price, amount string, openType, matchPrice, leverRate int) (string, error)

	//取消订单
	FutureCancelOrder(cp CurrencyPair, contractType, orderId string) (bool, error)

	//用户持仓查询; btc_usd: 比特币 ltc_usd: 莱特币; 合约类型: this_week:当周 next_week:下周 month:当月 quarter:季度
	GetFuturePosition(cp CurrencyPair, contractType string) ([]FuturePosition, error)

	//获取订单信息
	GetFutureOrders(orderIds []string, cp CurrencyPair, contractType string) ([]FutureOrder, error)

	//获取未完成订单信息
	GetUnfinishedFutureOrders(cp CurrencyPair, contractType string) ([]FutureOrder, error)

	//获取交易费
	GetFee() (float64, error)

	//获取交易所的美元人民币汇率
	GetExchangeRate() (float64, error)

	//获取每张合约价值
	GetContractValue(cp CurrencyPair) (float64, error)

	//获取交割时间 星期(0,1,2,3,4,5,6)，小时，分，秒
	GetDeliveryTime() (int, int, int, int)

	//获取K线数据
	GetKlineRecords(contract_type string, cp CurrencyPair, period string, size, since int) ([]FutureKline, error)

	//获取交易所名字
	GetExchangeName() string
}
