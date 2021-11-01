package rule

import (
	"fmt"
	vmCfg "github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/config"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/datasource"
	"os"
	"path/filepath"
	"github.com/woodliu/alertRuleEngine/pkg/comm"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
	"log"
	"sync"
)

type Handler struct {
	l      sync.RWMutex
	mapP   map[string]*Processor
	delGrp chan string
	ps     []*Processor
}

type Processor struct {
	ruleCh chan interface{}
	delGrp chan string
	apps   *Rule
}

func (h *Handler) getGrpProcessor(grpId string) *Processor {
	h.l.RLock()
	defer h.l.RUnlock()
	return h.mapP[grpId]
}

func NewProcessorFunc(dir string, delGrp chan string) func() *Processor {
	return func() *Processor {
		p := &Processor{
			delGrp: delGrp,
			ruleCh: make(chan interface{}),
			apps: &Rule{
				dir: dir,
				mApp:make(map[string]*appRules),
			},
		}
		go p.Exec()
		return p
	}
}

func (h *Handler) getAvailableProc(new func() *Processor) (*Processor, bool) {
	h.l.RLock()
	defer h.l.RUnlock()
	for _, p := range h.mapP {
		if len(p.apps.mApp) < maxGrpPerProcessor {
			return p, false
		}
	}

	return new(), true
}

// Exec 单个服务单个规则文件，可以降低单文件故障以及访问冲突。每个processor负责各自的app规则的增删改查，不使用锁
func (p *Processor) Exec() {
	for data := range p.ruleCh {
		switch data.(type) {
		case *comm.ListReq:
			msg := data.(*comm.ListReq)
			appRule := p.apps.mApp[msg.Rule.Group]
			if nil == appRule {
				p.delGrp <- msg.Rule.Group
				msg.Resp <- nil
				continue
			}

			resp := VmGrp2ReqRule(appRule.grp)
			msg.Resp <- &comm.ListResp{Res: resp}
		case *comm.ErrReq:
			msg := data.(*comm.ErrReq)
			switch msg.Rule.Action {
			case proto.Rule_Update:
				// 新增组
				grp := ReqRule2VmGrp(msg.Rule)
				if err := grp.Validate(false, true); nil != err {
					msg.Resp <- err
					continue
				}

				appRule := p.apps.mApp[msg.Rule.Group]
				if nil == appRule{
					grp.File = filepath.Join(p.apps.dir, msg.Rule.Group+".rules")
					p.apps.mApp[msg.Rule.Group] = &appRules{
						grp:    grp,
						mRules: getRuleMap(grp.Rules),
					}

					msg.Resp <- p.apps.flush(msg.Rule.Group)
					continue
				}

				var rm ruleMap = map[uint64]*vmCfg.Rule{}
				var changed bool //防止不必要的写文件操作
				for _, v := range msg.Rule.Alert {
					// 更新组规则
					if 0 != v.Id {
						// 因为ruleService会尝试向所有节点更新某个App的告警规则，
						// 而该节点有可能并不存在该规则，此时也无需返回错误
						if _, ok := appRule.mRules[v.Id]; !ok {
							continue
						} else {
							// 存在则更新组规则
							newRule := updateRule(v)
							if !appRule.mRules.isDupRule(newRule){
								changed = true
								appRule.mRules[newRule.ID] = newRule
								delete(appRule.mRules, v.Id)
							}
						}
					}

					// 新增组规则
					if 0 == v.Id {
						vmRule := vmCfg.Rule{
							Type: datasource.NewPrometheusType(),
							Alert: v.Name,
							Expr:  v.Expr,
							Labels: kv2Map(v.Labels),
						}
						vmRule.ID = vmCfg.HashRule(vmRule)

						// 防止添加重复的告警规则
						if !appRule.mRules.isDupRule(&vmRule){
							changed = true
							rm[vmRule.ID] = &vmRule
						}
					}
				}

				for k,v := range rm{
					kk := k
					vv := v
					appRule.mRules[kk] = vv
				}

				if changed{
					appRule.grp.Rules = rm2Array(appRule.mRules)
					msg.Resp <- p.apps.flush(msg.Rule.Group)
				}else {
					msg.Resp <- nil
				}
			case proto.Rule_Remove:
				var changed bool
				appRule := p.apps.mApp[msg.Rule.Group]
				if nil == appRule {
					msg.Resp <- nil
					continue
				}

				if 0 == len(msg.Rule.Alert) { //todo:增加文档描述
					p.delGrp <- msg.Rule.Group
					os.Remove(appRule.grp.File)
					msg.Resp <- nil
					continue
				}

				for _, v := range msg.Rule.Alert {
					if _,ok := appRule.mRules[v.Id];ok{
						changed = true
						delete(appRule.mRules, v.Id)
						if 0 == len(appRule.mRules) {
							p.delGrp <- msg.Rule.Group
							os.Remove(appRule.grp.File)
							delete(p.apps.mApp, msg.Rule.Group)
							msg.Resp <- nil
							goto RemoveEnd
						}
					}
				}

				if changed{
					appRule.grp.Rules = rm2Array(appRule.mRules)
					msg.Resp <- p.apps.flush(msg.Rule.Group)
				}
			RemoveEnd:
				continue
			default:
				msg.Resp <- fmt.Errorf("invalid action:%d", msg.Rule.Action)
			}
		default:
			log.Print("invalid msg type")
		}
	}
}

func updateRule(update *proto.AlertDesc)*vmCfg.Rule{
	newRule := &vmCfg.Rule{
		Type: datasource.NewPrometheusType(),
		Alert: update.Name,
		Expr: update.Expr,
		Labels: make(map[string]string),
		Annotations: make(map[string]string),
	}
	newRule.Labels = kv2Map(update.Labels)
	newRule.Annotations = kv2Map(update.Annotation)
	newRule.ID = vmCfg.HashRule(*newRule)
	return newRule
}