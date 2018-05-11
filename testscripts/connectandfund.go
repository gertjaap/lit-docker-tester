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

func ConnectAndFund() (*rpc.Client, *rpc.Client, *websocket.Conn, *websocket.Conn) {
	fmt.Println("LIT Fund & Connect Script")

	fmt.Println("Connecting to LIT1")
	wsConn1, err := websocket.Dial("ws://lit1:8001/ws", "", "http://127.0.0.1/")
	handleErrorIfNeeded(err)

	fmt.Println("Connecting to LIT2")
	wsConn2, err := websocket.Dial("ws://lit2:8001/ws", "", "http://127.0.0.1/")
	handleErrorIfNeeded(err)

	rpcCon1 := jsonrpc.NewClient(wsConn1)
	rpcCon2 := jsonrpc.NewClient(wsConn2)

	fmt.Println("Connecting both LIT clients to each other")

	lit1ListenDetails, err := commands.Listen(rpcCon1, ":2448")
	handleErrorIfNeeded(err)

	con2Result, err := commands.Connect(rpcCon2, lit1ListenDetails.Adr+"@lit1")
	handleErrorIfNeeded(err)

	if !strings.HasPrefix(con2Result.Status, "connected to peer") {
		handleErrorIfNeeded(fmt.Errorf("Connect result unexpected: %s", con2Result.Status))
	}

	err = btc.WaitReady()
	handleErrorIfNeeded(err)

	fmt.Println("Getting addresses from LIT nodes")
	lit1Addresses, err := commands.GetAddresses(rpcCon1)
	handleErrorIfNeeded(err)
	lit2Addresses, err := commands.GetAddresses(rpcCon2)
	handleErrorIfNeeded(err)

	fmt.Println("Funding the LIT nodes")
	err = btc.SendCoins(lit1Addresses.WitAddresses[0], initialFundingAmount)
	handleErrorIfNeeded(err)
	err = btc.SendCoins(lit2Addresses.WitAddresses[0], initialFundingAmount)
	handleErrorIfNeeded(err)

	mineBlockAndWait()

	fmt.Println("Checking balances on the LIT nodes")
	lit1Balance, err := commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)

	if lit1Balance.Balances[0].CoinType == 257 && lit1Balance.Balances[0].MatureWitty == initialFundingAmount {
		fmt.Println("LIT1 correctly funded")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT1 did not have the correct balance or cointype: %d / %d", lit1Balance.Balances[0].CoinType, lit1Balance.Balances[0].MatureWitty))
	}

	lit2Balance, err := commands.GetBalance(rpcCon2)
	handleErrorIfNeeded(err)

	if lit2Balance.Balances[0].CoinType == 257 && lit2Balance.Balances[0].MatureWitty == initialFundingAmount {
		fmt.Println("LIT2 correctly funded")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT2 did not have the correct balance or cointype: %d / %d", lit2Balance.Balances[0].CoinType, lit2Balance.Balances[0].MatureWitty))
	}

	fmt.Println("Funding done!")
	return rpcCon1, rpcCon2, wsConn1, wsConn2
}
