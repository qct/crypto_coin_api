package coinapi

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

type FutureSubAccountV2 struct {
    Currency      string
    AccountRights float64 //账户权益
    KeepDeposit   float64 //保证金
    ProfitReal    float64 //已实现盈亏
    ProfitUnreal  float64
    RiskRate      float64 //保证金率
}

type FutureAccountV2 struct {
    FutureSubAccounts map[string]FutureSubAccountV2
}

type FutureOrderV2 struct {
    Price        float64
    Amount       float64
    AvgPrice     float64
    DealAmount   float64
    OrderID      int64
    OrderTime    int64
    Status       TradeStatus
    CurrencyPair string
    OType        int     //1：开多 2：开空 3：平多 4： 平空
    LeverRate    int     //倍数
    Fee          float64 //手续费
    ContractName string
}

type FuturePositionV2 struct {
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
