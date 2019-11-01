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

func Init(serverName string, inits ...xunray.Initialize) {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())
	xunray.Server.Init(serverName, inits...)
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
	log.Println("[Mock] Response:")
	log.Printf("HttpCode: %d", result.StatusCode)
	jsonData, _ := jsoniter.MarshalIndent(body, "", "  ")
	log.Printf("Body: %s", string(jsonData))
	return &response{
		body: body,
		rsp:  w.Result(),
		t:    this.t,
	}
}

func (this *request) _makeRequest(name string, body ...interface{}) *http.Request {
	_url := "/?service=" + name
	var req *http.Request

	log.Println("[Mock] Request:")
	log.Printf("\tService: %s", name)

	if len(body) > 0 {
		mockJsonData, _ := jsoniter.MarshalIndent(body[0], "", "  ")
		log.Printf("\tData:\n%s", string(mockJsonData))
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
