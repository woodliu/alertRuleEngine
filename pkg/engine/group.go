package engine

import (
    "context"
    "github.com/woodliu/alertRuleEngine/pkg/common"
    "github.com/woodliu/alertRuleEngine/pkg/util"
    "log"
    "time"
)

const(
    QueryInterval     = time.Second * 10
    GroupWaitPoll     = time.Second * 15
    GroupIntervalPoll = time.Second * 30
    GroupRepeatPoll   = time.Minute
)

type group struct {
    rules []common.AppRule
}

type groupChains struct {
    e *Engine

    reload <- chan struct{}
    query chan <- *common.AlertReq    // query prometheus
    receive <- chan *common.AlertResp // receive alert response
    send chan <- interface{}  // send alert

    queryChan,waitChan,intervalChan,repeatChan chan *util.Node

    queryChain *util.Chain
    waitChain *util.Chain
    intervalChain *util.Chain
    repeatChain *util.Chain

    gWaitTimeout time.Duration
    gIntervalTimeout time.Duration
    gRepeatTimeout time.Duration
}

func NewGroupChain(e *Engine,query chan *common.AlertReq,receive chan *common.AlertResp, send chan interface{}, reload chan struct{}) *groupChains{
    return &groupChains{
        e: e,

        reload: reload,
        query: query,
        send: send,
        receive: receive,

        queryChan: make(chan *util.Node),
        waitChan: make(chan *util.Node),
        intervalChan: make(chan *util.Node),
        repeatChan: make(chan *util.Node),

        queryChain: util.InitChain(),
        waitChain: util.InitChain(),
        intervalChain: util.InitChain(),
        repeatChain: util.InitChain(),
    }
}

func (gChain *groupChains)Run(ctx context.Context)error{
    gChain.gWaitTimeout = gChain.e.groupConf.GroupWait
    gChain.gIntervalTimeout = gChain.e.groupConf.GroupInterval
    gChain.gRepeatTimeout = gChain.e.groupConf.RepeatInterval

    go gChain.runQuery(ctx)
    go gChain.receiveResp(ctx)

    go gChain.runPollQueryChain(ctx)
    go gChain.runPollWaitChain(ctx)
    go gChain.runPollRepeatChain(ctx)

    return nil
}

