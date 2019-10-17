package storage

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
	"xorm.io/core"
)

type Database struct {
	connections map[string]*xorm.EngineGroup
}

type Source struct {
	dbType string
	addr   string
	user   string
	pswd   string
	name   string
}

func (s *Source) String() string {
	u := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&interpolateParams=true&parseTime=true&loc=Local",
		//u := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		s.user,
		s.pswd,
		s.addr,
		s.name)
	//"Local")
	return u
}

var DB *Database

func openDB(key, masterUrl string, slaveUrls []string) *xorm.EngineGroup {
	if masterUrl == "" {
		log.Panicf("there has no master source in the dbconfig: %s", key)
	}

	master, err := xorm.NewEngine("mysql", masterUrl)
	if err != nil {
		log.Panic(err)
	}

	slaves := make([]*xorm.Engine, 0)
	if len(slaveUrls) > 0 {
		for _, v := range slaveUrls {
			eg, err := xorm.NewEngine("mysql", v)
			if err != nil {
				log.Panic(err)
			}
			slaves = append(slaves, eg)
		}
	}

	group, err := xorm.NewEngineGroup(master, slaves)
	if err != nil {
		log.Panic(err)
	}

	setupDB(key, group)

	return group
}

func setupDB(key string, group *xorm.EngineGroup) {
	if viper.GetBool(fmt.Sprintf("db.%s.showLog", key)) == true {
		group.ShowSQL(true)
		group.Logger().SetLevel(core.LOG_INFO)
	} else {
		group.ShowSQL(false)
	}

	var (
		maxIdleConns int
		maxOpenConns int
	)

	maxIdleConns = viper.GetInt(fmt.Sprintf("db.%s.idleConns", key))
	maxOpenConns = viper.GetInt(fmt.Sprintf("db.%s.openConns", key))

	if maxOpenConns == 0 {
		maxOpenConns = 40
	}

	if maxIdleConns == 0 {
		maxIdleConns = 20
	}

	local, _ := time.LoadLocation("Asia/Shanghai")
	group.DatabaseTZ = local
	group.TZLocation = local
	group.SetMaxOpenConns(maxOpenConns)
	group.SetMaxIdleConns(maxIdleConns)
	//group.SetConnMaxLifetime(300 * time.Second)

	if err := group.Ping(); err != nil {
		log.Errorf("Db: %s connected failed", key)
	} else {
		log.Infof("Db: %s connected success", key)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				group.Ping()
			}
		}
	}()
}

func (db *Database) Init() {
	DB = &Database{
		connections: make(map[string]*xorm.EngineGroup),
	}

	dbConfigs := viper.GetStringMap("db")
	if len(dbConfigs) < 1 {
		return
		//panic("db config is empty...")
	}

	for key, _ := range dbConfigs {
		var (
			interfaceTemp interface{}
		)
		interfaceTemp = viper.Get(fmt.Sprintf("db.%s.master", key))
		if interfaceTemp == nil {
			log.Panicf("dbconfig: %s missing master conf", key)
		}

		masterConfig := interfaceTemp.(map[string]interface{})
		s := &Source{
			dbType: "mysql",
			addr:   masterConfig["host"].(string),
			user:   masterConfig["user"].(string),
			name:   masterConfig["name"].(string),
			pswd:   masterConfig["pswd"].(string),
		}
		masterSource := s.String()

		slaveSources := make([]string, 0)
		interfaceTemp = viper.Get(fmt.Sprintf("db.%s.slaves", key))
		if interfaceTemp != nil {

			switch interfaceTemp.(type) {
			case []interface{}:
			default:
				log.Panicf("dbConfig: %s slave config is not valid", key)
			}

			slaveConfigs := interfaceTemp.([]interface{})
			for _, v := range slaveConfigs {
				switch v.(type) {
				case map[interface{}]interface{}:
				default:
					log.Panicf("dbConfig: %s slave config is not valid", key)
				}

				vv := v.(map[interface{}]interface{})

				s := &Source{dbType: "mysql"}
				for k, vvv := range vv {
					switch k.(string) {
					case "host":
						s.addr = vvv.(string)
					case "user":
						s.user = vvv.(string)
					case "name":
						s.name = vvv.(string)
					case "pswd":
						s.pswd = vvv.(string)
					}
				}
				slaveSources = append(slaveSources, s.String())
			}
		}
		DB.connections[key] = openDB(key, masterSource, slaveSources)
	}
}

func (db *Database) Close() {
	for _, c := range DB.connections {
		c.Close()
	}
}

func (db *Database) Use(dbKey string) *xorm.EngineGroup {
	if d, ok := DB.connections[dbKey]; ok {
		return d
	} else {
		return nil
	}
}
