package config

import (
	"github.com/jinzhu/configor"
	"os"
	"strconv"
)

type Config struct {
	APP struct {
		Name string
		IP   string
		Port int
		Mode string
	}
	DB struct {
		Name     string `yaml:"name"`
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Port     uint   `yaml:"port"`
	}
}

var Conf = Config{}

func InitConfig() error {
	err := configor.Load(&Conf, "config.yaml")
	if err != nil {
		return err
	}
	Conf.loadConfFromEnv()
	return nil
}

// 编辑环境变量
func (c *Config) loadConfFromEnv() {
	if dbHost, ok := os.LookupEnv("DB_Host"); ok {
		c.DB.Host = dbHost
	}
	if dbName, ok := os.LookupEnv("DB_Name"); ok {
		c.DB.Name = dbName
	}
	if dbPortStr, ok := os.LookupEnv("DB_Port"); ok {
		if dbPort, err := strconv.Atoi(dbPortStr); err != nil {
			c.DB.Port = uint(dbPort)
		}
	}
	if dbUser, ok := os.LookupEnv("DB_User"); ok {
		c.DB.User = dbUser
	}
	if dbPassword, ok := os.LookupEnv("DB_Password"); ok {
		c.DB.Host = dbPassword
	}
}
