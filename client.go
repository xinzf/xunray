package xunray

import (
	"fmt"
	"github.com/cstockton/go-conv"
	"github.com/gin-gonic/gin"
	consul "github.com/hashicorp/consul/api"
	"github.com/json-iterator/go"
	"github.com/spf13/viper"
	"github.com/xinzf/xunray/httpclient"
	"math/rand"
	"reflect"
	"time"
)

var Client _client

type _client struct {
}

func (_client) Call(serviceName string, body interface{}, rsp interface{}, headers ...map[string]interface{}) (err error) {
	client := httpclient.New()

	var req = &_serviceRequest{
		name: serviceName,
		body: body,
		rsp:  rsp,
		headers: map[string]string{
			"Content-Type": gin.MIMEJSON,
		},
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			vv, _ := conv.String(v)
			req.headers[k] = vv
		}
	}
	client.AddRequest(req)

	err = client.Exec()
	if err != nil {
		return err
	}
	if req.err != nil {
		return req.err
	}

	return
}

type _serviceRequest struct {
	name string
	body interface{}

	uri string
	rsp interface{}
	err error

	headers map[string]string
}

func (this *_serviceRequest) Prepare() error {
	if reflect.TypeOf(this.rsp).Kind() != reflect.Ptr {
		return fmt.Errorf("service.client: %s's response is not pointer", this.name)
	}

	config := consul.DefaultConfig()

	if addr := viper.GetString("consul.addr"); addr != "" {
		config.Address = addr
	}

	client, err := consul.NewClient(config)
	if err != nil {
		return err
	}

	services, _, err := client.Catalog().Service(this.name, "", &consul.QueryOptions{})
	if err != nil {
		return err
	}

	domains := make([]string, 0)
	for _, s := range services {
		domains = append(domains, fmt.Sprintf("http://%s:%d", s.ServiceAddress, s.ServicePort))
	}

	if len(domains) == 0 {
		return fmt.Errorf("未找到 service: %s 所在节点", this.name)
	}

	rand.Seed(time.Now().UnixNano())
	this.uri = fmt.Sprintf("%s?service=%s", domains[rand.Intn(len(domains))], this.name)
	return nil
}

func (this *_serviceRequest) GetURI() string {
	return this.uri
}

func (this *_serviceRequest) GetPostData() []byte {
	if this.body == nil {
		return nil
	}

	b, _ := jsoniter.Marshal(this.body)
	return b
}
func (this *_serviceRequest) GetHeaders() map[string]string {
	return this.headers
}

func (this *_serviceRequest) GetMethod() string {
	return "POST"
}

func (this *_serviceRequest) Handle(rsp []byte, httpStatus int, err error) {
	if err != nil {
		this.err = fmt.Errorf("request '%s' failed, err: %s", this.name, err.Error())
	} else if httpStatus != 200 {
		this.err = fmt.Errorf("request '%s' failed, http status: %d", this.name, httpStatus)
	} else if err = jsoniter.Unmarshal(rsp, this.rsp); err != nil {
		this.err = fmt.Errorf("request '%s' failed, err: %s", this.name, err.Error())
	}
}

func (this *_serviceRequest) Error() error {
	return this.err
}
