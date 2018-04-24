package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"time"

	"github.com/gertjaap/lit-docker-tester/btc"
	"github.com/gertjaap/lit-docker-tester/commands"

	"golang.org/x/net/websocket"
)

func handleErrorIfNeeded(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
}

var initialFundingAmount = int64(20000000)

func main() {
	fmt.Println("LIT Tester Script")

	fmt.Println("Connecting to LIT1")
	wsConn1, err := websocket.Dial("ws://lit1:8001/ws", "", "http://127.0.0.1/")
	handleErrorIfNeeded(err)
	defer wsConn1.Close()

	fmt.Println("Connecting to LIT2")
	wsConn2, err := websocket.Dial("ws://lit2:8001/ws", "", "http://127.0.0.1/")
	handleErrorIfNeeded(err)
	defer wsConn2.Close()

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

	if lit1Balance.Balances[0].CoinType == 257 && lit1Balance.Balances[0].MatureWitty == initialFundingAmount {
		fmt.Println("LIT1 correctly funded")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT1 did not have the correct balance or cointype: %d / %d", lit1Balance.Balances[0].CoinType, lit1Balance.Balances[0].MatureWitty))
	}

	lit2Balance, err := commands.GetBalance(rpcCon2)

	if lit2Balance.Balances[0].CoinType == 257 && lit2Balance.Balances[0].MatureWitty == initialFundingAmount {
		fmt.Println("LIT2 correctly funded")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT2 did not have the correct balance or cointype: %d / %d", lit2Balance.Balances[0].CoinType, lit2Balance.Balances[0].MatureWitty))
	}

	fmt.Println("Concluded tests succesfully")
}

func mineBlock() {
	mineBlocks(1)
}

func mineBlocks(num uint32) {
	fmt.Printf("Mining %d blocks on BTC Regtest\n", num)
	err := btc.MineBlocks(num)
	handleErrorIfNeeded(err)
}

func mineBlockAndWait() {
	mineBlocks(1)
	fmt.Println("Waiting for the mined block to be processed...")
	time.Sleep(2 * time.Second)
}
