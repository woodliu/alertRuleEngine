package rule

import (
    "encoding/json"
    "io/ioutil"
    "os"
)

type EngineCfg struct {
    Alert struct {
        Receiver  string `json:"receiver"`
        RulesPath string `json:"rulesPath"`
        VMAlert   string `json:"vmAlert"`
    } `json:"alert"`
    Port int             `json:"port"`
}

func LoadEngineCfg(path string)(*EngineCfg,error){
    f,err := os.Open(path)
    if nil != err {
        return nil,err
    }

    defer f.Close()
    c,err := ioutil.ReadAll(f)
    if nil != err {
        return nil,err
    }

    var cfg EngineCfg
    err = json.Unmarshal(c,&cfg)
    if nil != err {
        return nil,err
    }
    return &cfg,nil
}
