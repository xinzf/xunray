package mock

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/xinzf/xunray"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

var (
	g *gin.Engine
)

func Init(projectName string) {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())
	xunray.Server.Init(projectName)
	g, err = xunray.Server.Start(true)
	if err != nil {
		log.Panic(err)
	}
}

func New(t *testing.T) *request {
	return &request{t: t}
}

type request struct {
	t       *testing.T
	host    string
	headers map[string]string
}

func (this *request) Query(serviceName string, body ...interface{}) *response {
	req := this._makeRequest(serviceName, body...)
	return this._request(req)
}

//
//func (this *request) GET(path string) *response {
//	req := this._makeRequest(path, "get")
//	return this._request(req)
//}
//
//func (this *request) POST(path string, body ...interface{}) *response {
//	req := this._makeRequest(path, "post", body...)
//	return this._request(req)
//}
//
//func (this *request) PUT(path string, body ...interface{}) *response {
//	req := this._makeRequest(path, "put", body...)
//	return this._request(req)
//}
//
//func (this *request) DELETE(path string, body ...interface{}) *response {
//	req := this._makeRequest(path, "delete", body...)
//	return this._request(req)
//}

func (this *request) Headers(mp map[string]string) *request {
	this.headers = mp
	return this
}

func (this *request) _request(req *http.Request) *response {
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	body, _ := ioutil.ReadAll(result.Body)
	return &response{
		body: body,
		rsp:  w.Result(),
		t:    this.t,
	}
}

func (this *request) _makeRequest(name string, body ...interface{}) *http.Request {
	_url := "/?service=" + name
	var req *http.Request

	if len(body) > 0 {
		jsonData, _ := jsoniter.Marshal(body[0])
		_body := bytes.NewReader(jsonData)
		req = httptest.NewRequest("POST", _url, _body)
	} else {
		req = httptest.NewRequest("POST", _url, nil)
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range this.headers {
		req.Header.Add(k, v)
	}

	return req
}

//func (this *request) parseToStr(mp map[string]interface{}) string {
//	data := make(url.Values)
//
//	for key, val := range mp {
//		data[key] = []string{fmt.Sprintf("%v", val)}
//	}
//
//	return data.Encode()
//}
