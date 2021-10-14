package engine

import (
	"context"
	"github.com/woodliu/alertRuleEngine/pkg/common"
	"github.com/woodliu/alertRuleEngine/pkg/config"
	prom "github.com/woodliu/alertRuleEngine/pkg/prom"
	tmpl "github.com/woodliu/alertRuleEngine/pkg/template"
	"log"
	"net/http"
	"time"
)

type Engine struct {
	store Storage

	groupConf *groupConf
	ruleMap map[int64]tmpl.Rule //for select rule by ruleId
	groupMap map[string]*group  //for select group by groupId

	transport http.RoundTripper
}

type groupConf struct {
	GroupWait      time.Duration
	GroupInterval  time.Duration
	RepeatInterval time.Duration
}

func newGroup(g *tmpl.Group)(*groupConf,error){
	waitTimeout,err := time.ParseDuration(g.GroupWait)
	if nil != err{
		return nil, err
	}
	intervalTimeout,err := time.ParseDuration(g.GroupInterval)
	if nil != err{
		return nil, err
	}
	repeatTimeout,err := time.ParseDuration(g.RepeatInterval)
	if nil != err{
		return nil, err
	}

	return &groupConf{
		waitTimeout,intervalTimeout,repeatTimeout,
	},nil
}

func NewEngine(store Storage, tmpl *tmpl.RuleTmpl)*Engine{
	 e := &Engine{
		store: store,
		ruleMap: tmpl.GenTmplRuleMap(),
		groupMap: make(map[string]*group),
		transport: prom.NewTransport(),
	}

	var err error
	e.groupConf,err = newGroup(&tmpl.Group)
	if nil != err{
		log.Panicf("newGroup Err %s",err)
	}

	return e
}

func (e *Engine)LoadRules() error{
	rules,err := e.store.GetAllAppRules()
	if nil != err {
		return  err
	}

	for _,v := range *rules{
		if _,ok := e.groupMap[v.AppId];!ok{
			e.groupMap[v.AppId] = &group{}
		}

		e.groupMap[v.AppId].rules = append(e.groupMap[v.AppId].rules, v)
	}

	return nil
}

func (e *Engine)QueryProm()[]group{
	return nil
}

func StartEngine(ctx context.Context,store Storage, reload chan struct{}, tmpl *tmpl.RuleTmpl,conf *config.Conf){
	e := NewEngine(store,tmpl)
	e.LoadRules()

	query := make(chan *common.AlertReq)
	receive := make(chan *common.AlertResp)
	send := make(chan interface{})

	groupChains := NewGroupChain(e, query, receive, send, reload)
	groupChains.Run(ctx)

	prometheus := prom.NewProm(e.ruleMap, conf, query, receive)
	prometheus.Run(ctx)


	trans := prom.NewTrans(send)
	trans.Send(ctx)
}
