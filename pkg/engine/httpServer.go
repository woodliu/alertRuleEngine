package engine

import (
    "encoding/json"
    "github.com/woodliu/alertRuleEngine/pkg/common"
    tmpl "github.com/woodliu/alertRuleEngine/pkg/template"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
)

type httpServer struct {
    storage Storage
    reload chan struct{}
    ruleMap map[int64]tmpl.Rule
}

func ServeHttpRules(port int, storage Storage, reload chan struct{}, tmpl *tmpl.RuleTmpl){
    srv := http.Server{
        Addr: ":" + strconv.Itoa(port),
        Handler: &httpServer{
            storage: storage,
            reload: reload,
            ruleMap: tmpl.GenTmplRuleMap(),
        },
    }

    go func() {
        log.Printf("listen at 0.0.0.0:%d",port)
        if err := srv.ListenAndServe(); err != nil {
            log.Fatal(err.Error())
        }
    }()
}

const(
    ADD    = "/addAppRule"
    DELETE = "/deleteAppRule"
    EDIT   = "/editAppRule"
    GET    = "/getAppRule"
)

type HttpResp struct {
    data []common.AppRuleResp
    msg string
}

//TODO: v1/api group
func (h *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    body,err := ioutil.ReadAll(r.Body)
    if nil != err {
        log.Println(err)
        return
    }
    defer r.Body.Close()

    var rule common.AppRule
    err = json.Unmarshal(body, &rule)
    if nil != err {
        log.Println(err)
        return
    }

    //TODO：print request
    switch path {
    case ADD:
        if err := rule.Validate();nil != err{
            log.Println(err)
            return
        }

        err := h.storage.AddAppRule(&rule)
        if nil != err{
            writeResp(w, http.StatusInternalServerError, &HttpResp{msg: err.Error()})
            return
        }

        h.reload <- struct{}{}
        writeResp(w, http.StatusOK, nil)
    case DELETE:
        if err := h.storage.DeleteAppRuleByAppId(&rule);nil != err{
            writeResp(w, http.StatusInternalServerError, &HttpResp{msg: err.Error()})
            return
        }

        h.reload <- struct{}{}
        writeResp(w, http.StatusOK,nil)
    case EDIT:
        if err := h.storage.EditAppRuleByAppId(&rule);nil != err{
            writeResp(w, http.StatusInternalServerError, &HttpResp{msg: err.Error()})
            return
        }

        h.reload <- struct{}{}
        writeResp(w, http.StatusOK,nil)
    case GET:
       rules,err := h.storage.GetByAppId(&rule)
        if nil != err{
            writeResp(w, http.StatusInternalServerError, &HttpResp{msg: err.Error()})
            return
        }

        if nil == rules{
            writeResp(w, http.StatusOK, nil)
            return
        }

        var resp []common.AppRuleResp
        for _,v := range *rules{
            resp = append(resp,common.AppRuleResp{
                Env: v.Env,
                RuleId:v.RuleId,
                Expr: h.ruleMap[v.RuleId].Expr,
                Oper:v.Oper,
                Value: v.Value,
            })
        }
        writeResp(w, http.StatusOK, &HttpResp{data: resp})
    default:
        w.WriteHeader(http.StatusNotFound)
        return
    }
}

func writeResp(w http.ResponseWriter,statusCode int, resp *HttpResp){
    w.WriteHeader(statusCode)
    if nil != resp{
        bytes,_ := json.Marshal(resp)
        w.Write(bytes)
    }
}
