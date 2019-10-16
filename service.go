package xunray

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"reflect"
)

func newService(name string, hdl interface{}, metaData ...map[string]string) (*service, error) {
	md := make(map[string]string)
	if len(metaData) > 0 {
		md = metaData[0]
	}

	config := consul.DefaultConfig()
	client, err := consul.NewClient(config)
	if err != nil {
		return &service{}, err
	}

	_id, _ := uuid.NewV4()
	s := &service{
		id:       _id.String(),
		name:     name,
		handFunc: hdl,
		metaData: md,
		client:   client,
	}

	s.handTyp = reflect.TypeOf(hdl)
	if s.handTyp.Kind() != reflect.Func {
		return &service{}, fmt.Errorf("服务：%s 不是有效的方法", name)
	}

	if s.handTyp.NumIn() > 3 || s.handTyp.NumIn() < 2 {
		return &service{}, fmt.Errorf("服务：%s 入参数量错误", name)
	}

	if s.handTyp.NumOut() != 1 {
		return &service{}, fmt.Errorf("服务：%s 出参数量错误", name)
	}

	s.args.num = s.handTyp.NumIn()
	if s.handTyp.NumIn() == 2 {
		s.args.reqTpy = s.handTyp.In(0)
		s.args.rspTpy = s.handTyp.In(1)
	} else if s.handTyp.NumIn() == 3 {
		s.args.ctxTpy = s.handTyp.In(0)
		s.args.reqTpy = s.handTyp.In(1)
		s.args.rspTpy = s.handTyp.In(2)
	}

	if s.args.reqTpy.Kind() == reflect.Ptr {
		return &service{}, fmt.Errorf("服务：%s 的 request 是一个指针类型", name)
	}

	if s.args.rspTpy.Kind() != reflect.Ptr {
		return &service{}, fmt.Errorf("服务：%s 的 response 不是有效的指针类型", name)
	}

	//fmt.Println(s.args.ctxTpy.String())

	if s.args.num == 3 && s.args.ctxTpy.String() != "*gin.Context" {
		return &service{}, fmt.Errorf("服务：%s 的第一个参数不是 *gin.Context", name)
	}

	s.returnTyp = s.handTyp.Out(0)
	if s.returnTyp.String() != "error" {
		return &service{}, fmt.Errorf("服务：%s 的出参不是有效的 error", name)
	}

	return s, nil
}

type service struct {
	id       string
	name     string
	handFunc interface{}
	handTyp  reflect.Type
	args     struct {
		num    int
		ctxTpy reflect.Type
		reqTpy reflect.Type
		rspTpy reflect.Type
	}
	returnTyp reflect.Type
	metaData  map[string]string
	client    *consul.Client
}

func (this *service) Call(ctx *gin.Context, rawData []byte) (interface{}, error) {

	reqVal := reflect.New(this.args.reqTpy)
	if len(rawData) > 0 {
		if err := jsoniter.Unmarshal(rawData, reqVal.Interface()); err != nil {
			return nil, err
		}
	}

	rspVal := reflect.New(this.args.rspTpy.Elem())
	values := make([]reflect.Value, 0)
	if this.args.num == 2 {
		values = reflect.ValueOf(this.handFunc).Call([]reflect.Value{
			reqVal.Elem(),
			rspVal,
		})
	} else {
		values = reflect.ValueOf(this.handFunc).Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reqVal.Elem(),
			rspVal,
		})
	}

	if values[0].IsNil() == false {
		er := values[0].Interface().(error)
		return nil, er
	}

	rsp := rspVal.Convert(this.args.rspTpy)
	return rsp.Interface(), nil
}

func (this *service) Register(address, hostname string, port int) error {
	this.metaData["hostname"] = hostname
	tags := this.encodeMetadata(this.metaData)

	_id, _ := uuid.NewV4()
	this.id = _id.String()
	asr := &consul.AgentServiceRegistration{
		ID:      this.id,
		Name:    this.name,
		Tags:    tags,
		Port:    port,
		Address: address,
		//Check: &consul.AgentServiceCheck{
		//	HTTP:                           fmt.Sprintf("http://%s:%d?service=%s", address, port, this.name),
		//	Timeout:                        time.Duration(2*time.Second).String(),
		//	Interval:                       time.Duration(time.Minute).String(),
		//	Method:                         http.MethodHead,
		//	DeregisterCriticalServiceAfter: "30s",
		//},
	}

	if err := this.client.Agent().ServiceRegister(asr); err != nil {
		logrus.Errorf("service: %s register failed,%s\n", this.name, err.Error())
		return err
	}

	logrus.Infof("service: %s register success\n", this.name)
	return nil
}

func (this *service) Deregister() error {
	err := this.client.Agent().ServiceDeregister(this.id)
	if err != nil {
		logrus.Errorf("service: %s deregister failed,%s\n", this.name, err.Error())
		return err
	}
	logrus.Infof("service: %s deregister success\n", this.name)
	return nil
}

func (this *service) encodeMetadata(md map[string]string) []string {
	var tags []string
	for k, v := range md {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return tags
}
