package exchange

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/huobirdcenter/huobi_golang/pkg/client/marketwebsocketclient"
	"github.com/huobirdcenter/huobi_golang/pkg/model/base"
	"github.com/huobirdcenter/huobi_golang/pkg/model/market"
	"github.com/huobirdcenter/huobi_golang/pkg/model/order"
	"github.com/shopspring/decimal"
	"go-trading/conf"
	"go-trading/utils/DB"
	"go-trading/utils/log"
	"strings"
)

type Huobi struct {
	Name   string
	client *marketwebsocketclient.CandlestickWebSocketClient
}

type SubscribeCandlestickResponse struct {
	Base base.WebSocketResponseBase
	Tick *market.Tick
	Data []market.Tick
}

// 连接 监听数据
func (h *Huobi) Start() (err error) {
	cf := conf.Get().Ex.Huobi
	spew.Dump(cf)
	h.client = new(marketwebsocketclient.CandlestickWebSocketClient).Init(cf.Host)
	h.client.SetHandler(
		h.subscribe,
		h.handler,
	)
	go h.startListener()
	return
}

func (h *Huobi) startListener() {
	h.client.Connect(true)
}

func (h *Huobi) subscribe() {
	var ls []*DB.Stocks
	if err := DB.GetDB().Table("stocks").Find(&ls).Error; err != nil {
		log.Error("db.find() err(%v)", err)
		panic(err)
	}
	applogger.Info("get db %s", spew.Sdump(ls))
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
				h.saveCandleData(resp.Ch, t.Open, t.Close, t.Low, t.High, t.Vol, t.Id)
				applogger.Info("Candlestick update, channel %v id: %d, count: %d, vol: %v [%v-%v-%v-%v]", resp.Base,
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
		applogger.Warn("Unknown response: %v", response)
	}
}

// saveCandleData save to stocks and stockDaily
func (h *Huobi) saveCandleData(key string, o, c, l, hi, vol decimal.Decimal, ts int64) (err error) {
	st := new(DB.Stocks)
	symbols := strings.Split(key, ".")
	sb := strings.Replace(symbols[1], "usdt", "", 1)
	if err = DB.GetDB().Table("stocks").Where("symbol=?", sb).First(&st).Error; err != nil {
		log.Error("saveCandleData err(%v)", err)
		return
	}
	st.Open, _ = o.Float64()
	st.Close, _ = c.Float64()
	st.Low, _ = l.Float64()
	st.High, _ = hi.Float64()
	st.TS = ts
	st.Volume, _ = vol.Float64()
	if err = DB.GetDB().Table("stocks").Save(st).Error; err != nil {
		log.Error("saveCandleData save stocks(%v) err(%v)", st, err)
		return
	}
	sd := new(DB.Stocks)
	sd.Open, _ = o.Float64()
	sd.Close, _ = c.Float64()
	sd.Low, _ = l.Float64()
	sd.High, _ = hi.Float64()
	sd.TS = ts
	sd.Volume, _ = vol.Float64()
	sd.Symbol = st.Symbol
	if err = DB.GetDB().Table("stock_daily").Set(
		"gorm:insert_option",
		"ON DUPLICATE KEY UPDATE open = VALUES(open),close=VALUES(close),low=VALUES(low),high=VALUES(high),volume=VALUES(volume)",
	).Create(sd).Error; err != nil {
		log.Error("saveCandleData save stock_daily(%v) err(%v)", sd, err)
		return
	}
	return
}

func (h *Huobi) Close() (err error) {
	h.client.Close()
	close(Candle1DayChan)
	return
}

func (h *Huobi) Trade(td *TradeMsg) (err error) {
	cf := conf.Get().Ex.Huobi
	u := new(DB.User)
	if err = DB.GetDB().Table("users").Where("user_id=?", td.UserId).First(u).Error; err != nil {
		log.Error("huobi Trade(%v) get user() err(%v)", td, err)
		return
	}
	ct := new(client.OrderClient).Init(u.AppKey, u.Secret, cf.APIHost)
	od := &order.PlaceOrderRequest{
		AccountId: "1608",
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
