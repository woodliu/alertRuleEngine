package main

import (
    "net/http"
    "github.com/woodliu/alertRuleEngine/pkg/svc"
    "log"
)
var path = "D:\\code\\gosrc\\src\\stash.weimob.com\\devops\\alertruleengine\\service.json"
func main(){
    cfg,err := svc.LoadProxyCfg(path)
    if nil != err{
        log.Fatal(err)
    }

    p := svc.NewService(cfg.Clusters)
    http.HandleFunc("/api/v1/addRule",p.Add)
    http.HandleFunc("/api/v1/removeRule",p.Remove)
    http.HandleFunc("/api/v1/updateRule",p.Update)
    http.HandleFunc("/api/v1/listRule",p.List)
}
