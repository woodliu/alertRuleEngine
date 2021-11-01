package svc

import (
	"context"
	"fmt"
	"github.com/woodliu/alertRuleEngine/pkg/proto"
	"strconv"
	"testing"
)

func TestBinarySearch0(t *testing.T) {
	var arr1 []int

	if BinarySearch(60, arr1, 0, len(arr1)) != -1 {
		t.Fail()
	}
}

func TestBinarySearch1(t *testing.T) {
	arr1 := []int{100}

	if BinarySearch(60, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(100, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(200, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}
}

func TestBinarySearch2(t *testing.T) {
	arr1 := []int{100, 200}

	if BinarySearch(60, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(100, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(150, arr1, 0, len(arr1)) != 1 {
		t.Fail()
	}

	if BinarySearch(200, arr1, 0, len(arr1)) != 1 {
		t.Fail()
	}

	if BinarySearch(250, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}
}

func TestBinarySearch3(t *testing.T) {
	arr1 := []int{100, 200, 300, 400, 500}

	if BinarySearch(60, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(100, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}

	if BinarySearch(120, arr1, 0, len(arr1)) != 1 {
		t.Fail()
	}

	if BinarySearch(200, arr1, 0, len(arr1)) != 1 {
		t.Fail()
	}

	if BinarySearch(480, arr1, 0, len(arr1)) != 4 {
		t.Fail()
	}

	if BinarySearch(500, arr1, 0, len(arr1)) != 4 {
		t.Fail()
	}

	if BinarySearch(660, arr1, 0, len(arr1)) != 0 {
		t.Fail()
	}
}

func TestBinarySearch4(t *testing.T) {
	arr1 := []int{100, 200, 200, 200, 600}

	res := BinarySearch(200, arr1, 0, len(arr1))
	if res != 3 {
		t.Fail()
	}
}

func TestRpcListAppRules(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)
	fmt.Println("111", conn.GetState())
	// 不存在的app
	for i := 1; i < 2; i++ {
		req1 := &proto.Rule{
			Group: "appid111",
		}

		reply1, err := c.List(context.Background(), req1)
		if nil != err {
			t.Fail()
		}
		fmt.Println("reply1", reply1, err)
	}

	// 存在的app
	for i := 1; i < 3; i++ {
		req2 := &proto.Rule{
			Group: "appid3",
		}

		reply2, err := c.List(context.Background(), req2)
		if nil != err {
			t.Fail()
		}
		fmt.Println("reply2", reply2, err)
	}
}

func TestRpcAddAppRules1(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	req1 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid3",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
		},
	}

	// 首次添加，正常
	reply1, err := c.Update(context.Background(), req1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("reply1", reply1)

	// 重复添加，返回错误
	reply11, err := c.Update(context.Background(), req1)
	if nil == err {
		t.Fail()
	}
	fmt.Println("reply1", reply11)

	//查看添加的规则
	lreq1 := &proto.Rule{
		Group: "appid3",
	}

	lreply1, err := c.List(context.Background(), lreq1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("lreply1", lreply1, err)

	// 错误表达式的告警规则
	req2 := &proto.Rule{
	   Action: proto.Rule_Update,
	   Group: "appid4",
	   App: "app4",
	   Alert: []*proto.AlertDesc{
	       {
	           Name: "added alert 1",
	           Expr: "sum(((((((up)>100",
	       },
	   },
	}
	fmt.Println("start replay2")
	reply2,err := c.Update(context.Background(), req2)
	if nil == err{
	   t.Fail()
	}
	fmt.Println(reply2,err)
}

func TestRpcAddAppRules2(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	req1 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid3",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
		},
	}

	for i:=0;i<100;i++{
		req1.Group = "appid3"
		req1.Group += strconv.Itoa(i)
		reply1, err := c.Update(context.Background(), req1)
		if nil != err {
			t.Fail()
		}
		fmt.Println("reply1", reply1)
	}
}

func TestRpcUpdateAppRules1(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	// 更新不存在的应用规则,rpc不支持这种场景，会直接添加该应用规则
	req1 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "unknown",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
		},
	}

	//reply1, err := c.Update(context.Background(), req1) //此处不能多次执行，注意删除生成的文件
	//if nil != err {
	//	t.Fail()
	//}
	//fmt.Println("reply1", reply1,err)

	// 查询存在的appid2
	req1.Group = "appid2"
	lreply1, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("lreply1", lreply1, err)

	req2 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid2",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Id: 1000, // 更新应用不存在的规则
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
			{
				Id: lreply1.Res.Alert[1].Id, // 更新应用存在的规则
				Name: "update-fffff",
				Expr: "sum(up)>116611",
			},
		},
	}

	reply2, err := c.Update(context.Background(), req2)
	if nil != err {
		t.Fail()
	}
	fmt.Println("reply2", reply2,err)
}


