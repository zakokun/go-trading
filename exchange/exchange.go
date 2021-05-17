package exchange

var ex Ex

var (
	tickChan       = make(chan *TickData, 1024)
	Candle1DayChan = make(chan *CandleData, 1024)
)

// Ex 所有交易所要实现的接口
type Ex interface {
	// 启动交易所连接
	Start() error
	// 关闭交易所连接
	Close() error
	// 获取实时行情价格管道
	TickListener() chan *TickData
	// 获取日K数据管道
	Kindle1DayListener() chan *CandleData
	// 实际交易
	Trade(td *TradeMsg) error
}

type TradeMsg struct {
	UserId int64
	// 交易动作,包括buy-market, sell-market, buy-limit, sell-limit
	Tp string
	// 交易价格
	Price float64
	// 交易数量
	Num float64
	// 交易对象
	Symbol string
}

// 交易所返回的实时价格消息
type TickData struct {
	From   string  // 交易所名称
	Symbol string  // 交易对名称
	Price  float64 // 价格
	TS     int64   // 时间戳
}

// 交易所返回的K线价格消息
type CandleData struct {
	From   string  // 交易所名称
	Symbol string  // 交易对名称
	Open   float64 // 价格
	Close  float64 // 价格
	Low    float64 // 价格
	High   float64 // 价格
	TS     int64   // 时间戳
}

func New() Ex {
	ex = new(Huobi)
	if err := ex.Start(); err != nil {
		panic(err)
	}
	return ex
}
