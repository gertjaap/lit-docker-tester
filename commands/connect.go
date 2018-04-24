package commands

import (
	"net/rpc"
)

type ConnectArgs struct {
	LNAddr string
}

func Connect(c *rpc.Client, addr string) (*StatusReply, error) {
	args := new(ConnectArgs)
	args.LNAddr = addr

	reply := new(StatusReply)
	err := c.Call("LitRPC.Connect", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
