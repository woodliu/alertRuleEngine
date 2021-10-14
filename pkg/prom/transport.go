package prom

import (
    "context"
    "fmt"
    "net"
    "net/http"
    "github.com/woodliu/alertRuleEngine/pkg/common"
    "time"
)

type Trans interface {
    Send(ctx context.Context)
}

type AlertChannel struct {
    send <- chan interface{}
    trans http.RoundTripper
    alertCount int64  //TODO:delete
}

func NewTransport() http.RoundTripper{
    return &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: 10 * time.Second,
            // value taken from http.DefaultTransport
            KeepAlive: 30 * time.Second,
        }).DialContext,
    }
}

func NewTrans(send chan interface{})Trans{
    return &AlertChannel{
        send:send,
        trans: NewTransport(),
    }
}

func (ac *AlertChannel) Send(ctx context.Context) {
    //go ac.testPerformance()
    go func() {
        for {
            select {
            case data := <- ac.send:
                ac.alertCount ++
                alerts := common.UnMarshalRespData(data)
                if nil != alerts{
                    // TODO：send to alerts channel
                    req, _ := http.NewRequest("GET", "", nil)
                    ac.trans.RoundTrip(req)
                }

            case <-ctx.Done():
                return
            }
        }
    }()
}

func (ac *AlertChannel)testPerformance(){
    ticker := time.NewTicker(time.Second)
    for {
        select {
        case <- ticker.C:
            lastAlertCount := ac.alertCount
            fmt.Println("process alerts number:", ac.alertCount - lastAlertCount)
        }
    }
}