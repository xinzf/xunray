package utils

import (
	"errors"
	"github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"reflect"
	"strconv"
	"strings"
)

type Convert struct {
	data interface{}
	kind reflect.Kind
}

func NewConvert(obj interface{}) *Convert {
	return &Convert{
		data: obj,
		kind: reflect.ValueOf(obj).Kind(),
	}
}

func (this *Convert) Int(defaultVal ...int) int {
	var val int
	if len(defaultVal) > 0 {
		val = defaultVal[0]
	}
	if this.data == nil {
		return val
	}
	switch this.kind {
	case reflect.Int:
		return this.data.(int)
	case reflect.Int64:
		return int(this.data.(int64))
	case reflect.Int32:
		return int(this.data.(int32))
	case reflect.Int8:
		return int(this.data.(int8))
	case reflect.Float64:
		d := decimal.NewFromFloat(this.data.(float64))
		return int(d.IntPart())
	case reflect.Float32:
		d := decimal.NewFromFloat32(this.data.(float32))
		return int(d.IntPart())
	case reflect.String:
		d, err := decimal.NewFromString(this.data.(string))
		if err != nil {
			return val
		}
		return int(d.IntPart())
	default:
		return val
	}
}

func (this *Convert) Int64(defaultVal ...int64) int64 {
	var val int64
	if len(defaultVal) > 0 {
		val = defaultVal[0]
	}

	if this.data == nil {
		return val
	}
	switch this.kind {
	case reflect.Int64:
		return this.data.(int64)
	default:
		return int64(this.Int(int(val)))
	}
}

func (this *Convert) Float64(defaultVal ...float64) float64 {
	var val float64
	if len(defaultVal) > 0 {
		val = defaultVal[0]
	}

	if this.data == nil {
		return val
	}

	switch this.kind {
	case reflect.Float64:
		return this.data.(float64)
	case reflect.Float32:
		return float64(this.data.(float32))
	case reflect.String:
		d, err := strconv.ParseFloat(this.data.(string), 64)
		if err != nil {
			return val
		}
		return d
	default:
		return float64(this.Int64(int64(val)))
	}
}

func (this *Convert) String(defaultVal ...string) string {
	var val string
	if len(defaultVal) > 0 {
		val = defaultVal[0]
	}

	if this.data == nil {
		return val
	}
	switch this.kind {
	case reflect.String:
		return this.data.(string)
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int8, reflect.Float32, reflect.Float64:
		d := decimal.NewFromFloat(this.Float64())
		return d.String()
	default:
		return val
	}
}

func (this *Convert) GetData() interface{} {
	return this.data
}

func (this *Convert) GetKind() reflect.Kind {
	return this.kind
}

func (this *Convert) Bind(obj interface{}) error {
	if this.data == nil {
		return nil
	}

	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return errors.New("绑定 Value 失败，因为目标参数不是有效指针")
	}

	b, err := jsoniter.Marshal(this.data)
	if err != nil {
		return err
	}

	return jsoniter.Unmarshal(b, obj)
}

func (this *Convert) SeparateStringSlice(separator string) []string {
	return strings.Split(this.String(), separator)
}

func (this *Convert) SeparateIntSlice(separator string) []int {
	datas := make([]int, 0)
	for _, s := range this.SeparateStringSlice(separator) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		datas = append(datas, i)
	}
	return datas
}

func (this *Convert) SeparateFloat64Slice(separator string) []float64 {
	datas := make([]float64, 0)
	for _, s := range this.SeparateStringSlice(separator) {
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		datas = append(datas, i)
	}

	return datas
}

func (this *Convert) Boolean() bool {
	if this.kind == reflect.Bool {
		return this.data.(bool)
	}

	var ret bool
	switch this.String() {
	case "true", "True", "TRUE", "1":
		ret = true
	}
	return ret
}

func (this *Convert) SeparateBooleanSlice(separator string) []bool {
	datas := make([]bool, 0)
	for _, s := range this.SeparateStringSlice(separator) {
		switch s {
		case "true", "True", "TRUE", "1":
			datas = append(datas, true)
		default:
			datas = append(datas, false)
		}
	}
	return datas
}

func (this *Convert) Float64Slice() (numbers []float64) {
	strs := this.StringSlice()
	numbers = make([]float64, 0)
	for _, s := range strs {
		d, flag := decimal.NewFromString(s)
		if flag != nil {
			n, _ := d.Float64()
			numbers = append(numbers, n)
		}
	}
	return
}

func (this *Convert) StringSlice() (strs []string) {
	strs = make([]string, 0)

	switch this.kind {
	case reflect.Slice:
		data := this.data.([]interface{})
		if len(data) == 0 {
			return
		}

		for _, v := range data {
			strs = append(strs, NewConvert(v).StringSlice()...)
		}
	default:
		strs = append(strs, this.String())
	}

	return strs
}

func (this *Convert) IntSlice() (ints []int) {
	ints = make([]int, 0)

	strs := this.StringSlice()
	for _, s := range strs {
		num, _ := strconv.Atoi(s)
		ints = append(ints, num)
	}

	//switch this.kind {
	//case reflect.Slice:
	//	data := this.data.([]interface{})
	//	if len(data) == 0 {
	//		return
	//	}
	//
	//	switch reflect.ValueOf(data[0]).Kind() {
	//	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Int16, reflect.Float32, reflect.Float64:
	//	default:
	//		return
	//	}
	//
	//	for _, v := range data {
	//		switch reflect.ValueOf(v).Kind() {
	//		case reflect.Int:
	//			ints = append(ints, v.(int))
	//		case reflect.Int8:
	//			ints = append(ints, int(v.(int8)))
	//		case reflect.Int16:
	//			ints = append(ints, int(v.(int16)))
	//		case reflect.Int32:
	//			ints = append(ints, int(v.(int32)))
	//		case reflect.Int64:
	//			ints = append(ints, int(v.(int64)))
	//		case reflect.Float32:
	//			ints = append(ints, int(v.(float32)))
	//		case reflect.Float64:
	//			ints = append(ints, int(v.(float64)))
	//		}
	//	}
	//}

	return ints
}
