package xunray

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"github.com/json-iterator/go"
	"github.com/xinzf/xunray/httpclient"
	"math/rand"
	"time"
)

func NewClient() *_client {
	return &_client{
		requests: make([]httpclient.Requester, 0),
	}
}

type _client struct {
	requests []httpclient.Requester
}

func (this *_client) Query(serviceName string, body interface{}, fn func(rsp []byte, err error)) *_client {
	req := &request{
		name: serviceName,
		body: body,
		fn:   fn,
	}

	this.requests = append(this.requests, req)
	return this
}

func (this *_client) Exec() error {
	client := httpclient.New()
	client.AddRequest(this.requests...)

	if err := client.Exec(); err != nil {
		return err
	}

	return nil
}

type request struct {
	name string
	body interface{}

	uri string
	fn  func(rsp []byte, err error)
	err error
}

func (this *request) Prepare() error {
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

func (this *request) GetURI() string {
	return this.uri
}

func (this *request) GetPostData() []byte {
	if this.body == nil {
		return nil
	}

	b, _ := jsoniter.Marshal(this.body)
	return b
}
func (this *request) GetHeaders() map[string]string {
	return nil
}

func (this *request) GetMethod() string {
	return "POST"
}

func (this *request) Handle(rsp []byte, httpStatus int, err error) {
	if err != nil {
		this.err = fmt.Errorf("request '%s' failed, err: %s", this.name, err.Error())
	} else if httpStatus != 200 {
		this.err = fmt.Errorf("request '%s' failed, http status: %d", this.name, httpStatus)
	}

	this.fn(rsp, this.err)
}

func (this *request) Error() error {
	return this.err
}
