package common

import (
	tmpl "github.com/woodliu/alertRuleEngine/pkg/template"
	"time"
)

type AlertReq struct {
	AppRule AppRule
	Rule    tmpl.Rule
}

type AlertResp struct{
	AppId string
	RuleId int64

	Name string
	Labels map[string]string
	Annotations map[string]string

	Result float64
	ActiveAt   time.Time // the first when alert active
	FiredAt    time.Time // The time when become a alert
	LastSentAt time.Time
}


func UnMarshalRespData(data interface{})*[]AlertResp{
	if dataTmp,ok := data.(*[]AlertResp);!ok{
		return nil
	}else {
		return dataTmp
	}
}
