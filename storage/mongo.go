package storage

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
)

var Mongo *mongo

func init() {
	Mongo = new(mongo)
}

type mongo struct {
	session *mgo.Session
}

func (this *mongo) Init() {
	var err error
	this.session, err = mgo.Dial(viper.GetString("mongo.host"))
	if err != nil {
		logrus.Debugln("Mongodb init failed,err: ", err)
	}

	//mgo.SetDebug(true)
	//mgo.SetLogger(log.New(os.Stderr,"mgo: ",log.LstdFlags))

	this.session.SetMode(mgo.Monotonic, true)
	logrus.Debugln("MongoDB init success.")
}

func (this *mongo) Use(dbName string) *mgo.Database {
	s := this.session.Copy()
	return s.DB(dbName)
}

func (this *mongo) Close() {
	this.session.Close()
}
