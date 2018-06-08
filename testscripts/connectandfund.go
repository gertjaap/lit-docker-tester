package testscripts

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	"github.com/gertjaap/lit-docker-tester/btc"
	"github.com/gertjaap/lit-docker-tester/commands"

	"golang.org/x/net/websocket"
)

var initialFundingAmount = int64(1000000000)

func ConnectAndFund() ([]*rpc.Client, []*websocket.Conn) {
	fmt.Println("LIT Fund & Connect Script")

	wsConns := make([]*websocket.Conn, 10)
	rpcConns := make([]*rpc.Client, 10)

	for i := 0; i < 10; i++ {
		fmt.Printf("Connecting to LIT%d\n", i+1)
		wsConn, err := websocket.Dial(fmt.Sprintf("ws://lit%d:8001/ws", i+1), "", "http://127.0.0.1/")
		wsConns[i] = wsConn
		handleErrorIfNeeded(err)
		rpcConns[i] = jsonrpc.NewClient(wsConns[i])
	}

	err := btc.WaitReady()
	handleErrorIfNeeded(err)

	fmt.Println("Funding the LIT nodes")
	for i := 0; i < 10; i++ {
		addr, err := commands.GetAddresses(rpcConns[i])
		handleErrorIfNeeded(err)
		err = btc.SendCoins(addr.WitAddresses[0], initialFundingAmount)
		handleErrorIfNeeded(err)
	}

	mineBlockAndWait()

	fmt.Println("Checking balances on the LIT nodes")
	for i := 0; i < 10; i++ {
		balance, err := commands.GetBalance(rpcConns[i])
		handleErrorIfNeeded(err)
		if balance.Balances[0].CoinType == 257 && balance.Balances[0].MatureWitty == initialFundingAmount {
			fmt.Printf("LIT%d correctly funded\n", i+1)
		} else {
			handleErrorIfNeeded(fmt.Errorf("LIT%d did not have the correct balance or cointype: %d / %d", i+1, balance.Balances[0].CoinType, balance.Balances[0].MatureWitty))
		}
	}
	fmt.Println("Funding done!")
	return rpcConns, wsConns
}

func ConnectTogether(rpcCon1, rpcCon2 *rpc.Client, hostName1 string) {
	lit1ListenDetails, err := commands.Listen(rpcCon1, ":2448")
	handleErrorIfNeeded(err)
	if err != nil { // already listening
		lit1ListenDetails, _ = commands.GetListeningPorts(rpcCon1)
	}

	// make two also listen just in case
	_, err = commands.Listen(rpcCon2, ":2448")
	handleErrorIfNeeded(err)

	con2Result, err := commands.Connect(rpcCon2, lit1ListenDetails.Adr+"@"+hostName1)
	handleErrorIfNeeded(err)

	if !strings.HasPrefix(con2Result.Status, "connected to peer") {
		handleErrorIfNeeded(fmt.Errorf("Connect result unexpected: %s", con2Result.Status))
	}
}
