package rule

import (
	vmCfg "github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/config"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/datasource"
	"os"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
)

type Rule struct {
	dir  string
	mApp map[string]*appRules
}

type ruleMap map[uint64]*vmCfg.Rule
type appRules struct {
	grp    *vmCfg.Group
	mRules ruleMap
}

func (rm ruleMap)isDupRule(r *vmCfg.Rule)bool{
	if _, ok := rm[r.ID]; !ok {
		return false
	}
	return true
}

func getRuleMap(rules []vmCfg.Rule) map[uint64]*vmCfg.Rule {
	m := make(map[uint64]*vmCfg.Rule)
	for _, v := range rules {
		vv := v
		m[vv.ID] = &vv
	}

	return m
}

func (r *Rule) flush(grp string) error {
	data, err := vmGrp2File(r.mApp[grp].grp)
	if nil != err {
		return err
	}

	tmpFile := r.mApp[grp].grp.File + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if nil != err {
		return err
	}

	if _, err := f.Write(data); nil != err {
		f.Close()
		os.Remove(tmpFile)
		return err
	}

	f.Close()
	if err := os.Rename(tmpFile, r.mApp[grp].grp.File); nil != err {
		return err
	}

	return nil
}

func kv2Map(labels []*proto.Label)map[string]string{
	m := make(map[string]string)
	for _,v := range labels{
		m[v.Name] = v.Value
	}

	return m
}

func ReqRule2VmGrp(reqRule *proto.Rule) *vmCfg.Group {
	var vmRules []vmCfg.Rule
	for _, v := range reqRule.Alert {
		vmRule := vmCfg.Rule{
			Type: datasource.NewPrometheusType(),
			Alert: v.Name,
			Expr:  v.Expr,
			Labels: kv2Map(v.Labels),
		}
		vmRule.ID = vmCfg.HashRule(vmRule)
		vmRules = append(vmRules, vmRule)
	}

	return &vmCfg.Group{
		Name:  reqRule.Group,
		Rules: vmRules,
	}
}

func VmGrp2ReqRule(grp *vmCfg.Group) *proto.Rule {
	var resp proto.Rule
	resp.Group = grp.Name
	for _, v := range grp.Rules {
		resp.Alert = append(resp.Alert, &proto.AlertDesc{
			Id:   v.ID,
			Name: v.Name(),
			Expr: v.Expr,
		})
	}

	return &resp
}

func VmRule2AppRule(rs *[]vmCfg.Rule) ruleMap {
	m := make(map[uint64]*vmCfg.Rule)
	for _, v := range *rs {
		vv := v
		m[vv.ID] = &vv
	}
	return m
}

func rm2Array(r ruleMap)[]vmCfg.Rule{
	var rs []vmCfg.Rule

	for _,v := range r{
		vv := v
		rs = append(rs, *vv)
	}

	return rs
}