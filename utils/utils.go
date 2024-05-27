package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/wonderivan/logger"
)

func HTTPClient(method string, url string, data interface{}, response interface{}) (int, error) {
	contentType := "application/json"
	client := &http.Client{Timeout: 600 * time.Second}
	if method == "POST" {
		jsonStr, err := json.Marshal(data)
		if err != nil {
			return 400, err
		}

		resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
		if err != nil {
			return 500, err
		}
		defer resp.Body.Close()

		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {

			return 400, err
		}

		err = json.Unmarshal(result, &response)
		if err != nil {

			return 400, err
		}

		return resp.StatusCode, nil
	} else if method == "GET" {
		str, err := json.Marshal(data)
		if err != nil {
			return 400, err
		}

		req, err := http.NewRequest("GET", url, bytes.NewReader(str))
		if err != nil {
			return 400, err
		}
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Info(err)
			return 400, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return 400, err
		}
		if resp.StatusCode != 200 {
			return resp.StatusCode, errors.New(string(body))
		} else if resp.StatusCode == 200 {
			err = json.Unmarshal(body, &response)
			if err != nil {
				return 400, err
			}
			return 200, nil
		}
	}
	return 200, nil
}

type CommonHTTPResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// String 其他类型转string
func String(s interface{}) string {
	return fmt.Sprintf("%v", s)
}

// Float64ToString float64 转 string保留一位小数
func Float64ToString(s interface{}) string {
	return fmt.Sprintf("%.1f", s)
}

// Int string 转 int64 慎用, 注意报错会返回0
func Int(s string) int64 {
	// 不判断错误, 错误时num=0
	num, _ := strconv.ParseInt(s, 10, 64)
	return num
}

// Float string 转 float64 慎用, 注意报错会返回0
func Float(s string) float64 {
	// 不判断错误, 错误时num=0
	num, _ := strconv.ParseFloat(s, 64)
	return num
}

// UInt string 转 uint 慎用, 注意报错会返回0
func UInt(s string) uint {
	// 不判断错误, 错误时num=0
	num, _ := strconv.ParseInt(s, 10, 64)
	return uint(num)
}

// Int0 string 转 int 慎用, 注意报错会返回0
func Int0(s string) int {
	// 不判断错误, 错误时num=0
	num, _ := strconv.ParseInt(s, 10, 64)
	return int(num)
}

// ParamEmptyCheck 判断结构体指定字段是否为空
func ParamEmptyCheck(cantBeEmpty []string, paramStruct interface{}) error {
	t := reflect.TypeOf(paramStruct)
	v := reflect.ValueOf(paramStruct)
	for i := 0; i < t.NumField(); i++ {
		// fmt.Println(t.Field(i).Name)
		for _, field := range cantBeEmpty {
			structFieldName := t.Field(i).Name
			if field == structFieldName {
				switch v.Field(i).Interface().(type) {
				case uint:
					if v.Field(i).Uint() == 0 {
						return errors.New("filed " + structFieldName + " can't be empty")
					}
				case int:
					if v.Field(i).Int() == 0 {
						return errors.New("filed " + structFieldName + " can't be empty")
					}
				case string:
					if v.Field(i).String() == "" {
						return errors.New("filed " + structFieldName + " can't be empty")
					}
				default:
					return nil
				}

			}
		}
	}
	return nil
}

func In(element interface{}, list []interface{}) bool {
	for _, e := range list {
		if element == e {
			return true
		}
	}
	return false
}

func UIntIn(element uint, list []uint) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}
	return false
}
func StringIn(element string, list []string) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}
	return false
}

type T interface{}

func SafeSend(ch chan int, value int) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()
	ch <- value
	return false
}

// Minus 求两个数组的差集合 SA-SB
func Minus(SA, SB []string) ([]string, []string) {
	var SC, SD []string
	setSB := make(map[string]bool)
	for _, sb := range SB {
		setSB[sb] = true
	}
	setSA := make(map[string]bool)
	for _, sa := range SA {
		setSA[sa] = true
		if !setSB[sa] {
			SC = append(SC, sa)
		}
	}
	for _, sb := range SB {
		if !setSA[sb] {
			SD = append(SD, sb)
		}
	}

	return SC, SD
}

