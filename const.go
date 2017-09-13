package coinapi

const (
    OK_CN      = "okcoin.cn"
    OK_EX      = "okex.com"
    OK_COM     = "okcoin.com"
    HUOBI      = "huobi.com"
    CHBTC      = "chbtc.com"
    YUNBI      = "yunbi.com"
    POLONIEX   = "poloniex.com"
    COIN_CHECK = "coincheck.com"
    ZAIF       = "zaif.jp"
)

const (
    ORDER_UNFINISHED  = iota + 1
    ORDER_PART_FINISH
    ORDER_FINISH
    ORDER_CANCEL
    ORDER_REJECT
    ORDER_CANCELING
)

const (
    THIS_WEEK_CONTRACT = "this_week" //周合约
    NEXT_WEEK_CONTRACT = "next_week" //次周合约
    QUARTER_CONTRACT   = "quarter"   //季度合约
)

const (
    OPEN_BUY   = 1 + iota //开多
    OPEN_SELL             //开空
    CLOSE_BUY             //平多
    CLOSE_SELL            //平空
)

const (
    BUY         = 1 + iota
    SELL
    BUY_MARKET
    SELL_MARKET
    UNKNOWN
)

type TradeStatus int

type TradeSide int

func (ts TradeStatus) String() string {
    switch ts {
    case ORDER_UNFINISHED:
        return "UNFINISH"
    case ORDER_PART_FINISH:
        return "PART_FINISH"
    case ORDER_FINISH:
        return "FINISH"
    case ORDER_CANCEL:
        return "CANCEL"
    case ORDER_REJECT:
        return "REJECT"
    case ORDER_CANCELING:
        return "CANCEL_ING"
    default:
        return "UNKNOWN"
    }
}

func (ts TradeSide) String() string {
    switch ts {
    case BUY:
        return "BUY"
    case SELL:
        return "SELL"
    case BUY_MARKET:
        return "BUY_MARKET"
    case SELL_MARKET:
        return "SELL_MARKET"
    default:
        return "UNKNOWN"
    }
}

func StringToTradeSide(s string) TradeSide {
    switch s {
    case TradeSide(BUY).String():
        return BUY
    case TradeSide(SELL).String():
        return SELL
    case TradeSide(BUY_MARKET).String():
        return BUY_MARKET
    case TradeSide(SELL_MARKET).String():
        return SELL_MARKET
    default:
        return UNKNOWN
    }
}
