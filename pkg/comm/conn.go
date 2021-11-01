package comm

import "github.com/woodliu/alertRuleEngine/pkg/proto"

type ListResp struct {
    Err error
    Res *proto.Rule
}

type ListReq struct {
    Rule *proto.Rule
    Resp chan *ListResp
}

type ErrResp error
type ErrReq struct {
    Rule *proto.Rule
    Resp chan ErrResp
}

