package rule

import (
	vmCfg "github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/config"
	"github.com/woodliu/alertRuleEngine/pkg/comm"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
)

func NewHandler(dir,path string, delGrp chan string) (*Handler, error) {
	groups, err := vmCfg.Parse([]string{path}, false, true)
	if nil != err {
		return nil, err
	}

	return getHandler(dir, groups, delGrp), nil
}

func (h *Handler) Run() {
	for _, v := range h.ps {
		go v.Exec()
	}

	go func() {
		for grpName := range h.delGrp {
			h.l.Lock()
			if p, ok := h.mapP[grpName]; ok {
				delete(h.mapP, grpName)
				if 0 == len(p.apps.mApp) {
					close(p.ruleCh)
				}
			}
			h.l.Unlock()
		}
	}()
}

func (h *Handler) ListAppRules(r *comm.ListReq) {
	p := h.getGrpProcessor(r.Rule.Group)
	if nil == p {
		r.Resp <- nil
		return
	}
	p.ruleCh <- r
}

func (h *Handler) ModifyRules(r *comm.ErrReq, new func() *Processor) {
	p := h.getGrpProcessor(r.Rule.Group)
	if nil == p {
		// 新增组
		if proto.Rule_Update == r.Rule.Action {
			avaP, isNew := h.getAvailableProc(new)
			if isNew {
				h.l.Lock()
				h.ps = append(h.ps, avaP)
				h.l.Unlock()
			}
			h.l.Lock()
			h.mapP[r.Rule.Group] = avaP
			h.l.Unlock()
			avaP.ruleCh <- r
			return
		} else {
			r.Resp <- nil
			return
		}
	}
	p.ruleCh <- r
}

const maxGrpPerProcessor = 100

func getHandler(dir string, groups []vmCfg.Group, delGrp chan string) *Handler {
	var pNum int
	grpNum := len(groups)

	if grpNum%maxGrpPerProcessor > 0 {
		pNum = grpNum/maxGrpPerProcessor + 1
	} else {
		pNum = grpNum / maxGrpPerProcessor
	}

	handler := &Handler{
		mapP:   make(map[string]*Processor),
		delGrp: delGrp,
	}

	getRule := func(groups []vmCfg.Group) *Rule {
		r := &Rule{
			dir:  dir,
			mApp: map[string]*appRules{},
		}
		for _, v := range groups {
			vv := v
			r.mApp[vv.Name] = &appRules{
				grp:    &vv,
				mRules: VmRule2AppRule(&vv.Rules),
			}
		}
		return r
	}

	if 1 == pNum {
		p := &Processor{
			ruleCh: make(chan interface{}),
			delGrp: delGrp,
			apps:   getRule(groups),
		}

		handler.ps = append(handler.ps, p)
		for _, v := range groups {
			handler.mapP[v.Name] = p
		}
		return handler
	}

	for i := 0; i < pNum; i++ {
		p := &Processor{
			ruleCh: make(chan interface{}),
			delGrp: delGrp,
		}
		handler.ps = append(handler.ps, p)
		if i != pNum-1 {
			pGroups := groups[(maxGrpPerProcessor * i):(maxGrpPerProcessor * (i + 1))]
			p.apps = getRule(pGroups)
			for _, v := range pGroups {
				handler.mapP[v.Name] = p
			}
		} else {
			pGroups := groups[(maxGrpPerProcessor * i):]
			p.apps = getRule(pGroups)
			for _, v := range pGroups {
				handler.mapP[v.Name] = p
				return handler
			}
		}
	}

	return handler
}
