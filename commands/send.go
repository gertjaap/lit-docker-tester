package commands

import (
	"net/rpc"
)

type SendArgs struct {
	DestAddrs []string
	Amts      []int64
}

func Send(c *rpc.Client, adr string, amount int64) (*TxidsReply, error) {
	args := new(SendArgs)
	args.DestAddrs = []string{adr}
	args.Amts = []int64{amount}

	reply := new(TxidsReply)
	err := c.Call("LitRPC.Send", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