func (gChain *groupChains)runQuery(ctx context.Context){
    queryTicker := time.NewTicker(QueryInterval)
    defer queryTicker.Stop()

    for{
        select {
        case <- queryTicker.C:
            for appId,group := range gChain.e.groupMap{
                for _,appRule := range group.rules{
                    if rule,ok := gChain.e.ruleMap[appRule.RuleId];!ok{
                        log.Warnf("appId：%s has no ruleId:%d",appId, appRule.RuleId)
                        continue
                    }else {
                        gChain.query <- &common.AlertReq{AppRule: appRule, Rule: rule}
                    }
                }
            }
        case <- gChain.reload:
            gChain.e.LoadRules()
        case <-ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)receiveResp(ctx context.Context){
    for {
        select {
        case resp := <- gChain.receive:
            if nil == resp{
                continue
            }

            if resp.ActiveAt.IsZero(){
                resp.ActiveAt = time.Now()
            }

            forDuration,_ := time.ParseDuration(gChain.e.ruleMap[resp.RuleId].For)
            // Need to re-query
            if time.Now().Sub(resp.ActiveAt) < forDuration{
                gChain.queryChan <- util.NewNode(resp.AppId, gChain.e.groupConf.GroupWait, &[]common.AlertResp{*resp})
            // Or just add to waitChain
            }else{
                gChain.waitChan <- util.NewNode(resp.AppId, gChain.gWaitTimeout, &[]common.AlertResp{*resp})
            }
        case <- ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)runPollQueryChain(ctx context.Context){
    queryTicker  := time.NewTicker(QueryInterval)
    defer queryTicker.Stop()

    for{
        select {
        case nodes := <- gChain.queryChan:
            gChain.queryChain.AddNodes(nodes, updateNodeData)
        case <- queryTicker.C:
            // Add timeout nodes to waitChain, and do re-query
            nodes := gChain.queryChain.Poll(QueryInterval, isCut)
            if nil == nodes{
                continue
            }
            gChain.waitChan <- nodes
            gChain.queryChain.ExecNodeData(gChain.sendAlertResp)
        case <-ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)runPollWaitChain(ctx context.Context){
    waitTicker  := time.NewTicker(GroupWaitPoll)
    defer waitTicker.Stop()

    for {
        select {
        case nodes := <-gChain.waitChan:
            setFiredAt(nodes)
            setNodeDuration(nodes,gChain.gWaitTimeout)
            gChain.waitChain.AddNodes(nodes,updateNodeData)
        case <- waitTicker.C:
            nodes := gChain.waitChain.Poll(GroupWaitPoll,isCut)
            if nil == nodes{
                continue
            }
            gChain.intervalChan <- nodes
        case <- ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)runPollIntervalChain(ctx context.Context){
    waitTicker  := time.NewTicker(GroupIntervalPoll)
    defer waitTicker.Stop()

    for {
        select {
        case nodes := <- gChain.intervalChan:
            setNodeDuration(nodes, gChain.gIntervalTimeout)
            gChain.intervalChain.AddNodes(nodes, updateNodeData)

            // Get new added groups,and add to repeatChain
            newGroups := gChain.intervalChain.CopyNodes(false,copyAlerts)
            if nil == newGroups{
                continue
            }
            gChain.repeatChan <- newGroups
        case <- waitTicker.C:
            nodes := gChain.intervalChain.Poll(GroupIntervalPoll,isCut)
            if nil == nodes{
                continue
            }
            gChain.repeatChan <- nodes
        case <- ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)runPollRepeatChain(ctx context.Context){
    waitTicker  := time.NewTicker(GroupRepeatPoll)
    defer waitTicker.Stop()

    for {
        select {
        case nodes := <-gChain.repeatChan:
            setNodeDuration(nodes, gChain.gRepeatTimeout)
            gChain.repeatChain.AddNodes(nodes, updateNodeData)

            //The new and repeat data should be send immediately
            newGroups := gChain.repeatChain.CopyNodes(true, copyAlerts)
            newGroups.ExecNodeData(gChain.sendAlertResp)
        case <- waitTicker.C:
            // repeat timeout,send the data
            nodes := gChain.intervalChain.Poll(GroupRepeatPoll,isCut)
            nodes.ExecNodeData(gChain.sendAlertResp)
        case <- ctx.Done():
            return
        }
    }
}

func (gChain *groupChains)sendAlertResp(data interface{}){
    resp := common.UnMarshalRespData(data)
    if nil == resp{
        return
    }

    for _,v := range *resp{
        v.LastSentAt = time.Now()
    }

    gChain.send <- resp
}

func updateNodeData(curNode,newNode *util.Node){
    if nil == curNode || nil == newNode{
        return
    }

    curNodeData := common.UnMarshalRespData(curNode.Data)
    newNodeData := common.UnMarshalRespData(newNode.Data)

    curNodeDataMap := make(map[int64]common.AlertResp)
    newNodeDataMap := make(map[int64]common.AlertResp)

    for _,v := range *curNodeData{
        curNodeDataMap[v.RuleId] = v
    }

    for _,v := range *newNodeData{
        newNodeDataMap[v.RuleId] = v
    }

    for ruleId,data := range newNodeDataMap{
        if _,ok := curNodeDataMap[ruleId];!ok{
            curNode.Repeat = true
        }
        curNodeDataMap[ruleId] = data
    }

    var resDatas []common.AlertResp
    for _,v := range curNodeDataMap{
        resDatas = append(resDatas,v)
    }

    curNode.Data = &resDatas
}

func isCut(node *util.Node)bool{
   return  node.Duration <= 0
}

func copyAlerts(alerts interface{})interface{}{
    alertResp,_ := alerts.(*[]common.AlertResp)
    newAlertResp := make([]common.AlertResp,len(*alertResp))

    copy := func(m map[string]string) map[string]string{
        copyMap := make(map[string]string)
        for k,v := range m{
            copyMap[k] = v
        }
        return copyMap
    }

    for k,v := range *alertResp {
        newAlertResp[k] = v
        newAlertResp[k].Labels = copy(v.Labels)
        newAlertResp[k].Annotations = copy(v.Annotations)
    }

    return &newAlertResp
}

func setNodeDuration(node *util.Node,timeout time.Duration){
    for curNode := node; nil != curNode;curNode = curNode.Next{
        curNode.Duration = timeout
    }
}

func setFiredAt(node *util.Node){
    for curNode := node; nil != curNode;curNode = curNode.Next{
        resp := common.UnMarshalRespData(curNode.Data)
        if nil == resp{
            continue
        }

        for _,v := range *resp{
            v.FiredAt = time.Now()
        }
    }
}