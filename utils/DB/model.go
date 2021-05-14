package DB

type Stocks struct {
	Id     int64
	TS     int64
	Symbol string
	Open   float64
	Close  float64
	Low    float64
	High   float64
	Volume float64
}

type StockDaily struct {
	Id     int64
	TS     int64
	Symbol string
	Open   float64
	Close  float64
	Low    float64
	High   float64
	Volume int64
}
