package model

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	TimeFormatLayout = "2006-01-02 15:04:05"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"update_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// CreateEntity create any db entity use *DBModelType
func CreateEntity(DBExecutor *gorm.DB, entity interface{}) error {
	if err := DBExecutor.Create(entity).Debug().Error; err != nil {
		return err
	}
	return nil
}

// CreateEntities batch create db entity use *[]DBModelType
func CreateEntities(DBExecutor *gorm.DB, entities interface{}) error {
	DBResult := DBExecutor.Create(entities)
	if err := DBResult.Error; err != nil {
		return err
	}
	return nil
}

// DeleteEntity delete any db entity use *DBModelType
func DeleteEntity(DBExecutor *gorm.DB, entity interface{}) (int64, error) {
	DBResult := DBExecutor.Where(entity).Delete(entity)
	if err := DBResult.Error; err != nil {
		return 0, err
	}
	return DBResult.RowsAffected, nil
}

// DeleteEntities delete batch entities by filter
func DeleteEntities(DBExecutor *gorm.DB, filter *map[string]interface{}, entity interface{}) (int64, error) {
	query := DBExecutor
	if filter != nil {
		for key, value := range *filter {
			switch {
			case strings.HasSuffix(key, "IN"):
				query = query.Where(key+" (?)", value)
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}
	result := query.Delete(entity)
	return result.RowsAffected, result.Error
}

// QueryEntity query any single entity
func QueryEntity(entityID interface{}, entity interface{}, DBExecutor ...*gorm.DB) error {
	// 不存在时，会返回错误
	transaction := DB
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}
	result := transaction.Debug().Where("ID = ?", entityID).First(entity)
	if result.RowsAffected < 1 {
		return errors.New("查询结果为空")
	}
	return result.Error
}

// QueryEntityByFilter QueryEntities delete batch entities by filter
func QueryEntityByFilter(filter *map[string]interface{}, entity interface{}, DBExecutor ...*gorm.DB) error {
	// 不存在时，不会返回错误
	transaction := DB // .Debug()
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}
	query := transaction.Select("*").Where("deleted_at is null")
	if filter != nil {
		for key, value := range *filter {
			if key == "order" {
				query = query.Order(value)
			} else if key == "group" {
				query = query.Group(value.(string))
			} else if strings.HasSuffix(key, "IN") {
				query = query.Where(key+" (?)", value)
			} else if strings.HasSuffix(key, "LIKE") {
				query = query.Where(key+" ?", "%"+value.(string)+"%")
			} else if key == "select" {
				query = query.Select(value.([]string))
			} else {
				query = query.Where(key+" = ?", value)
			}
		}
	}
	result := query.Find(entity).Debug()
	return result.Error
}

