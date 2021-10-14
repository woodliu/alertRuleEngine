package util

import (
    "sync"
    "time"
)

type Node struct {
    NodeId string
    Duration time.Duration
    Data interface{}
    Last *Node
    Next *Node
    Repeat bool
    New bool
}

type Chain struct {
    ChainLock sync.Mutex
    NodeMap map[string]*Node
    Len int64
    Start *Node
    End *Node
}

func InitChain()*Chain{
    return &Chain{
        NodeMap: make(map[string]*Node),
        Start: nil,
        End: nil,
    }
}

func NewNode(nodeId string,timeout time.Duration,data interface{})*Node{
    return &Node{
        Duration: timeout,
        NodeId: nodeId,
        Data: data,
        Last: nil,
        Next: nil,
    }
}

func (c *Chain)addNode(node *Node,updateNodeData func(curNode,newNode *Node)){
    // If exist, just update node data
    if _,ok := c.NodeMap[node.NodeId];ok{
        c.updateNodeData(node, updateNodeData)
        return
    }

    if nil == c.Start{
        c.Start = node
        c.End = node

        node.Last = nil
        node.Next = nil
    }else{
        c.Start.Last = node
        node.Next = c.Start
        node.Last = nil
        c.Start = node
    }

    node.New = true
    c.NodeMap[node.NodeId] = node

    c.Len ++
}

// AddNodes If is new node, append it set node.New = ture. If is existed node,just update the data, and
// check if it is repeated.
func (c *Chain)AddNodes(node *Node, updateNodeData func(curNode,newNode *Node)){
    c.ChainLock.Lock()
    defer c.ChainLock.Unlock()

    for curNode := node;curNode != nil;curNode = curNode.Next{
        c.addNode(node,updateNodeData)
    }
}

func (c *Chain)GetNode(nodeId string)*Node{
    c.ChainLock.Lock()
    defer c.ChainLock.Unlock()
    return c.NodeMap[nodeId]
}

func (c *Chain)Poll(pollInterval time.Duration, isCut func (node *Node)bool)*Node{
    c.ChainLock.Lock()
    defer c.ChainLock.Unlock()

    for curNode := c.End;curNode != nil;curNode = curNode.Last{
       curNode.Duration -= pollInterval
    }

    var startCutNode *Node
    for curNode := c.End;curNode != nil;curNode = curNode.Last{
        if isCut(curNode){
            startCutNode = curNode

            c.Len --
            delete(c.NodeMap, curNode.NodeId)
            continue
        }
        break
    }

    if nil == startCutNode{
        return nil
    }

    c.End = startCutNode.Last
    c.End.Next = nil
    startCutNode.Last = nil

    return startCutNode
}

func (c *Chain)CopyNodes(copyRepeat bool, copyData func(data interface{})interface{})*Node{
    c.ChainLock.Lock()
    defer c.ChainLock.Unlock()

    cp := func(start, node *Node)*Node {
        copyNode := NewNode(node.NodeId, 0, copyData(node.Data))
        if nil == start{
            start = copyNode
        }else{
            start.Last = node
            node.Next = start
            start = node
        }

        node.New = false
        return start
    }

    var start *Node
    for curNode := c.Start;curNode != nil;curNode = curNode.Next{
        if curNode.New{
            start = cp(start,curNode)
        }

        if copyRepeat{
            if curNode.Repeat{
                start = cp(start,curNode)
                curNode.Repeat = false
            }
        }
    }
    return start
}

func (c *Chain)ExecNodeData(exec func(data interface{})){
    c.ChainLock.Lock()
    defer c.ChainLock.Unlock()

    for curNode := c.Start;curNode != nil;curNode = curNode.Next{
        exec(curNode.Data)
    }
}

func (n *Node)ExecNodeData(exec func(data interface{})){
    for curNode := n;curNode != nil;curNode = curNode.Next{
        exec(curNode.Data)
    }
}

func (c *Chain)updateNodeData(node *Node, updateNodeData func(curNode,newNode *Node)){
    if n,ok := c.NodeMap[node.NodeId];ok{
        if nil != updateNodeData{
            updateNodeData(n, node)
        }
    }
}
