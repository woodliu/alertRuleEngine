package svc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Handle interface {
	Add(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Remove(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
}

type Service map[string]*ClusterSvc

type ClusterSvc struct {
	nodeHash    []int              // 节点的哈希列表，可能存在相同的哈希值
	nodeMap     map[int]*Node      // 哈希值与节点的对应关系
	nodeClients []*grpc.ClientConn // 建立的grpc客户端列表
}

type Node struct {
	CurIndex int
	clients  []*grpc.ClientConn
}

// 用于轮询具有相同哈希值的节点
func (ns *Node) nextIndex() *grpc.ClientConn {
	clientLen := len(ns.clients)
	if 0 == clientLen {
		return nil
	}

	if ns.CurIndex < clientLen-1 {
		return ns.clients[ns.CurIndex+1]
	}

	return ns.clients[0]
}

func setKeepAliveConn(addr string) *grpc.ClientConn {
	var opts []grpc.DialOption
	var kaCp = keepalive.ClientParameters{
		Time:                time.Minute, // send pings every 10 seconds if there is no activity
		PermitWithoutStream: true,        // send pings even without active streams
	}

	opts = append(
		opts, grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kaCp),
	)

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("connect to v err:%s", err.Error())
	}

	return conn
}

func NewService(clusters []Cluster) Handle {
	var svc Service
	for _, c := range clusters {
		var clusterProxy ClusterSvc
		clusterProxy.nodeMap = make(map[int]*Node)

		for _, v := range c.Engines {
			hs := hash(v)
			if _, ok := clusterProxy.nodeMap[hs]; !ok {
				clusterProxy.nodeMap[hs] = &Node{}
			}

			conn := setKeepAliveConn(v)
			clusterProxy.nodeHash = append(clusterProxy.nodeHash, hs)
			clusterProxy.nodeMap[hs].clients = append(clusterProxy.nodeMap[hs].clients, conn)
			clusterProxy.nodeClients = append(clusterProxy.nodeClients, conn)
		}

		sort.Ints(clusterProxy.nodeHash)
		svc[c.Name] = &clusterProxy
	}

	return &svc
}

func (s Service) distributeCluster(r *http.Request) (*ClusterSvc, *proto.Rule, error) {
	var req proto.Rule
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &req)
	if nil != err {
		return nil, nil, err
	}

	cs := s[req.Type]
	if nil == cs {
		return nil, nil, errors.New("err alert type")
	}

	return cs, &req, nil
}

func setErrResp(statusCode int, err string, w http.ResponseWriter) {
	log.Print(err)
	b, _ := json.Marshal(&proto.ListResp{Err: err})
	w.WriteHeader(statusCode)
	w.Write(b)
}

// Add rcp 的连接错误指标
func (s Service) Add(w http.ResponseWriter, r *http.Request) {
	cs, req, err := s.distributeCluster(r)
	if nil != err {
		setErrResp(http.StatusBadRequest, err.Error(), w)
		return
	}

	resp,_,_,_ := s.getGrpFromCluster(cs, req)
	if nil != resp{
		setErrResp(http.StatusBadRequest, fmt.Sprintf("add exist group:%s",req.Group), w)
		return
	}

	index := BinarySearch(hash(req.Group), cs.nodeHash, 0, len(cs.nodeHash))
	conn := cs.nodeMap[cs.nodeHash[index]].nextIndex()

	c := proto.NewRequestClient(conn)
	_, err = c.Update(context.Background(), req)
	if nil != err {
		log.Println(err)
		setErrResp(http.StatusInternalServerError, err.Error(), w)
		return
	}

	return
}

func (s Service)getGrpFromCluster(cs *ClusterSvc, req *proto.Rule)(*proto.ListResp,error,error,*[]proto.RequestClient){
	var l sync.Mutex
	var rs []*proto.Rule
	var reqClits []proto.RequestClient
	var errHappen bool

	wg := sync.WaitGroup{}
	wg.Add(len(cs.nodeClients))

	for _, conn := range cs.nodeClients {
		c := conn
		go func(conn *grpc.ClientConn) {
			defer wg.Done()

			reqCli := proto.NewRequestClient(c)
			reply, err := reqCli.List(context.Background(), req)
			if nil != err {
				log.Print(err)
				errHappen = true
				return
			}

			if nil != reply {
				reqClits = append(reqClits, reqCli)
				l.Lock()
				rs = append(rs, reply.Res)
				l.Unlock()
			}
		}(conn)
	}

	wg.Wait()

	var connErr error
	if errHappen {
		connErr = fmt.Errorf("may partial data")
	}

	resp,err := merge(rs)
	return resp, err, connErr, &reqClits
}

func (s Service) Remove(w http.ResponseWriter, r *http.Request) {
	cs, req, err := s.distributeCluster(r)
	if nil != err {
		setErrResp(http.StatusBadRequest, err.Error(), w)
		return
	}

	var errHappen bool
	wg := sync.WaitGroup{}
	wg.Add(len(cs.nodeClients))

	for _, conn := range cs.nodeClients {
		go func(conn *grpc.ClientConn) {
			defer wg.Done()
			c := proto.NewRequestClient(conn)
			_, err := c.Remove(context.Background(), req)
			if nil != err {
				errHappen = true
				log.Println(err)
				return
			}
		}(conn)
	}

	wg.Wait()

	if errHappen {
		setErrResp(http.StatusBadRequest, "may partial removed", w)
		return
	}
	return
}

func (s Service) Update(w http.ResponseWriter, r *http.Request) {
	cs, req, err := s.distributeCluster(r)
	if nil != err {
		setErrResp(http.StatusBadRequest, err.Error(), w)
		return
	}

	resp,_,_, conns := s.getGrpFromCluster(cs, req)
	if nil != resp{
		setErrResp(http.StatusBadRequest, fmt.Sprintf("update non-existed group:%s",req.Group), w)
		return
	}

	if nil == conns {
		setErrResp(http.StatusInternalServerError, fmt.Sprintf("no available node"), w)
		return
	}

	_, err = (*conns)[0].Update(context.Background(), req)
	if nil != err {
		setErrResp(http.StatusBadRequest, err.Error(), w)
		return
	}

	return
}

func (s Service) List(w http.ResponseWriter, r *http.Request) {
	cs, req, err := s.distributeCluster(r)
	if nil != err {
		setErrResp(http.StatusBadRequest, err.Error(), w)
		return
	}

	resp,err,connErr,_ := s.getGrpFromCluster(cs, req)
	if nil != err{
		setErrResp(http.StatusInternalServerError, err.Error(), w)
		return
	}

	if nil == resp {
		return
	}

	resp.Err = connErr.Error()
	b, _ := json.Marshal(&resp)
	w.Write(b)
	return
}

func checkDup(r *proto.Rule)error{
	m := make(map[string]struct{})
	for _,v := range r.Alert{
		if _,ok := m[v.Name];!ok{
			m[v.Name] = struct{}{}
		}else {
			return fmt.Errorf("duplicate rule for group:%s, rule name:%s",r.Group,v.Name)
		}
	}

	return nil
}

func merge(rs []*proto.Rule) (*proto.ListResp,error) {
	var r proto.Rule

	if nil == rs || 0 == len(rs) {
		return nil,nil
	}

	for _, v := range rs {
		r.Alert = append(r.Alert, v.Alert...)
	}

	err := checkDup(&r)
	if nil != err{
		return nil,err
	}

	r.Group = rs[0].Group
	r.Type = rs[0].Type

	return &proto.ListResp{Res: &r},nil
}
