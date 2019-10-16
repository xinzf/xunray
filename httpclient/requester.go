package httpclient

import (
	"bytes"
	"errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Requester interface {
	Prepare() error
	GetURI() string
	GetPostData() []byte
	GetHeaders() map[string]string
	GetMethod() string
	Handle(rsp []byte, httpStatus int, err error)
	Error() error
}

//var HttpClient *curl

func New() *curl {
	return &curl{
		requests: make([]Requester, 0),
	}
}

type curl struct {
	requests []Requester
}

func (this *curl) AddRequest(req ...Requester) {
	this.requests = append(this.requests, req...)
}

func (this *curl) Exec() error {

	_exec := func(fn func(req Requester), req1 Requester, wg *sync.WaitGroup) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case error:
					logrus.Errorln("Error: ", err.(error).Error(), " in curl.Exec()")
				}
			}

			wg.Done()
		}()

		fn(req1)
	}

	for _, req := range this.requests {
		if err := req.Prepare(); err != nil {
			return err
		}
	}

	wg := new(sync.WaitGroup)
	for _, req := range this.requests {
		wg.Add(1)
		method := strings.ToUpper(req.GetMethod())
		if method == "GET" {
			go _exec(this.get, req, wg)
		} else if method == "POST" {
			go _exec(this.post, req, wg)
		} else if method == "PUT" {
			go _exec(this.put, req, wg)
		} else if method == "DELETE" {
			go _exec(this.delete, req, wg)
		}
	}

	wg.Wait()
	this.requests = []Requester{}
	return nil
}

func (this *curl) setHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func (this *curl) do(client *http.Client, req *http.Request) (body []byte, status int, err error) {
	rsp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, rsp.StatusCode, errors.New("request failed to " + req.URL.String())
	}
	body, _ = ioutil.ReadAll(rsp.Body)
	return body, 200, nil
}

func (this *curl) get(req Requester) {
	client := &http.Client{}
	req1, _ := http.NewRequest("GET", req.GetURI(), nil)

	this.setHeaders(req1, req.GetHeaders())

	body, status, err := this.do(client, req1)
	req.Handle(body, status, err)
}

func (this *curl) delete(req Requester) {
	client := &http.Client{}
	req1, _ := http.NewRequest("DELETE", req.GetURI(), nil)

	this.setHeaders(req1, req.GetHeaders())

	body, status, err := this.do(client, req1)
	req.Handle(body, status, err)
}

func (this *curl) post(req Requester) {
	client := &http.Client{}
	rawData := bytes.NewBuffer(req.GetPostData())
	req1, _ := http.NewRequest("POST", req.GetURI(), rawData)
	this.setHeaders(req1, req.GetHeaders())
	body, status, err := this.do(client, req1)
	req.Handle(body, status, err)
}

func (this *curl) put(req Requester) {
	client := &http.Client{}
	rawData := bytes.NewBuffer(req.GetPostData())
	req1, _ := http.NewRequest("PUT", req.GetURI(), rawData)
	this.setHeaders(req1, req.GetHeaders())
	body, status, err := this.do(client, req1)
	req.Handle(body, status, err)
}
