package commands

import (
	"net/rpc"
)

type RemoteControlAuthorization struct {
	Allowed bool
}

type RCAuthArgs struct {
	PubKey        [33]byte
	Authorization *RemoteControlAuthorization
}

type RCSendArgs struct {
	PeerIdx uint32
	Msg     []byte
}

func RCAuth(c *rpc.Client, pubKey [33]byte, auth bool) (*StatusReply, error) {
	args := new(RCAuthArgs)
	args.Authorization = new(RemoteControlAuthorization)
	args.Authorization.Allowed = auth
	args.PubKey = pubKey

	reply := new(StatusReply)
	err := c.Call("LitRPC.RemoteControlAuth", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func RCSend(c *rpc.Client, peerIdx uint32, msg []byte) (*StatusReply, error) {
	args := new(RCSendArgs)
	args.PeerIdx = peerIdx
	args.Msg = msg

	reply := new(StatusReply)
	err := c.Call("LitRPC.RemoteControlSend", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