func TestRpcUpdateAppRules2(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	req1 := &proto.Rule{Action: proto.Rule_Update}

	// 查询存在的appid2
	req1.Group = "appid2"
	lreply1, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}

	oldRuleNum := len(lreply1.Res.Alert)

	req2 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid2",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Id: 0, // 添加新规则
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
			{
				Id: 0, // 重复新增新规则
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
			{
				Id: 0, // 添加新规则
				Name: "added alert 2",
				Expr: "sum(up)>1000",
			},
			{
				Id: lreply1.Res.Alert[1].Id, // 更新应用存在的规则
				Name: "update-6666",
				Expr: "sum(up)>999999",
			},
		},
	}

	reply2, err := c.Update(context.Background(), req2)
	if nil == err {
		t.Fail()
	}
	fmt.Println("reply2", reply2,err)

	// 因为存在重复的规则，无法更新
	lreply2, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}

	if oldRuleNum != len(lreply2.Res.Alert){
		t.Fail()
	}

	req3 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid2",
		App:    "app3",
		Alert: []*proto.AlertDesc{
			{
				Id: 0, // 添加新规则
				Name: "added alert 1",
				Expr: "sum(up)>100",
			},
		},
	}

	reply3, err := c.Update(context.Background(), req3)
	if nil != err {
		t.Fail()
	}
	fmt.Println("reply3", reply3,err)
}

func TestRpcUpdateAppRules3(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	req1 := &proto.Rule{Action: proto.Rule_Update}

	// 查询存在的appid2
	req1.Group = "appid2"
	lreply1, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}

	fmt.Println("lreply1",lreply1)
	//req2 := &proto.Rule{
	//	Action: proto.Rule_Update,
	//	Group:  "appid2",
	//	App:    "app2",
	//	Alert: []*proto.AlertDesc{
	//		{
	//			Id: 0, // 添加新规则
	//			Name: "added alert 3",
	//			Expr: "sum(up)>10000",
	//		},
	//		{
	//			Id: lreply1.Res.Alert[0].Id, // 更新应用存在的规则
	//			Name: "update-001",
	//			Expr: "sum(up)>111111",
	//		},
	//	},
	//}
	//
	//reply2, err := c.Update(context.Background(), req2)
	//if nil != err {
	//	t.Fail()
	//}
	//fmt.Println("reply2", reply2,err)


	lreply2, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("lreply2",lreply2)
	req3 := &proto.Rule{
		Action: proto.Rule_Update,
		Group:  "appid2",
		App:    "app2",
		Alert: []*proto.AlertDesc{
			{
				Id: 10, // 添加新规则
				Name: "added alert 4",
				Expr: "sum(up)>10000",
			},
			{
				Id: 0, // 添加新规则
				Name: "added alert 6",
				Expr: "sum(up)>600000",
			},
			{
				Id: lreply2.Res.Alert[0].Id, // 更新应用存在的规则
				Name: "update-003",
				Expr: "sum(up)>2330000",
			},
		},
	}

	reply3, err := c.Update(context.Background(), req3)
	if nil != err {
		t.Fail()
	}
	fmt.Println("reply3", reply3,err)

	lreply3, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("lreply3",lreply3)
}


func TestRpcUpdateRemoveRules1(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	//删除不存在的应用告警
	req1 := &proto.Rule{
		Action: proto.Rule_Remove,
		Group:  "unexist",
	}
	reply1, err := c.Remove(context.Background(), req1)
	if nil != err {
		t.Fail()
	}

	fmt.Println("reply1", reply1)

	// 删除应用不存在的告警
	req2 := &proto.Rule{
		Action: proto.Rule_Remove,
		Group:  "appid2",
		Alert: []*proto.AlertDesc{
			{
				Id: 12345,
			},
		},
	}
	reply2, err := c.Remove(context.Background(), req2)
	if nil != err {
		t.Fail()
	}

	fmt.Println("reply2", reply2)
}


func TestRpcUpdateRemoveRules2(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	req1 := &proto.Rule{
		Group: "appid2",
	}

	lreply1, err := c.List(context.Background(), req1)
	if nil != err {
		t.Fail()
	}
	fmt.Println("lreply1", lreply1, err)

	// 删除应用的部分告警
	req2 := &proto.Rule{
		Action: proto.Rule_Remove,
		Group:  "appid2",
		Alert: []*proto.AlertDesc{
			{
				Id: lreply1.Res.Alert[0].Id,
			},
		},
	}
	reply2, err := c.Remove(context.Background(), req2)
	if nil != err {
		t.Fail()
	}

	fmt.Println("reply2", reply2)
}

func TestRpcUpdateRemoveRules3(t *testing.T) {
	conn := setKeepAliveConn("localhost:8888")
	defer conn.Close()
	c := proto.NewRequestClient(conn)

	// 删除应用告警
	req1 := &proto.Rule{
		Action: proto.Rule_Remove,
		Group:  "appid2",
	}
	for i:=0;i<3;i++{
		reply1, err := c.Remove(context.Background(), req1)
		if nil != err {
			t.Fail()
		}

		fmt.Println("reply1", reply1)
	}
}