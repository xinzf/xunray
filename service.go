package xunray

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
	"time"
)

const (
	CHECK_TTL = 10
)

func newService(name string, hdl interface{}, metaData ...map[string]string) (*service, error) {
	md := make(map[string]string)
	if len(metaData) > 0 {
		md = metaData[0]
	}

	config := consul.DefaultConfig()

	if addr := viper.GetString("consul.addr"); addr != "" {
		config.Address = addr
	}

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

	if s.handTyp.NumOut() < 1 || s.handTyp.NumOut() > 2 {
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

	s.out.num = s.handTyp.NumOut()
	if s.handTyp.NumOut() == 1 {
		s.out.errTpy = s.handTyp.Out(0)
	} else if s.handTyp.NumOut() == 2 {
		s.out.errTpy = s.handTyp.Out(0)
		s.out.codeTpy = s.handTyp.Out(1)
	}

	if s.out.errTpy.String() != "error" {
		return &service{}, fmt.Errorf("服务：%s 的出参不是有效的 error", name)
	}

	if s.handTyp.NumOut() == 2 {
		if s.out.codeTpy.String() != "int" {
			return &service{}, fmt.Errorf("服务：%s 的出参不是有效的 int", name)
		}
	}

	s.ctx, s.cancel = context.WithCancel(context.TODO())
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
	out struct {
		num     int
		errTpy  reflect.Type
		codeTpy reflect.Type
	}
	metaData map[string]string
	client   *consul.Client

	heartbeatTicker *time.Ticker
	ctx             context.Context
	cancel          context.CancelFunc
}

func (this *service) Call(ctx *gin.Context, rawData []byte) (interface{}, error, bool, int) {

	var hasCode bool
	if this.out.num == 2 {
		hasCode = true
	}

	reqVal := reflect.New(this.args.reqTpy)
	if len(rawData) > 0 {
		if err := jsoniter.Unmarshal(rawData, reqVal.Interface()); err != nil {
			return nil, err, hasCode, 0
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

	var (
		er   error
		code int
	)
	if values[0].IsNil() == false {
		er = values[0].Interface().(error)
		if hasCode {
			code = values[1].Interface().(int)
		}

		return nil, er, hasCode, code
	}

	rsp := rspVal.Convert(this.args.rspTpy)
	return rsp.Interface(), er, hasCode, code
}

func (this *service) Register(address, hostname string, port int) error {
	this.metaData["hostname"] = hostname
	tags := this.encodeMetadata(this.metaData)

	this.id = fmt.Sprintf("%s-%s-%d", hostname, this.name, time.Now().UnixNano())
	asr := &consul.AgentServiceRegistration{
		ID:      this.id,
		Name:    this.name,
		Tags:    tags,
		Port:    port,
		Address: address,
		Meta:    this.metaData,
		Check: &consul.AgentServiceCheck{
			CheckID:                        fmt.Sprintf("service:%s", this.id),
			DeregisterCriticalServiceAfter: "30s",
			TTL:                            fmt.Sprintf("%ds", CHECK_TTL),
			Notes:                          fmt.Sprintf("Check service: %s", this.name),
		},
	}

	if err := this.client.Agent().ServiceRegister(asr); err != nil {
		logrus.Errorf("service: %s register failed,%s\n", this.name, err.Error())
		return err
	}

	this.heartbeatTicker = time.NewTicker(time.Duration(CHECK_TTL-1) * time.Second)

	addr := viper.GetString("consul.addr")
	if addr == "" {
		addr = "http://127.0.0.1:8500"
	}
	checkURL := fmt.Sprintf("%s/v1/agent/check/pass/%s", addr, fmt.Sprintf("service:%s", this.id))

	go func() {
		defer func() {
			this.heartbeatTicker.Stop()
		}()

		req:=gorequest.New().Timeout(time.Second).SetDebug(true)
		for {
			select {
			case <-this.heartbeatTicker.C:
				_, _, errs := req.Clone().Put(checkURL).End()
				if len(errs) > 0 {
					for _, err := range errs {
						logrus.Errorln(err)
					}
				}
			case <-this.ctx.Done():
				return
			}
		}
	}()

	logrus.Infof("service: %s register success\n", this.name)

	return nil
}

func (this *service) Deregister() error {
	this.cancel()
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
