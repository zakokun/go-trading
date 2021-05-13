package DB

import (
	"go-trading/conf"
	"fmt"
	"github.com/jinzhu/gorm"
)

var (
	DB *gorm.DB
)

type dayCandle struct {
	Symbol string
	Open   float64
	Close  float64
	Low    float64
	High   float64
	TS     int64
}

func initDB() {
	conn, err := gorm.Open("mysql", buildDSN())
	if err != nil {
		panic(err)
	}
	DB = conn
}

func buildDSN() string {
	sbl := "%s:%s@tcp(%s:%d)/%s?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8"
	dbConf := conf.Get().DB
	return fmt.Sprintf(sbl, dbConf.User, dbConf.Password, dbConf.Addr, dbConf.Port, dbConf.DBName)
}

func GetDB() *gorm.DB {
	if DB == nil {
		initDB()
	}
	return DB
}
