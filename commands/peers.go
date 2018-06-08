package commands

import (
	"net/rpc"
)

type ListConnectionsReply struct {
	Connections []PeerInfo
	MyPKH       string
}

type PeerInfo struct {
	PeerNumber uint32
	RemoteHost string
	Nickname   string
}

func ListConnections(c *rpc.Client) (*ListConnectionsReply, error) {
	args := new(NoArgs)
	reply := new(ListConnectionsReply)
	err := c.Call("LitRPC.ListConnections", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
