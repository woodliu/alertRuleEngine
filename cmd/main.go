package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/woodliu/alertRuleEngine/pkg/engine"
	tmpl "github.com/woodliu/alertRuleEngine/pkg/template"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/woodliu/alertRuleEngine/pkg/config"
	"gorm.io/driver/mysql"
)

const Path = "D:\\rules.yaml"
const Conf = "D:\\config.json"

func main() {
	tmpl,err := tmpl.Load(Path)
	if nil != err{
		return
	}

	//TODO：可能需要自定义多个级别的告警，目前仅支持一条
	conf,err := config.Load(Conf)
	if nil != err{
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	storage := engine.NewStore(newDb(conf))

	reload := make(chan struct{})
	engine.StartEngine(ctx, storage, reload, tmpl, conf)
	engine.ServeHttpRules(conf.Port, storage, reload, tmpl)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-c
	log.Printf("system shutdown by signal:[%s]",s.String())
	cancel()
}

func newDb(conf *config.Conf)*gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		conf.DB.Username,
		conf.DB.Password,
		conf.DB.Host,
		conf.Port,
		conf.DB.Database,
		conf.DB.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Discard,
	})

	if err != nil {
		logrus.Fatalf("failed to Open DB %s:%d, err:%v", conf.DB.Host, conf.Port, err)
	}
	return db
}