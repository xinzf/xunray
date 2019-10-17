package xunray

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"github.com/json-iterator/go"
	"github.com/xinzf/xunray/httpclient"
	"math/rand"
	"reflect"
	"time"
)

var Client _client

type _client struct {
}

func (_client) NewRequest(serviceName string, body interface{}, rsp interface{}) *ServiceRequest {
	return &ServiceRequest{
		name: serviceName,
		body: body,
		rsp:  rsp,
	}
}

func (_client) Call(requests ...*ServiceRequest) (err error) {
	client := httpclient.New()

	reqs := make([]httpclient.Requester, 0)
	for _, r := range requests {
		reqs = append(reqs, httpclient.Requester(r))
	}
	client.AddRequest(reqs...)

	err = client.Exec()
	return
}

type ServiceRequest struct {
	name string
	body interface{}

	uri string
	rsp interface{}
	err error
}

func (this *ServiceRequest) Prepare() error {
	if reflect.TypeOf(this.rsp).Kind() != reflect.Ptr {
		return fmt.Errorf("service.client: %s's response is not pointer", this.name)
	}

	client, err := consul.NewClient(consul.DefaultConfig())
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

func (this *ServiceRequest) GetURI() string {
	return this.uri
}

func (this *ServiceRequest) GetPostData() []byte {
	if this.body == nil {
		return nil
	}

	b, _ := jsoniter.Marshal(this.body)
	return b
}
func (this *ServiceRequest) GetHeaders() map[string]string {
	return nil
}

func (this *ServiceRequest) GetMethod() string {
	return "POST"
}

func (this *ServiceRequest) Handle(rsp []byte, httpStatus int, err error) {
	if err != nil {
		this.err = fmt.Errorf("request '%s' failed, err: %s", this.name, err.Error())
	} else if httpStatus != 200 {
		this.err = fmt.Errorf("request '%s' failed, http status: %d", this.name, httpStatus)
	} else if err = jsoniter.Unmarshal(rsp, this.rsp); err != nil {
		this.err = fmt.Errorf("request '%s' failed, err: %s", this.name, err.Error())
	}
}

func (this *ServiceRequest) Error() error {
	return this.err
}
