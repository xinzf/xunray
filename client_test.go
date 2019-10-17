package xunray

import "testing"

func Test_client_Exec(t *testing.T) {
	var sendResponse map[string]interface{}
	var verifyResponse map[string]interface{}
	reqs := make([]*ServiceRequest, 0)
	reqs = append(
		reqs,
		Client.NewRequest("message.sms.secode.send", nil, &sendResponse),
		Client.NewRequest("message.sms.secode.verify", nil, &verifyResponse),
	)

	if err := Client.Call(reqs...);err!=nil{
		t.Error(err)
	}

	t.Log(sendResponse,verifyResponse)
}
