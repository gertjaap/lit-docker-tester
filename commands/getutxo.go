package commands

import (
	"net/rpc"
)

type TxoInfo struct {
	OutPoint string
	Amt      int64
	Height   int32
	Delay    int32
	CoinType string
	Witty    bool

	KeyPath string
}
type TxoListReply struct {
	Txos []TxoInfo
}

func ListUtxos(c *rpc.Client) (*TxoListReply, error) {
	reply := new(TxoListReply)
	err := c.Call("LitRPC.TxoList", nil, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
