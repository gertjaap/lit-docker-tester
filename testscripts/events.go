package testscripts

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/gertjaap/lit-docker-tester/btc"
	"github.com/gertjaap/lit-docker-tester/commands"
)

func Events() {
	fmt.Println("Testing event system..")

	rpcConns, wsConns := ConnectAndFund()
	for _, wsConn := range wsConns {
		defer wsConn.Close()
	}

	FundChannelBetween(rpcConns[0], rpcConns[1], "lit1", 257, 10000000, 5000000)
	btc.MineBlocks(5)
	time.Sleep(1 * time.Second)

	replyChan := make(chan commands.LitEvent)
	go func() {
		GetEvent(rpcConns[1], replyChan)
	}()

	Push(rpcConns[0], 1, 10000)

	event := <-replyChan

	fmt.Printf("Received event of type %d", event)

}

func GetEvent(rpcConn *rpc.Client, replyTo chan commands.LitEvent) {
	reply, err := commands.GetEvent(rpcConn)
	handleErrorIfNeeded(err)
	replyTo <- reply.Event
}
