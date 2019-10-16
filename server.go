package xunray

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	nettools "github.com/toolkits/net"
	"github.com/xinzf/xunray/config"
	"github.com/xinzf/xunray/storage"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	Server _server
)

func (s _server) Init(projectName string) {
	ctx := context.WithValue(context.TODO(), "project_info", map[string]string{
		"name": projectName,
	})

	var err error
	if err = config.Init(ctx); err != nil {
		panic(err)
	}

	s.services = make(map[string]*service)
	s.registered = make([]string, 0)

	s.hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}

	s.address = viper.GetString("server.addr")
	s.port = viper.GetInt("server.port")

	if s.address == "" {
		ips, err := nettools.IntranetIP()
		if err != nil {
			panic(err)
		}
		s.address = ips[0]
	}

	if s.port == 0 {
		l, _ := net.Listen("tcp", ":0")
		s.port = l.Addr().(*net.TCPAddr).Port
		l.Close()
	}

	s.g = gin.New()
	runmode := viper.GetString("server.runmode")
	if runmode == "" {
		runmode = "debug"
	}
	gin.SetMode(runmode)

	s.g.Use(gin.Recovery())

	s.g.NoRoute(s._exec)

	storage.Init()

	s.errHandler = s._errorHandler
}

type _server struct {
	address  string
	hostname string
	port     int

	services   map[string]*service
	registered []string

	g   *gin.Engine
	ctx context.Context

	errHandler func(err error) interface{}
}

func (s _server) Register(name string, hdl interface{}, metaData ...map[string]string) {
	_, found := s.services[name]
	if found {
		log.Panic(fmt.Sprintf("service: %s conflict", name))
	}

	srv, err := newService(name, hdl, metaData...)
	if err != nil {
		panic(err)
	}

	s.services[name] = srv
}

func (s _server) Start(mock ...bool) (*gin.Engine, error) {

	for _, srv := range s.services {
		if err := srv.Register(s.address, s.hostname, s.port); err != nil {
			s.Stop()
			return s.g, err
		}

		s.registered = append(s.registered, srv.name)
	}

	_mock := false
	if len(mock) > 0 {
		_mock = mock[0]
	}
	if _mock == false {
		address := s.address + ":" + strconv.Itoa(s.port)
		log.Infoln("http server listen on", address)
		err := http.ListenAndServe(address, s.g).Error()
		if err != "" {
			return s.g, errors.New(err)
		}
	}

	return s.g, nil
}

func (s _server) Stop() error {
	for _, name := range s.registered {
		s.services[name].Deregister()
	}

	fmt.Println("")
	log.Infoln("http service stopped")
	return nil
}

func (s _server) _exec(ctx *gin.Context) {
	name := ctx.Query("service")
	if name == "" {
		ctx.JSON(200,s.errHandler(errors.New("missing service's name")))
		return
	}

	srv, found := s.services[name]
	if !found {
		ctx.JSON(200,s.errHandler(fmt.Errorf("service: %s not found",name)))
		return
	}

	rawData, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(200,s.errHandler(fmt.Errorf("failed to get request data,err: %s",err.Error())))
		return
	}

	rsp, err := srv.Call(ctx, rawData)
	if err != nil {
		ctx.JSON(200,s.errHandler(err))
		return
	}

	ctx.JSON(200, rsp)
}

func (s _server) _errorHandler(err error) interface{} {
	return map[string]interface{}{
		"code": -1,
		"message":err.Error(),
	}
}
