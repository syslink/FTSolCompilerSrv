package dbEntity

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"sync"
	"time"
)

var dbInstance *gorm.DB
var once sync.Once

const (
	NETWORK  = "tcp"
	PORT     = 3306
)

func InitDB(userName string, password string, server string, database string) {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", userName, password, NETWORK, server, PORT, database)
		db, err := gorm.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Open mysql failed,err:%v\n", err)
			panic("Fail to open mysql.")
		}
		dbInstance = db
		dbInstance.SingularTable(true)
		dbInstance.DB().SetMaxIdleConns(100)
		dbInstance.DB().SetMaxOpenConns(1000)
		dbInstance.DB().SetConnMaxLifetime(time.Hour)
	})
}

func GetDB() *gorm.DB {
	if dbInstance == nil {
		panic("DB hasn't inited.")
	}
	return dbInstance
}

func main2() {
	InitDB("xchainunion", "pirobot001&", "rm-2ze6ix958dd7np1kg2o.mysql.rds.aliyuncs.com", "xcu_codingshare")
	fmt.Println(GetDB())
}