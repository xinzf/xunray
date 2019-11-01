package mock

import (
	"github.com/json-iterator/go"
	"net/http"
	"reflect"
	"testing"
)

type response struct {
	body []byte
	rsp  *http.Response
	t    *testing.T
}

func (this *response) Equal(val interface{}) *response {
	var v interface{}
	err := jsoniter.Unmarshal(this.body, &v)
	if err != nil {
		this.t.Error("Unmarshal response failed: ", err)
	} else if reflect.DeepEqual(v, val) == false {
		this.t.Errorf("%s = %v, want %v", this.rsp.Request.URL, v, val)
	}

	return this
}

func (this *response) SeeJson(mp map[string]interface{}) *response {

	for key, val := range mp {
		v := jsoniter.Get(this.body, key).GetInterface()
		if reflect.DeepEqual(v, val) == false {
			this.t.Errorf("the key %s = %v, want %v", key, v, val)
		}
	}

	return this
}

func (this *response) Bind(val interface{}) *response {
	_v := reflect.ValueOf(val).Interface()

	if err := jsoniter.Unmarshal(this.body, &_v); err != nil {
		this.t.Errorf("Unmashal response failed,err: %s", err.Error())
	}

	if reflect.DeepEqual(_v, val) == false {
		this.t.Errorf("The %s != %+v,want %+v", this.rsp.Request.URL, _v, val)
	}
	return this
}

func (this *response) Replay(status int) *response {
	if this.rsp.StatusCode != status {
		this.t.Errorf("Http Status != %d,it is %d", status, this.rsp.StatusCode)
	}
	return this
}
