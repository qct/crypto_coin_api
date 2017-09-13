package coinapi

type DepthRecord struct {
    Price  float64
    Amount float64
}

type DepthRecords []DepthRecord

func (dr DepthRecords) Len() int {
    return len(dr)
}

func (dr DepthRecords) Swap(i, j int) {
    dr[i], dr[j] = dr[j], dr[i]
}

func (dr DepthRecords) Less(i, j int) bool {
    return dr[i].Price < dr[j].Price
}

type Depth struct {
    AskList DepthRecords
    BidList DepthRecords
}

type Ticker struct {
    Last float64 `json:"last"`
    Buy  float64 `json:"buy"`
    Sell float64 `json:"sell"`
    High float64 `json:"high"`
    Low  float64 `json:"low"`
    Vol  float64 `json:"vol"`
    Date uint64 `json:"date"`
}

type Kline struct {
    Timestamp int64
    Open      float64
    Close     float64
    High      float64
    Low       float64
    Vol       float64
}

type Trade struct {
    Tid    int64 `json:"tid"`
    Type   string `json:"type"`
    Amount float64 `json:"amount,string"`
    Price  float64 `json:"price,string"`
    Date   int64 `json:"date_ms"`
}

type PoloniexCurrency struct {
    ID             int `json:"id"`
    Name           string `json:"name"`
    TxFee          float64 `json:"txFee"`
    MinConf        int `json:"minConf"`
    DepositAddress string `json:"depositAddress"`
    Disabled       int `json:"disabled"`
    Delisted       int `json:"delisted"`
    Frozen         int `json:"frozen"`
}

type OrderV2 struct {
    Price        float64
    Amount       float64
    AvgPrice     float64
    DealAmount   float64
    Fee          float64
    OrderID      int
    OrderTime    int
    Status       TradeStatus
    CurrencyPair string
    Side         TradeSide
}

type SubAccountV2 struct {
    Currency     string
    Amount       float64
    FrozenAmount float64
    LoanAmount   float64
}

type AccountV2 struct {
    Exchange      string
    Asset         float64 //总资产
    NetAsset      float64 //净资产
    SubAccountsV2 map[string]SubAccountV2
}

//-------------------------- Future ------------------------------------
type FutureKline struct {
    *Kline
    Vol2 float64 //个数
}

type FutureSubAccount struct {
    Currency      string
    AccountRights float64 //账户权益
    KeepDeposit   float64 //保证金
    ProfitReal    float64 //已实现盈亏
    ProfitUnreal  float64
    RiskRate      float64 //保证金率
}

type FutureAccount struct {
    FutureSubAccounts map[string]FutureSubAccount
}

type FutureOrder struct {
    Price        float64
    Amount       float64
    AvgPrice     float64
    DealAmount   float64
    OrderID      int64
    OrderTime    int64
    Status       TradeStatus
    Currency     string
    OType        int     //1：开多 2：开空 3：平多 4： 平空
    LeverRate    int     //倍数
    Fee          float64 //手续费
    ContractName string
}

type FuturePosition struct {
    BuyAmount      float64
    BuyAvailable   float64
    BuyPriceAvg    float64
    BuyPriceCost   float64
    BuyProfitReal  float64
    CreateDate     int64
    LeverRate      int
    SellAmount     float64
    SellAvailable  float64
    SellPriceAvg   float64
    SellPriceCost  float64
    SellProfitReal float64
    Symbol         string //btc_usd:比特币,ltc_usd:莱特币
    ContractType   string
    ContractId     int64
    ForceLiquPrice float64 //预估爆仓价
}
