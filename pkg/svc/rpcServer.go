package svc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"github.com/woodliu/alertRuleEngine/pkg/comm"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
	"github.com/woodliu/alertRuleEngine/pkg/rule"
	"log"
	"sync"
	"time"
)

type Listener struct {
	handler *rule.Handler
	newP    func() *rule.Processor
	proto.UnsafeRequestServer
}

const rpcPort = 8888

func StartRpcServer(handler *rule.Handler, dir string, delGrpCh chan string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	enforcement := keepalive.EnforcementPolicy{
		MinTime:             time.Minute,
		PermitWithoutStream: true,
	}
	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(enforcement), // here
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: time.Minute,
		}),
	)
	proto.RegisterRequestServer(grpcServer, &Listener{
		handler: handler,
		newP:    rule.NewProcessorFunc(dir, delGrpCh),
	})
	grpcServer.Serve(lis)
}

var listReqPool = sync.Pool{
	New: func() interface{} {
		return &comm.ListReq{
			Resp: make(chan *comm.ListResp),
		}
	},
}

func (l *Listener) List(ctx context.Context, r *proto.Rule) (*proto.ListResp, error) {
	req := listReqPool.Get().(*comm.ListReq)
	defer listReqPool.Put(req)

	req.Rule = r
	req.Rule.Action = proto.Rule_List
	go l.handler.ListAppRules(req)

	res := <-req.Resp
	if nil == res {
		return &proto.ListResp{}, nil
	}

	if nil != res.Err {
		return &proto.ListResp{Res: res.Res, Err: res.Err.Error()}, res.Err
	}
	return &proto.ListResp{Res: res.Res}, nil
}

var errReqPool = sync.Pool{
	New: func() interface{} {
		return &comm.ErrReq{
			Resp: make(chan comm.ErrResp),
		}
	},
}

func (l *Listener) Update(ctx context.Context, r *proto.Rule) (*proto.ErrResp, error) {
	return l.changeRule(proto.Rule_Update,r)
}

func (l *Listener) Remove(ctx context.Context, r *proto.Rule) (*proto.ErrResp, error) {
	return l.changeRule(proto.Rule_Remove, r)
}

func (l *Listener) changeRule(ac proto.Rule_ActionType, r *proto.Rule) (*proto.ErrResp, error) {
	req := errReqPool.Get().(*comm.ErrReq)
	defer errReqPool.Put(req)

	req.Rule = r
	req.Rule.Action = ac
	go l.handler.ModifyRules(req, l.newP)

	err := <-req.Resp
	if nil != err{
		return &proto.ErrResp{Err: err.Error()},err
	}
	return new(proto.ErrResp), err
}