// StringInter 求两个字符串数组的交集
func StringInter(SA, SB []string) []string {
	var SC []string
	saMap := make(map[string]string)
	for _, sa := range SA {
		saMap[sa] = sa
	}
	for _, sb := range SB {
		if _, ok := saMap[sb]; ok {
			SC = append(SC, sb)
		}
	}
	return SC
}

// StringUnion求两个字符串数组的并集
func StringUnion(SA, SB []string) []string {
	saMap := make(map[string]string)
	for _, sa := range SA {
		saMap[sa] = sa
	}
	for _, sb := range SB {
		if _, ok := saMap[sb]; ok {
			SA = append(SA, sb)
		}
	}
	return SA
}

// StringUnion求两个字符串数组的差集
func StringDifferenceA(SA, SB []string) []string {
	var SC []string
	sbMap := make(map[string]int)
	for _, sb := range SB {
		sbMap[sb]++
	}
	for _, sa := range SA {
		times := sbMap[sa]
		if times == 0 {
			SC = append(SC, sa)
		}
	}
	return SC
}

// TimeCost @brief：耗时统计函数
func TimeCost(start time.Time, message string) {
	tc := time.Since(start)
	fmt.Printf("%s: time cost = %v\n", message, tc)
}

// CaseToCamel 下划线转驼峰
func CaseToCamel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Replace(name, "id", "ID", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

func CamelToCase(name string) string {
	var result strings.Builder
	if name == "ID" {
		return "id"
	}
	name = strings.Replace(name, "ID", "Id", -1)
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			r = unicode.ToLower(r)
		}
		result.WriteRune(r)
	}
	return result.String()
}

func dfs(field reflect.Value, target string, fieldName string) (bool, interface{}) {
	t := reflect.TypeOf(field.Interface())
	if t.Kind().String() != "struct" {
		if fieldName == target {
			return true, field.String()
		}
		return false, nil
	}
	if t.PkgPath() != "DSM/models" {
		return false, nil
	}
	for i := 0; i < t.NumField(); i++ {
		if match, val := dfs(field.Field(i), target, t.Field(i).Name); match {
			return true, val
		}
	}
	return false, nil
}

// entity必须为指针类型
func DfsAssignment(entity interface{}, name string, fields map[string]interface{}) {
	val := reflect.ValueOf(entity).Elem()
	typeof := reflect.TypeOf(entity).Elem()
	if val.Kind().String() != "struct" || typeof.PkgPath() != "DSM/models" {
		if v, ok := fields[CamelToCase(name)]; ok {
			val.Set(reflect.ValueOf(v))
		}
		return
	}
	for i := 0; i < typeof.NumField(); i++ {
		DfsAssignment(val.Field(i).Addr().Interface(), typeof.Field(i).Name, fields)
	}
}

func ReadConditions(list []interface{}, conditions []string) (map[string][]interface{}, error) {
	var err error
	mapReturn := make(map[string][]interface{})
	for _, entity := range list {
		t := reflect.TypeOf(entity)
		v := reflect.ValueOf(entity)
		for i := 0; i < t.NumField(); i++ {
			for _, field := range conditions {
				fieldCamel := CaseToCamel(field)
				if _, ok := mapReturn[field]; !ok {
					mapReturn[field] = []interface{}{}
				}
				if t.Field(i).Type.Kind().String() == "struct" {
					if match, val := dfs(v.Field(i), fieldCamel, t.Field(i).Name); match {
						// 空字符串 或者 已记录
						if !(v.Field(i).String() == "" || In(val, mapReturn[field])) {
							mapReturn[field] = append(mapReturn[field], val)
						}
					}
				} else {
					structFieldName := t.Field(i).Name
					if fieldCamel == structFieldName {
						// 空字符串 或者 已记录
						if !(v.Field(i).String() == "" || In(v.Field(i).String(), mapReturn[field])) {
							mapReturn[field] = append(mapReturn[field], v.Field(i).Interface())
						}
					}
				}
			}
			for _, field := range conditions {
				t := reflect.TypeOf(mapReturn[field]).Elem()
				switch t.String() {
				case "string":
					stringArray := make([]string, 0)
					for _, returnV := range mapReturn[field] {
						stringArray = append(stringArray, returnV.(string))
					}
					sort.Strings(stringArray)
					interfaceArray := make([]interface{}, 0)
					for _, stringV := range stringArray {
						interfaceArray = append(interfaceArray, stringV)
					}
					mapReturn[field] = interfaceArray
				}
			}
		}
	}
	return mapReturn, err
}

