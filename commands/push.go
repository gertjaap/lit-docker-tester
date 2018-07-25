package commands

import (
	"net/rpc"
)

type PushArgs struct {
	ChanIdx uint32
	Amt     int64
	Data    [32]byte
}
type PushReply struct {
	StateIndex uint64
}

func Push(c *rpc.Client, chanIdx uint32, amount int64, data [32]byte) (*PushReply, error) {
	args := new(PushArgs)
	args.ChanIdx = chanIdx
	args.Data = data
	args.Amt = amount

	reply := new(PushReply)
	err := c.Call("LitRPC.Push", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
