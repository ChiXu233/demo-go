package utils

import (
	. "demo-go/config"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis"
	"github.com/wonderivan/logger"
)

var RedisClient *redis.Client

func InitRedis() error {
	redisConf := Conf.Redis
	redisOptions := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConf.Host, redisConf.Port),
		Password: "",
		DB:       0,
	}
	RedisClient = redis.NewClient(&redisOptions)
	_, err := RedisClient.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

type RedisStruct struct {
	Count    uint `json:"count"`
	Now      uint `json:"now"`
	Pass     uint `json:"pass"`
	Failed   uint `json:"failed"`
	Success  uint `json:"success"`
	Reserved uint `json:"reserved"`
	Ratio    uint `json:"ratio"`
}

func ReadProgressFromRedis(key string) (RedisStruct, error) {
	//读取redis
	redisKey := Conf.Redis.MatchProgressKey
	matchKey := key
	var redisStruct RedisStruct
	matchProgress := RedisClient.HMGet(redisKey, matchKey).Val()
	if matchProgress != nil {
		if len(matchProgress) != 0 {
			err := json.Unmarshal([]byte(matchProgress[0].(string)), &redisStruct)
			if err != nil {
				logger.Error("查询进度失败", err)
				return RedisStruct{}, err
			}
			return redisStruct, nil
		} else {
			logger.Info("redisProgress = []")
			return RedisStruct{}, fmt.Errorf("redisProgress = []")
		}
	} else {
		logger.Info("redisProgress is nil")
		return RedisStruct{}, fmt.Errorf("redisProgress is nil")
	}

}

func RedisPush(key string, content interface{}, method string) error {
	var err error
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}
	if method == "L" {
		err = RedisClient.LPush(key, jsonContent).Err()
		return err
	}

	if method == "R" {
		err = RedisClient.RPush(key, jsonContent).Err()
		return err
	}
	return errors.New("unsupported method")
}

func InitRedisProgress(total int, redisKey string, contentKey string, url string, params interface{}, lock *sync.Mutex) error {
	var err error
	progressMap := make(map[string]interface{})
	progressContent := make(map[string]interface{})
	progressContent["status"] = "doing"
	progressContent["message"] = ""
	progressContent["total"] = total
	progressContent["processed"] = 0
	progressContent["success"] = 0
	progressContent["failed"] = 0
	progressContent["url"] = url
	progressContent["params"] = params
	progressContentJson, err := json.Marshal(progressContent)
	if err != nil {
		return err
	}
	progressMap[contentKey] = progressContentJson
	lock.Lock()
	defer lock.Unlock()
	if err = RedisClient.HMSet(redisKey, progressMap).Err(); err != nil {
		logger.Error(err)
	}
	return nil
}

func ResetRedisProgress(redisKey string, contentKey string, lock *sync.Mutex) error {
	var err error
	progressMap := make(map[string]interface{})
	progressContent := make(map[string]interface{})
	progressContent["status"] = "none"
	progressContent["message"] = ""
	progressContent["total"] = 0
	progressContent["processed"] = 0
	progressContent["success"] = 0
	progressContent["failed"] = 0
	progressContent["url"] = ""
	progressContentJson, err := json.Marshal(progressContent)
	if err != nil {
		return err
	}
	progressMap[contentKey] = progressContentJson
	lock.Lock()
	defer lock.Unlock()
	if err = RedisClient.HMSet(redisKey, progressMap).Err(); err != nil {
		logger.Error(err)
	}
	return nil
}

// UpdateRedisProgress 增加了一个成功或者失败后调用此方法
func UpdateRedisProgress(status string, redisKey string, contentKey string, attributes string, lock *sync.Mutex) (bool, error) {
	var err error
	var completed bool
	completed = false
	lock.Lock()
	defer lock.Unlock()
	progressMapString := make(map[string]string)
	progressMap := make(map[string]interface{})
	if err = RedisClient.HGetAll(redisKey).Err(); err != nil {
		return completed, err
	}
	progressMapString = RedisClient.HGetAll(redisKey).Val()
	for k, v := range progressMapString {
		progressMap[k] = interface{}(v)
	}

	progressContent := make(map[string]interface{})
	// progress := utils.RedisClient.HMGet(redisKey, contentKey).Val()[0]
	err = json.Unmarshal([]byte(progressMap[contentKey].(string)), &progressContent)
	if err != nil {
		logger.Error(progressMap[contentKey])
		logger.Error(err)
		return completed, err
	}
	processed := progressContent["processed"].(float64)
	total := progressContent["total"].(float64)
	if attributes == "true" {
		progressContent["status"] = "done"
		progressContent["processed"] = total
	} else if attributes == "failed" {
		progressContent["status"] = "failed"
		progressContent["processed"] = total
	} else if attributes == "done" {
		progressContent["status"] = "done"
		progressContent["processed"] = total
	} else if attributes == "doing" { // 防止状态是进行中时，processed超过total （有初始化时预估total不准确的情况）
		progressContent["status"] = "doing"
		if processed+1 < total {
			progressContent["processed"] = processed + 1
		}
	} else {
		if processed+1 >= total {
			completed = true
			if attributes != "mock" {
				progressContent["status"] = "done"
			}
		}
		if status == "success" {
			if attributes != "mock" || completed == false {
				success := progressContent["success"].(float64)
				progressContent["success"] = success + 1
			}
		} else if status == "failed" {
			if attributes != "mock" || completed == false {
				failed := progressContent["failed"].(float64)
				progressContent["failed"] = failed + 1
			}
		}
		if attributes != "mock" || completed == false {
			progressContent["processed"] = processed + 1
		}
	}
	progressContentJson, err := json.Marshal(progressContent)
	if err != nil {
		logger.Error(err)
		return completed, err
	}
	progressMap[contentKey] = string(progressContentJson)
	if err = RedisClient.HMSet(redisKey, progressMap).Err(); err != nil {
		logger.Error(err)
		return completed, err
	}
	return completed, nil
}

func QueryRedisProgress(redisKey string, contentKey string, progressContent *map[string]interface{}) (bool, error) {
	var err error
	completed := false
	progress := RedisClient.HMGet(redisKey, contentKey).Val()
	if progress != nil && len(progress) != 0 && progress[0] != nil {
		err = json.Unmarshal([]byte(progress[0].(string)), &progressContent)
		if err != nil {
			return false, err
		}
	}
	if (*progressContent)["status"] == nil {
		(*progressContent)["status"] = "none"
		return false, err
	}
	processed := (*progressContent)["processed"].(float64)
	total := (*progressContent)["total"].(float64)
	if processed == total {
		completed = true
	}
	return completed, err
}

func SetKeyContent(redisKey string, contentKey string, content interface{}, lock *sync.Mutex) error {
	redisContent := make(map[string]interface{})
	redisContent[contentKey] = content
	lock.Lock()
	defer lock.Unlock()
	if err := RedisClient.HMSet(redisKey, redisContent).Err(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func CleanKeyContent(redisKey string, contentKey string, lock *sync.Mutex) error {
	redisContent := make(map[string]interface{})
	redisContent[contentKey] = nil
	lock.Lock()
	defer lock.Unlock()
	if err := RedisClient.HMSet(redisKey, redisContent).Err(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func RedisHExists(redisKey string, contentKey string) bool {
	exists, err := RedisClient.Exists(redisKey).Result()
	if exists <= 0 || err != nil {
		return false
	}
	hExists, _ := RedisClient.HExists(redisKey, contentKey).Result()
	return hExists

}
