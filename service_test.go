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

func Test_server_Start(t *testing.T) {
	Server.Init("MESSAGE")
	Server.Register("test1", Hdl)
	//Server.Register("test2", Hdl)
	Server.Start()
}
