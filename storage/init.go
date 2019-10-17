package storage

const (
	DbName = "driver_name"
)

func Init() {
	DB.Init()
	Mongo.Init()
	Redis.Init()
}
