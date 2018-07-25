package commands

import (
	"net/rpc"
)

type ChanArgs struct {
	ChanIdx uint32
}

func Break(c *rpc.Client, chanIdx uint32) (*StatusReply, error) {
	args := new(ChanArgs)
	args.ChanIdx = chanIdx

	reply := new(StatusReply)
	err := c.Call("LitRPC.BreakChannel", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
