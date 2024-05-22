package model

import (
	. "demo-go/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DBConf := Conf.DB
	//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s TimeZone=Asia/Shanghai",
		DBConf.Host, DBConf.Port, DBConf.User, DBConf.Name, DBConf.Password)
	fmt.Println(dbInfo, "链接信息")
	DB, err = gorm.Open(postgres.Open(dbInfo), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}
