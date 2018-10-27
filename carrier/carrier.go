package carrier

import (
	"github.com/BurntSushi/toml"
	"github.com/cool2645/youtube-to-netdisk/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/yanzay/log"
)

var config Config
var interfaces = make([]model.Interface, 0)
var uploaders = make([]model.Uploader, 0)
var broadcasters = make([]model.Broadcaster, 0)

func Use(i interface{}) {
	switch x := i.(type) {
	case model.Interface:
		interfaces = append(interfaces, x)
	case model.Uploader:
		uploaders = append(uploaders, x)
	case model.Broadcaster:
		broadcasters = append(broadcasters, x)
	default:
		panic("the parameter implements neither interface nor uploader nor broadcaster")
	}
}

func Start() {
	_, err := toml.DecodeFile("carrier.toml", &config)
	if err != nil {
		panic(err)
	}

	log.Infof("initializing db...")
	db, err := gorm.Open("mysql", parseDSN(config))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("database init done")
	defer db.Close()

	db.AutoMigrate(&model.Keyword{}, &model.Task{})
	model.Db = db

	model.CleanTasks(model.Db)

	for _, b := range broadcasters {
		log.Infof("starting broadcaster... %s", b.Driver())
		go b.Listen()
	}

	log.Info("starting daemon...")
	go runDaemon()

	for _, i := range interfaces {
		log.Infof("starting interface... %s", i.Driver())
		go i.Start()
	}

	<-make(chan bool)
}
