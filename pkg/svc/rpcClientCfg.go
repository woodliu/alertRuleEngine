package svc

import (
    "encoding/json"
    "io/ioutil"
    "os"
)

type ClientCfg struct {
    Clusters []Cluster `json:"clusters"`
    Port int           `json:"port"`
}

type Cluster struct {
    Engines []string `json:"engines"`
    Name    string   `json:"name"`
}

func LoadProxyCfg(path string)(*ClientCfg,error){
    f,err := os.Open(path)
    if nil != err {
        return nil,err
    }

    defer f.Close()
    c,err := ioutil.ReadAll(f)
    if nil != err {
        return nil,err
    }

    var cfg ClientCfg
    err = json.Unmarshal(c,&cfg)
    if nil != err {
        return nil,err
    }
    return &cfg,nil
}