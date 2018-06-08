package commands

import (
	"net/rpc"
)

type PayMultihopArgs struct {
	DestLNAdr string
	CoinType  uint32
	Amt       int64
}

func PayMultihop(c *rpc.Client, address string, coinType uint32, amount int64) (*StatusReply, error) {
	args := new(PayMultihopArgs)
	args.DestLNAdr = address
	args.CoinType = coinType
	args.Amt = amount

	reply := new(StatusReply)
	err := c.Call("LitRPC.PayMultihop", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