func SaveROIFile(filePath string, roi *[]float64) error {
	var err error
	data, err := json.MarshalIndent(roi, "", "")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0777)
	// c := exec.Command("python3", "listToPly.py", filePath)
	// output, err := c.CombinedOutput()
	//fmt.Println(string(output))
	return err
}

func ReadROIFile(filePath string, roi *[]float64) error {
	var err error
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, roi)
	return err
}

func EncodeOptions(options *[]string) string {
	optionsStr := ""
	for _, option := range *options {
		optionsStr = optionsStr + "--" + option
	}
	// fmt.Println("已编码: " + optionsStr)
	return base64.StdEncoding.EncodeToString([]byte(optionsStr))
}

func DecodeOptions(optionsEncoded string) []string {
	decodedByte, _ := base64.StdEncoding.DecodeString(optionsEncoded)
	// fmt.Printf("已解码: " + string(decodedByte))
	return strings.Split(string(decodedByte), "--")
}

// LargeLetterIncrease 字母递增
func LargeLetterIncrease(s string) (error, string) {
	i := []rune(s)
	j := i[0]
	if j < 65 || j > 90 {
		return fmt.Errorf("传入的参数不是大写字母"), s
	}
	for index := range i {
		i[index] = i[index] + 1
	}
	return nil, string(i)
}

// Float64ArrToString float64列表转化为字符串
func Float64ArrToString(arr []float64) string {
	res := "["
	for i := 0; i < len(arr); i++ {
		res += strconv.FormatFloat(arr[i], 'f', 3, 32)
		if i != len(arr)-1 {
			res += ","
		}
	}
	return res + "]"
}

// GetDifference 比较对象修改前后的不同
func GetDifference(old, new interface{}) string {
	operationLog := ""
	var typeInfo1 = reflect.TypeOf(old)
	var valInfo1 = reflect.ValueOf(old)
	var valInfo2 = reflect.ValueOf(new)
	num := typeInfo1.NumField()
	for i := 0; i < num; i++ {
		key := typeInfo1.Field(i).Name
		if key == "Model" {
			continue
		}
		val1 := String(valInfo1.Field(i).Interface())
		val2 := String(valInfo2.Field(i).Interface())
		tmp1, tmp2 := []byte(val1), []byte(val2)
		sort.Slice(tmp1, func(i, j int) bool {
			return tmp1[i] < tmp1[j]
		})
		sort.Slice(tmp2, func(i, j int) bool {
			return tmp2[i] < tmp2[j]
		})
		if string(tmp1) != string(tmp2) { //记录改变的属性
			tmp := fmt.Sprintf("%s:%v -> %v;", key, val1, val2)
			operationLog = operationLog + tmp
		}
	}
	return operationLog
}

// GetDifference2 比较对象修改前后的不同
func GetDifference2(obj interface{}, fields map[string]interface{}) string {
	operationLog := ""
	var typeInfo1 = reflect.TypeOf(obj)
	var valInfo1 = reflect.ValueOf(obj)
	num := typeInfo1.NumField()
	for i := 0; i < num; i++ {
		key := typeInfo1.Field(i).Name
		val1 := valInfo1.Field(i).Interface()
		if val2, ok := fields[key]; ok {
			if String(val1) != String(val2) {
				tmp := fmt.Sprintf("%s:%v >> %v; ", key, val1, val2)
				operationLog = operationLog + tmp
			}
		}
	}
	return operationLog
}

