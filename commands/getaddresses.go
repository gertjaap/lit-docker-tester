package commands

import (
	"net/rpc"
)

type AddressReply struct {
	WitAddresses    []string
	LegacyAddresses []string
}

func GetAddresses(c *rpc.Client) (*AddressReply, error) {
	reply := new(AddressReply)
	err := c.Call("LitRPC.Address", nil, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
