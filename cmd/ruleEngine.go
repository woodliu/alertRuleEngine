package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"github.com/woodliu/alertRuleEngine/pkg/rule"
	"github.com/woodliu/alertRuleEngine/pkg/svc"
	"log"
)

const Path = "D:\\code\\gosrc\\src\\stash.weimob.com\\devops\\alertruleengine\\engine.json"

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	cfg, err := rule.LoadEngineCfg(Path)
	if nil != err {
		log.Fatal(err)
	}

	var dir string
	dir, _ = filepath.Split(cfg.Alert.RulesPath)
	if _, err := os.Stat(dir); nil != err {
		log.Fatalf("dir:%s not exist", dir)
	}

	delGrpCh := make(chan string)
	handler, err := rule.NewHandler(dir, cfg.Alert.RulesPath, delGrpCh)
	if nil != err {
		log.Fatal(err)
	}
	handler.Run()

	svc.StartRpcServer(handler, dir, delGrpCh)
}