// BBoxesOverlap 判断是否重叠
func BBoxesOverlap(box1 []float64, box2 []float64, kuobian float64) bool {
	if len(box1) > 4 {
		// 样本数据标注框可能是多边形，先转成外接矩形
		minX, minY, maxX, maxY := 9999999.0, 9999999.0, 0.0, 0.0
		for i := 0; i < len(box1); i += 2 {
			if box1[i] < minX {
				minX = box1[i]
			}
			if box1[i] > maxX {
				maxX = box1[i]
			}
			if box1[i+1] < minY {
				minY = box1[i+1]
			}
			if box1[i+1] > maxY {
				maxY = box1[i+1]
			}
		}
		box1 = []float64{minX, minY, maxX, maxY}
	}
	if box1[0]-kuobian > box2[2] {
		return false
	}
	if box1[1]-kuobian > box2[3] {
		return false
	}
	if box1[2]+kuobian < box2[0] {
		return false
	}
	if box1[3]+kuobian < box2[1] {
		return false
	}
	return true
}

func RectangleContains(rectA, rectB [][]float64) bool {
	return rectA[0][0] <= rectB[0][0] && rectA[0][1] <= rectB[0][1] && rectA[1][0] >= rectB[1][0] && rectA[1][1] >= rectB[1][1]
}

// GetStringOfStruct 遍历结构体的属性和对应值，返回字符串
func GetStringOfStruct(obj interface{}) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	res := ""
	for k := 0; k < t.NumField(); k++ {
		if t.Field(k).Name == "Model" {
			continue
		}
		if t.Field(k).Name == "PointCloud" {
			continue
		}
		tmp := fmt.Sprintf("%s: %v  ", t.Field(k).Name, v.Field(k).Interface())
		res += tmp
	}
	return res
}

// 字符串转日期
func ParseTime(layout string, timeStr string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}

func CreateTimeFromStr(timeStr string) (time.Time, error) {
	var dateTime time.Time
	if timeStr != "" {
		timeTemplate := "2006-01-02 15:04:05"
		dataTime, err := time.ParseInLocation(timeTemplate, timeStr, time.Local)
		if err != nil {
			logger.Error("时间解析失败, %s, %v", timeStr, err)
			return dataTime, err
		}
		return dataTime, nil
	}
	return dateTime, nil
}

// IndexSort // 只能作用于正整数列表
func IndexSort(array []int, max int) []int {
	if max == 0 {
		for _, value := range array {
			if value > max {
				max = value
			}
		}
	}
	tempListForSort := make([]int, max+1)
	for _, num := range array {
		tempListForSort[num]++
	}
	var sortedList []int
	for index, value := range tempListForSort {
		if value != 0 {
			sortedList = append(sortedList, index)
		}
	}
	return sortedList
}

// QuickSortString 可以改成通用类型的
func QuickSortString(array []string, begin, end int) {
	if begin < end {
		loc := partition(array, begin, end)
		QuickSortString(array, begin, loc-1)
		QuickSortString(array, loc+1, end)
	}
}

func partition(array []string, begin, end int) int {
	i := begin + 1
	j := end

	for i < j {
		if strings.Compare(array[i], array[begin]) > 0 {
			array[i], array[j] = array[j], array[i] // 交换
			j--
		} else {
			i++
		}
	}
	if strings.Compare(array[i], array[begin]) >= 0 {
		i--
	}

	array[begin], array[i] = array[i], array[begin]
	return i
}

func CopyMap(m map[string]interface{}, keys ...string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if len(keys) > 0 {
		for _, key := range keys {
			if _, ok := m[key]; !ok {
				return nil, errors.New("键" + key + "不存在")
			}
			res[key] = m[key]
		}
	} else {
		for k, v := range m {
			res[k] = v
		}
	}
	return res, nil
}
func HasDuplicateString(arr []string) string {
	seen := make(map[string]bool)
	for _, str := range arr {
		if seen[str] {
			return str
		}
		seen[str] = true
	}
	return ""
}

// TimeStr 生成2024_01_01_00_00_00格式字符串
func TimeStr() string {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	dateTimeStr := fmt.Sprintf("%04d_%02d_%02d_%02d_%02d_%02d", year, month, day, hour, minute, second)
	return dateTimeStr
}

func Multiply(a, b interface{}) int {
	switch a.(type) {
	case int:
		switch b.(type) {
		case int:
			return a.(int) * b.(int)
		case float64:
			return int(float64(a.(int)) * b.(float64))
		}
	case float64:
		switch b.(type) {
		case int:
			return int(a.(float64) * float64(b.(int)))
		case float64:
			return int(a.(float64) * b.(float64))
		}
	}
	return 0
}
