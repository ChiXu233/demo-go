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
	Redis struct {
		Host                      string
		Port                      int
		Lock                      string `yaml:"lock_key"`
		KeyPostfix                string `yaml:"key_postfix"`
		MatchProgressKey          string `yaml:"match_progress_key"`
		GroupProgressKey          string `yaml:"group_progress_key"`
		SourceDataTaskQueueKey    string `yaml:"source_data_task_queue_key"`
		SourceDataTaskProgressKey string `yaml:"source_data_task_progress_key"`
		ADTrainingTaskQueueKey    string `yaml:"ad_training_task_queue_key"`
		ADTrainingTaskProgressKey string `yaml:"ad_training_task_progress_key"`
		ADTestTaskQueueKey        string `yaml:"ad_test_task_queue_key"`
		ADTestTaskProgressKey     string `yaml:"ad_test_task_progress_key"`
		ADPreProcessStatusKey     string `yaml:"ad_pre_process_status_key"`
		StandardGenProgressKey    string `yaml:"standard_gen_progress_key"`
	}
	Algorithm struct {
		IP                 string `yaml:"ip"`
		Port               int    `yaml:"port"`
		One                bool
		ADTrainDockerAPI   string `yaml:"ad_train_docker_api"`
		ADTrainDockerImage string `yaml:"ad_train_docker_image"`
		ADTrainGPU         string `yaml:"ad_train_gpu"`
		ADTrainDebug       bool   `yaml:"ad_train_debug"`
		ADTrainImageLimit  int    `yaml:"ad_train_image_limit"`
		ADTestImageLimit   int    `yaml:"ad_test_image_limit"`
		ADTestAPI          string `yaml:"ad_test_api"`
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
	if algIp, ok := os.LookupEnv("ALG_IP"); ok {
		c.Algorithm.IP = algIp
	}
	if algPortStr, ok := os.LookupEnv("ALG_PORT"); ok {
		if algPort, err := strconv.Atoi(algPortStr); err == nil {
			c.Algorithm.Port = algPort
		}
	}
	if algAdTrainApi, ok := os.LookupEnv("ALG_AD_TRAIN_API"); ok {
		c.Algorithm.ADTrainDockerAPI = algAdTrainApi
	}

	if algAdTestAPI, ok := os.LookupEnv("ALG_AD_TEST_API"); ok {
		c.Algorithm.ADTestAPI = algAdTestAPI
	}
}
