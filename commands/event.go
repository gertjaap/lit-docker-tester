package commands

import (
	"net/rpc"
)

// EventType is an enumeration containing the various events that
// can happen, that are forwarded to the RPC surface
type LitEventType int

const (
	LitEventTypeChannelPushReceived LitEventType = 0
)

type LitEvent struct {
	Type LitEventType
}

type EventReply struct {
	Event LitEvent
}

func GetEvent(c *rpc.Client) (*EventReply, error) {
	args := new(NoArgs)

	reply := new(EventReply)
	err := c.Call("LitRPC.GetEvent", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