// QueryCount query any db count
func QueryCount(params *map[string]interface{}, list interface{}, count *int64, DBExecutor ...*gorm.DB) error {
	transaction := DB
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}

	query := transaction.Where("deleted_at is null")
	if params != nil {
		for key, value := range *params {
			switch {
			case strings.HasSuffix(key, "IN"):
				query = query.Where(key+" (?)", value)
			case key == "distinct":
				query = query.Distinct(value)
			case strings.HasSuffix(key, "!="):
				if value == "" {
					query = query.Where(key + "''")
				} else if value == " " {
					query = query.Where(key + "' '")
				} else if value == 0 {
					query = query.Where(key + " 0")
				} else {
					query = query.Where(key+"?", value)
				}
			case strings.HasSuffix(key, "LIKE"):
				query = query.Where(key+" ?", "%"+value.(string)+"%")
			case strings.HasSuffix(key, "BETWEEN"):
				values := strings.Split(value.(string), ",")
				query = query.Where(key+"? AND ?", values[0], values[1])
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}
	if err := query.Find(list).Count(count).Error; err != nil {
		return err
	}
	return nil
}

// QueryList query any db entity list
func QueryList(params *map[string]interface{}, list interface{}, DBExecutor ...*gorm.DB) error {

	transaction := DB
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}

	query := transaction.Where("deleted_at is null") // .Debug()
	if params != nil {
		for key, value := range *params {
			switch {
			case key == "select":
				valueStr := ""
				flagStr := ""
				for _, v := range value.([]string) {
					valueStr = valueStr + flagStr + v
					flagStr = ","
				}
				query = query.Select(valueStr)
			case key == "table":
				query = query.Table(value.(string))
			case key == "distinct":
				query = query.Distinct(value)
			case key == "order":
				query = query.Order(value)
			case key == "limit":
				query = query.Limit(value.(int))
			case key == "offset":
				query = query.Offset(value.(int))
			case key == "time":
				query = query.Where("created_at >= ? AND created_at <= ?", value.([]time.Time)[0], value.([]time.Time)[1])
			case key == "group":
				query = query.Group(value.(string))
			case strings.HasSuffix(key, "IN"):
				query = query.Where(key+" (?)", value)
			case strings.HasSuffix(key, "!="):
				if value == "" {
					query = query.Where(key + "''")
				} else if value == " " {
					query = query.Where(key + "' '")
				} else if value == 0 {
					query = query.Where(key + " 0")
				} else {
					query = query.Where(key+" ?", value)
				}
			case strings.HasSuffix(key, "?"):
				query = query.Where(key, value)
			case strings.HasSuffix(key, "LIKE"):
				query = query.Where(key+" ?", "%"+value.(string)+"%")
			case strings.Index(key, "LIKE") != -1:
				queryInfo := strings.SplitN(value.(string), " ", 2)
				query = query.Where(queryInfo[0]+" LIKE ?", "%"+queryInfo[1]+"%")
			case strings.HasSuffix(key, "BETWEEN"):
				values := strings.Split(value.(string), ",")
				query = query.Where(key+" ? AND ?", values[0], values[1])
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}
	if err := query.Debug().Find(list).Error; err != nil {
		return err
	}
	return nil
}

// QueryListReturnNumber query any db entity list
func QueryListReturnNumber(params *map[string]interface{}, list interface{}, DBExecutor ...*gorm.DB) (int64, error) {

	transaction := DB
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}

	var total int64
	query := transaction.Where("deleted_at is null")
	if params != nil {
		for key, value := range *params {
			switch {
			case key == "select":
				valueStr := ""
				flagStr := ""
				for _, v := range value.([]string) {
					valueStr = valueStr + flagStr + v
					flagStr = ","
				}
				query = query.Select(valueStr)
			case key == "distinct":
				query = query.Distinct(value)
			case key == "order":
				query = query.Order(value)
			case key == "limit":
				continue
				// query = query.Limit(value.(int))
			case key == "offset":
				continue
				// query = query.Offset(value.(int))
			case key == "time":
				query = query.Where("created_at >= ? AND created_at <= ?", value.([]time.Time)[0], value.([]time.Time)[1])
			case key == "data_time":
				query = query.Where("data_time >= ? AND data_time <= ?", value.([]time.Time)[0], value.([]time.Time)[1])
			case key == "created_at":
				query = query.Where("created_at >= ? AND created_at <= ?", value.([]time.Time)[0], value.([]time.Time)[1])
			case strings.Index(key, "LIKE") != -1:
				queryInfo := strings.Split(value.(string), " ")
				query = query.Where(queryInfo[0]+" LIKE ?", "%"+queryInfo[1]+"%")
			case strings.HasSuffix(key, "BETWEEN"):
				values := strings.Split(value.(string), ",")
				query = query.Where("? between ? AND ?", values[0], values[1], values[2])
			case strings.HasSuffix(key, "?"):
				query = query.Where(key, value)
			case strings.HasSuffix(key, "IN"):
				query = query.Where(key+" (?)", value)
			case strings.HasSuffix(key, "!="):
				if value == "" {
					query = query.Where(key + "''")
				} else if value == " " {
					query = query.Where(key + "' '")
				} else if value == 0 {
					query = query.Where(key + " 0")
				} else {
					query = query.Where(key+" ?", value)
				}
			default:
				query = query.Where(key+" = ?", value)
			}
		}
		// 这里为了获取分页前结果的总个数，给前端计算页数使用
		total = query.Find(list).RowsAffected
		for key, value := range *params {
			if key == "limit" {
				query = query.Limit(value.(int))
			} else if key == "offset" {
				query = query.Offset(value.(int))
			} else {
				continue
			}
		}
	}
	if err := query.Debug().Find(list).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// UpdateEntities 更新全部字段
func UpdateEntities(DBExecutor *gorm.DB, entities interface{}) error {
	result := DBExecutor.Save(entities).Debug()
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

// UpdateFields 更新更改的字段
func UpdateFields(DBExecutor *gorm.DB, model interface{}, selector *map[string]interface{}, fields *map[string]interface{}) error {
	query := DBExecutor.Model(&model)
	if selector != nil {
		for key, value := range *selector {
			switch {
			case key == "order":
			case strings.HasSuffix(key, "IN"):
				query = query.Where(key+" (?)", value)
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}
	result := query.Updates(fields)
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

// UpdateEntity update any db entity
func UpdateEntity(DBExecutor *gorm.DB, entity interface{}) error {
	result := DBExecutor.Debug().Model(entity).Updates(entity)
	if err := result.Error; err != nil {
		return err
	}
	return nil
}
func UpdateEntityByFilter(DBExecutor *gorm.DB, selector *map[string]interface{}, entity interface{}) error {
	if selector != nil {
		for key, value := range *selector {
			switch {
			case key == "order":
			case strings.HasSuffix(key, "IN"):
				DBExecutor = DBExecutor.Where(key+" (?)", value)
			default:
				DBExecutor = DBExecutor.Where(key+" = ?", value)
			}
		}
	}
	result := DBExecutor.Debug().Model(entity).Updates(entity)
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

// UpdateEntityByID update any db entity
func UpdateEntityByID(DBExecutor *gorm.DB, ID uint, entity interface{}) error {
	result := DBExecutor.Model(entity).Where("id = ?", ID).Updates(entity)
	if err := result.Error; err != nil {
		return err
	}
	return nil
}

func GetOrCreate(DBExecutor *gorm.DB, entity interface{}) (int64, error) {
	var err error
	// 返回的int, 表示查询到的个数, 0则代表查询无果，创建之
	result := DBExecutor.Where(entity).Limit(1).Find(entity)
	if err = result.Error; err != nil {
		return 0, err
	}
	if result.RowsAffected == 0 {
		err = CreateEntity(DBExecutor, entity)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	return result.RowsAffected, nil
}

func UpdateOrCreate(DBExecutor *gorm.DB, filter map[string]interface{}, entity interface{}, fields ...string) error {
	var err error
	var finder int
	db := *DBExecutor
	result := db.Select("id").Where(filter).Limit(1).Find(&finder)
	if err = result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		err = CreateEntity(DB, entity)
		if err != nil {
			return err
		}
	} else {
		if len(fields) != 0 {
			DBExecutor = DBExecutor.Select(fields)
		}
		result = DB.Where(filter).Updates(entity)
		if err := result.Error; err != nil {
			return err
		}
	}
	return nil
}
