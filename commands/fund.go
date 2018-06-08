package commands

import (
	"net/rpc"
)

type FundArgs struct {
	Peer        uint32 // who to make the channel with
	CoinType    uint32 // what coin to use
	Capacity    int64  // later can be minimum capacity
	Roundup     int64  // ignore for now; can be used to round-up capacity
	InitialSend int64  // Initial send of -1 means "ALL"
	Data        [32]byte
}

func Fund(c *rpc.Client, peerIdx, coinType uint32, amount, initialSend int64) (*StatusReply, error) {
	args := new(FundArgs)
	args.Peer = peerIdx
	args.CoinType = coinType
	args.Capacity = amount
	args.InitialSend = initialSend

	reply := new(StatusReply)
	err := c.Call("LitRPC.FundChannel", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
