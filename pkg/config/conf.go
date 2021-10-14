package config

import (
	"encoding/json"
	"golang.org/x/time/rate"
	"io/ioutil"
	"os"
)

type Conf struct {
	Port  int `json:"port"`
	DB struct {
		Port     int32  `json:"port"`
		Host     string `json:"host"`
		Password string `json:"password"`
		Username string `json:"username"`
		Database string `json:"database"`
		Charset  string `json:"charset"`
	} `json:"db"`
	PromConf struct {
		Limiter struct {
			Limit rate.Limit `json:"limit"`
			Burst int        `json:"burst"`
		} `json:"limiter"`
		Prometheus []Prometheus `json:"prometheus"`
	} `json:"promConf"`
}

type Prometheus struct {
	Env  int64  `json:"env"`
	Addr string `json:"addr"`
}

func Load(path string)(*Conf,error){
	f,err := os.Open(path)
	if nil != err {
		return nil,err
	}

	defer f.Close()
	c,err := ioutil.ReadAll(f)
	if nil != err {
		return nil,err
	}

	var conf Conf
	err = json.Unmarshal(c,&conf)
	if nil != err {
		return nil,err
	}

	return &conf,nil
}

func (c *Conf)GenPromMap()map[int64]string{
	promMap := make(map[int64]string)
	for _,v := range c.PromConf.Prometheus{
		promMap[v.Env] = v.Addr
	}

	return promMap
}