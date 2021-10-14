package engine

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/woodliu/alertRuleEngine/pkg/common"
)

type Storage interface{
	AddAppRule(rule *common.AppRule) error
	GetByAppId(rule *common.AppRule) (*[]common.AppRule,error)
	EditAppRuleByAppId(rule *common.AppRule) error
	DeleteAppRuleByAppId(rule *common.AppRule) error //ruleId is 0,delete all; ruleId>0,delete the select one

	// above methods won't expose to app user,just for operators
	GetExprByRuleId(rule *common.AppRule) error
	EditExprByRuleId(rule *common.AppRule) error
	DeleteByRuleId(rule *common.AppRule) error

	GetAllAppRules()(*[]common.AppRule,error)
	StoreAlerts(rule *common.AppRule) error //send?
}

type Store struct {
	Db *gorm.DB
}

func NewStore(db *gorm.DB)Storage{
	db.AutoMigrate(&common.AppRule{})
	return &Store{
		Db: db,
	}
}

func (store * Store)AddAppRule(rule *common.AppRule) error{
	if exist,err := store.appRuleExist(rule);nil != err{
		return err
	}else if false == exist{
		return fmt.Errorf("appId:%s ruleId:%d exist",rule.AppId, rule.RuleId)
	}

	return store.Db.Model(&common.AppRule{}).Create(rule).Error
}

func (store * Store)GetByAppId(rule *common.AppRule) (*[]common.AppRule,error){
	var rules []common.AppRule
	if err := store.Db.Model(&common.AppRule{}).Where("`env` = ?",rule.Env).Where("`app_id` = ?",rule.AppId).Scan(&rules).Error;nil != err{
		return nil, err
	}

	return &rules,nil
}

func (store * Store)EditAppRuleByAppId(rule *common.AppRule) error{
	if exist,err := store.appRuleExist(rule);nil != err{
		return err
	}else if !exist{
		return fmt.Errorf("appId:%s ruleId:%d not exist",rule.AppId, rule.RuleId)
	}

	return store.Db.Model(&common.AppRule{}).Where("`env` = ?",rule.Env).Where("`app_id` = ?",rule.AppId).Where("`rule_id` = ?",rule.RuleId).Updates(rule).Error
}

func (store * Store)DeleteAppRuleByAppId(rule *common.AppRule) error{
	if 0 == rule.RuleId{
		return store.Db.Model(&common.AppRule{}).Where("`env` = ?",rule.Env).Where("`app_id` = ?",rule.AppId).Delete(&common.AppRule{}).Error
	}else{
		return store.Db.Model(&common.AppRule{}).Where("`env` = ?",rule.Env).Where("`app_id` = ?",rule.AppId).Where("`rule_id` = ?",rule.RuleId).Delete(&common.AppRule{}).Error
	}
}

func (store * Store)GetExprByRuleId(rule *common.AppRule) error{
	return nil
}
func (store * Store)EditExprByRuleId(rule *common.AppRule) error{
	return nil
}
func (store * Store)DeleteByRuleId(rule *common.AppRule) error{
	return nil
}

func (store * Store)StoreAlerts(rule *common.AppRule) error{
	return nil
}

func (store * Store)GetAllAppRules()(*[]common.AppRule,error){
	var rules []common.AppRule
	if err := store.Db.Model(&common.AppRule{}).Scan(&rules).Error;nil != err{
		return nil, err
	}
	return &rules,nil
}

func (store *Store)appRuleExist(rule *common.AppRule)(bool,error){
	var num int64

	if err := store.Db.Model(&common.AppRule{}).Where("`env` = ?",rule.Env).Where("`app_id` = ?",rule.AppId).Where("`rule_id` = ?",rule.RuleId).Count(&num).Error;nil != err{
		return false, err
	}
	return num ==0, nil
}