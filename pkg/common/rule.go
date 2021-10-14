package common

import (
	"fmt"
	"strconv"
)

const(
	ALERT_LEVEL_INFO = iota + 1
	ALERT_LEVEL_WARN
	ALERT_LEVEL_HIGH
	ALERT_LEVEL_SEVERITY
)

type AppRule struct {
	ID     int64  `json:"-"      ,gorm:"primary_key"`
	Env    int64  `json:"env"    ,gorm:"column:env;varchar(64);not null"`
	AppId  string `json:"appId"  ,gorm:"column:appId;varchar(64);not null"`
	RuleId int64  `json:"ruleId" ,gorm:"column:ruleId;not null"`
	Oper   string `json:"oper"   ,gorm:"column:oper;varchar(64);not null"`
	Value  string `json:"value"  ,gorm:"column:value;varchar(64);not null"`
}

type AppRuleResp struct {
	Env    int64  `json:"env"`
	RuleId int64  `json:"ruleId"`
	Expr   string `json:"expr"`
	Oper   string `json:"oper"`
	Value  string `json:"value"`
}

var OperMap = map[string]struct{}{
	">": {},
	"<": {},
	"==": {},
	"!=": {},
	">=": {},
	"<=": {},
}

func (rule * AppRule)Validate()error{
	if _,ok := OperMap[rule.Oper];!ok{
		return fmt.Errorf("oper %s error",rule.Oper)
	}

	if _,err := strconv.ParseFloat(rule.Value,64);nil != err{
		return fmt.Errorf("value %s should be number",rule.Value)
	}

	return nil
}