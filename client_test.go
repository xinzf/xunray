package xunray

import (
	"github.com/json-iterator/go"
	"reflect"
	"testing"
)

type Response struct {
	Code       int                 `json:"code"`
	Message    string              `json:"message"`
	Attachment jsoniter.RawMessage `json:"result"`
}

func (this *Response) Success() bool {
	if this.Code != 200 {
		return false
	}
	return true
}

func (this *Response) Error() string {
	return this.Message
}

func (this *Response) Convert(target interface{}) error {
	if reflect.TypeOf(target).Kind() != reflect.Ptr {

	}

	return jsoniter.Unmarshal(this.Attachment, target)
}

func Test_client_Exec(t *testing.T) {
	var rsp = new(Response)
	err := Client.Call("message.sms.secode.send", nil, rsp)
	if err != nil {
		t.Error(err)
	}

	if rsp.Success() == false {
		t.Error(rsp)
	}

	data := make(map[string]interface{})
	if err = rsp.Convert(&data); err != nil {
		t.Error(err)
	}

	t.Log(data)
}
