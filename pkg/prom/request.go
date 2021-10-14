package prom


import (
    "context"
    "fmt"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
    "golang.org/x/time/rate"
    "github.com/woodliu/alertRuleEngine/pkg/common"
    "github.com/woodliu/alertRuleEngine/pkg/config"
    tmpl "github.com/woodliu/alertRuleEngine/pkg/template"
    "log"
    "strings"
    "time"
    "github.com/prometheus/client_golang/api"
)

type Prom struct {
    limiter *rate.Limiter
    promMap map[int64]string
    ruleMap map[int64]tmpl.Rule

    query   <-chan *common.AlertReq
    receive chan <- *common.AlertResp
}

func NewProm(ruleMap map[int64]tmpl.Rule, promConf *config.Conf, query <-chan *common.AlertReq, receive chan <- *common.AlertResp)*Prom{
    promMap := make(map[int64]string)
    for _,v := range promConf.PromConf.Prometheus{
        promMap[v.Env] = v.Addr
    }

    return &Prom{
        limiter: rate.NewLimiter(promConf.PromConf.Limiter.Limit, promConf.PromConf.Limiter.Burst),
        promMap: promMap,
        ruleMap: ruleMap,
        query: query,
        receive: receive,
    }
}

func (p *Prom)Run(ctx context.Context){
    go func() {
        for {
            select {
            case query := <- p.query:
                alertResp,err := p.queryProm(ctx, *query)
                if nil != err{
                    log.Error(err)
                    continue
                }
                p.receive <- alertResp
            case <- ctx.Done():
                return
            }
        }
    }()
}

func (p *Prom)queryProm(ctx context.Context, query common.AlertReq)(*common.AlertResp,error){
    var alertResp common.AlertResp

    queryExpr := query.Rule.Expr + query.AppRule.Oper + query.AppRule.Value
    promAddr := p.promMap[query.AppRule.Env]
    if client, err := api.NewClient(api.Config{Address: promAddr});nil != err{
        return nil, fmt.Errorf("Creating prom client: %v", err)
    }else{
        // use limiter
        if err := p.limiter.Wait(ctx);nil != err{
            log.Error(err)
        }

        api := v1.NewAPI(client)
        val, warnings, err := api.Query(ctx, queryExpr, time.Now())
        if nil != err{
            log.Errorf(err.Error())
        }

        if 0 < len(warnings){
            log.Warn(strings.Join(warnings," "))
        }

        switch val.Type() {
        case model.ValVector:
            resp := val.(model.Vector)
            if 1 < resp.Len() {  //TODO:make sure there is only one alert response for a service
                return nil, fmt.Errorf("multi prometheus response for %s at %d",queryExpr, time.Now().Unix())
            }else if 0 == resp.Len(){ // not trigger alert
                return nil,nil
            }

            alertResp.Result = float64(resp[0].Value)
        default:
            return nil, fmt.Errorf("prometheus response type:%d err",val.Type())
        }
    }

    alertResp.AppId = query.AppRule.AppId
    alertResp.RuleId = query.Rule.ID
    alertResp.Name = query.Rule.Name
    alertResp.Labels = p.ruleMap[query.Rule.ID].Labels
    alertResp.Annotations = p.ruleMap[query.Rule.ID].Annotations

    return &alertResp, nil
}
