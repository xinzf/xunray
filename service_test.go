package xunray

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
)

type Req struct {
	UserName string `json:"user_name"`
}

type Rsp struct {
	Code int `json:"code"`
}

func Hdl(ctx *gin.Context, req Req, rsp *Rsp) error {
	fmt.Println(req.UserName)
	rsp.Code = 200
	return errors.New("ss")
}

func TestService_Call(t *testing.T) {
	srv, err := NewService("uims.test", Hdl)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := &gin.Context{}
	rawData := []byte(`{"user_name":"xiangzhi"}`)
	rsp, err := srv.Call(ctx, rawData)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(rsp)
}

func TestNewService(t *testing.T) {
	srv, err := NewService("uims.test", Hdl)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(srv, err)
}

func TestService_Deregister(t *testing.T) {
	srv, err := NewService("uims.account.login", func(ctx *gin.Context, req interface{}, rsp interface{}) error {
		return nil
	}, map[string]string{
		"title":    "用户登录",
		"describe": "使用用户名和密码登录",
	})
	if err != nil {
		t.Error(err)
		return
	}
	srv.id = "b3547ae8-6692-4456-919b-08ba45580d74"
	if err = srv.Deregister(); err != nil {
		t.Error(err)
	}
}

func TestService_Register(t *testing.T) {
	srv, err := NewService("uims.account.login", func(ctx *gin.Context, req interface{}, rsp interface{}) error {
		return nil
	}, map[string]string{
		"title":    "用户登录",
		"describe": "使用用户名和密码登录",
	})
	if err != nil {
		t.Error(err)
		return
	}

	if err = srv.Register("127.0.0.1", "xiangzhi mac", 4405); err != nil {
		t.Error(err)
		return
	}
}
