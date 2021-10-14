package tmpl

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Group struct {
	GroupBy        []string `yaml:"group_by"`  //group by appId //TODO:delete?
	GroupWait      string   `yaml:"group_wait"`
	GroupInterval  string   `yaml:"group_interval"`
	RepeatInterval string   `yaml:"repeat_interval"`
}

type Rule struct {
	Name   string `yaml:"name"`
	Expr   string `yaml:"expr"`
	ID     int64  `yaml:"id"`
	For    string `yaml:"for,omitempty"`
	Labels map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type RuleTmpl struct {
	Group Group    `yaml:"group"`
	Rules []Rule   `yaml:"rules"`
}

func Load(path string)(*RuleTmpl,error){
	f,err := os.Open(path)
	if nil != err {
		return nil,err
	}

	defer f.Close()
	c,err := ioutil.ReadAll(f)
	if nil != err {
		return nil,err
	}

	var tmpl RuleTmpl
	err = yaml.Unmarshal(c,&tmpl)
	if nil != err {
		return nil,err
	}

	for _,v := range tmpl.Rules{
		if 0 != len(v.For){
			if _,err := time.ParseDuration(v.For);nil != err{
				return nil, fmt.Errorf("ruleId:%d, 'for':%s err",v.ID,v.For)
			}
		}
	}

	return &tmpl,nil
}

func (tmpl *RuleTmpl)GenTmplRuleMap() map[int64]Rule{
	ruleMap := make(map[int64]Rule)
	for _,v := range tmpl.Rules{
		if _,ok := ruleMap[v.ID];ok{
			log.Fatalf("multi rule tmpl for ruleId:%d",v.ID)
		}
		ruleMap[v.ID] = v
	}
	return ruleMap
}