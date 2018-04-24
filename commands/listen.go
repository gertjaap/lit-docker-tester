package commands

import (
	"net/rpc"
)

type ListenArgs struct {
	Port string
}

type ListeningPortsReply struct {
	LisIpPorts []string
	Adr        string
}

func Listen(c *rpc.Client, port string) (*ListeningPortsReply, error) {
	args := new(ListenArgs)
	args.Port = port

	reply := new(ListeningPortsReply)
	err := c.Call("LitRPC.Listen", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
