package exchange

import (
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"go-trading/utils/DB"
	"time"

	"go-trading/conf"
	"go-trading/utils/log"

	"github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/huobirdcenter/huobi_golang/pkg/model/order"

	"github.com/huobirdcenter/huobi_golang/pkg/client/marketwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
)

type Huobi struct {
	Name   string
	AppKey string
	Secret string
	Wallet float64 // 余额
	Stock  float64 // 持仓
	client *marketwebsocketclient.CandlestickWebSocketClient
}

// 连接 监听数据，把各种数据写到对应的chan里面
func (h *Huobi) Start() (err error) {
	cf := conf.Get().Ex.Huobi
	h.client = new(marketwebsocketclient.CandlestickWebSocketClient).Init(cf.Host)
	h.client.SetHandler(
		h.subscribe,
		h.handler,
	)
	go h.startListener()
	return
}

func (h *Huobi) startListener() {
	for {
		cf := conf.Get().Ex.Huobi
		ct := new(client.MarketClient).Init(cf.APIHost)
		optionalRequest := market.GetCandlestickOptionalRequest{Period: market.DAY1, Size: 10}
		resp, err := ct.GetCandlestick("btcusdt", optionalRequest)
		if err != nil {
			log.Warn("ct.GetCandlestick(%s) err(%v)", "btcusdt", err)
		}
		log.Warn("v.Open.Float64(%v) err(%v)", resp, err)
		for _, v := range resp {
			op, ok := v.Open.Float64()
			if !ok {
				log.Warn("v.Open.Float64(%v) err(%v)", v.Open, op)
				continue
			}
			cl, ok := v.Close.Float64()
			if !ok {
				log.Warn("v.Close.Float64(%v) err(%v)", v.Open, err)
				continue
			}
			lo, ok := v.Low.Float64()
			if !ok {
				log.Warn("v.Low.Float64(%v) err(%v)", v.Open, err)
				continue
			}
			hi, ok := v.High.Float64()
			if !ok {
				log.Warn("v.High.Float64(%v) err(%v)", v.Open, err)
				continue
			}
			td := &CandleData{
				From:   "huobi",
				Symbol: conf.Get().Trade.Symbol,
				Open:   op,
				Close:  cl,
				Low:    lo,
				High:   hi,
				TS:     time.Now().Unix(),
			}
			Candle1DayChan <- td
		}
		time.Sleep(time.Second)
	}
}

func (h *Huobi) subscribe() {
	cf := conf.Get().Ex.Huobi

	var ls []*DB.Stocks
	if err := DB.GetDB().Table("stocks").Find(&ls).Error; err != nil {
		log.Error("db.find() err(%v)", err)
		panic(err)
	}
	for _, v := range ls {
		h.client.Subscribe(v.Symbol+"usdt", market.DAY1, "2118")
	}
}

func (h *Huobi) handler(response interface{}) {
	resp, ok := response.(market.SubscribeCandlestickResponse)
	if ok {
		if &resp != nil {
			if resp.Tick != nil {
				t := resp.Tick
				applogger.Info("Candlestick update, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
					t.Id, t.Count, t.Vol, t.Open, t.Close, t.Low, t.High)
			}

			if resp.Data != nil {
				applogger.Info("WebSocket returned data, count=%d", len(resp.Data))
				for _, t := range resp.Data {
					applogger.Info("Candlestick data, id: %d, count: %d, vol: %v [%v-%v-%v-%v]",
						t.Id, t.Count, t.Vol, t.Open, t.Count, t.Low, t.High)
				}
			}
		}
	} else {
		applogger.Warn("Unknown response: %v", resp)
	}
}

func (h *Huobi) Close() (err error) {
	h.client.Close()
	close(tickChan)
	close(Candle1DayChan)
	return
}

func (h *Huobi) Trade(td *TradeMsg) (err error) {
	cf := conf.Get().Ex.Huobi
	ct := new(client.OrderClient).Init(cf.AppKey, cf.Secret, cf.APIHost)
	od := &order.PlaceOrderRequest{
		AccountId: cf.ClientId,
		Symbol:    td.Symbol,
		Type:      td.Tp,
		Amount:    fmt.Sprintf("%.2f", td.Num),
		Price:     fmt.Sprintf("%.2f", td.Price),
	}
	_, err = ct.PlaceOrder(od)
	if err != nil {
		log.Info("PlaceOrder error!:%v", err)
		return
	}
	return
}

// TickListener 返回实时价格的channel
// 持续获取价格数据
func (h *Huobi) TickListener() chan *TickData {
	return tickChan
}

func (h *Huobi) Kindle1DayListener() chan *CandleData {
	return Candle1DayChan
}
