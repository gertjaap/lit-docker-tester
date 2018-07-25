package commands

import (
	"net/rpc"
)

type ClaimHTLCArgs struct {
	R [16]byte
}
type TxidsReply struct {
	Txids []string
}

func ClaimHTLC(c *rpc.Client, R [16]byte) (*TxidsReply, error) {
	args := new(ClaimHTLCArgs)
	copy(args.R[:], R[:])

	reply := new(TxidsReply)
	err := c.Call("LitRPC.ClaimHTLC", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
