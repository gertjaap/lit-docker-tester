package main

import (
	"bytes"
	"encoding/hex"
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

	fmt.Println("Importing oracle on LIT1")
	importResult, err := commands.ImportOracle(rpcCon1, "https://oracle.gertjaap.org/", "Gert-Jaap")
	handleErrorIfNeeded(err)

	if importResult.Oracle.Name != "Gert-Jaap" {
		handleErrorIfNeeded(fmt.Errorf("Oracle import failed. Name %s unexpected", importResult.Oracle.Name))
	}

	oIdx := importResult.Oracle.Idx

	fmt.Println("Creating contract on LIT1")
	contract, err := commands.NewContract(rpcCon1)
	handleErrorIfNeeded(err)

	cIdx := contract.Contract.Idx
	cKey := contract.Contract.PubKey

	_, err = commands.SetContractOracle(rpcCon1, cIdx, oIdx)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractSettlementTime(rpcCon1, cIdx, 1524156300)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractDatafeed(rpcCon1, cIdx, 1)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractFunding(rpcCon1, cIdx, 5000000, 5000000)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractSettlementDivision(rpcCon1, cIdx, 11950, 12100)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractCoinType(rpcCon1, cIdx, 257)
	handleErrorIfNeeded(err)
	_, err = commands.OfferContract(rpcCon1, cIdx, 1)
	handleErrorIfNeeded(err)

	time.Sleep(time.Millisecond * 500)

	fmt.Println("Checking if contract is received on LIT2")

	contracts2, err := commands.ListContracts(rpcCon2)

	cIdx2 := uint64(0)
	for _, contract := range contracts2.Contracts {
		if bytes.Equal(contract.PubKey[:], cKey[:]) {
			cIdx2 = contract.Idx
			break
		}
	}

	if cIdx2 == 0 {
		handleErrorIfNeeded(fmt.Errorf("Contract offer not received on LIT2"))
	}

	fmt.Println("Accepting contract on LIT2")
	_, err = commands.AcceptContract(rpcCon2, cIdx2)
	handleErrorIfNeeded(err)

	time.Sleep(time.Millisecond * 5000)

	fmt.Println("Checking if contract status on LIT1 is 'Active'")
	contract1, err := commands.GetContract(rpcCon1, cIdx)
	handleErrorIfNeeded(err)

	if contract1.Contract.Status != commands.ContractStatusActive {
		handleErrorIfNeeded(fmt.Errorf("Contract status on LIT1 not Active"))
	}

	mineBlockAndWait()

	fmt.Println("Checking if balances reflect the contract funding")

	lit1Balance, err = commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)

	if lit1Balance.Balances[0].CoinType == 257 && lit1Balance.Balances[0].MatureWitty == initialFundingAmount-5000500 {
		fmt.Println("LIT1 shows correct funds")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT1 did not have the correct balance or cointype: %d / %d", lit1Balance.Balances[0].CoinType, lit1Balance.Balances[0].MatureWitty))
	}

	lit2Balance, err = commands.GetBalance(rpcCon2)
	handleErrorIfNeeded(err)

	if lit2Balance.Balances[0].CoinType == 257 && lit2Balance.Balances[0].MatureWitty == initialFundingAmount-5000500 {
		fmt.Println("LIT2 shows correct funds")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT2 did not have the correct balance or cointype: %d / %d", lit2Balance.Balances[0].CoinType, lit2Balance.Balances[0].MatureWitty))
	}

	fmt.Println("Settling contract on LIT1")
	var oracleSig [32]byte
	oracleSigBytes, _ := hex.DecodeString("34c51d6246f814491cb0458720fc6692ae93cb75b71d4a7e6373d0fc486081b2")
	copy(oracleSig[:], oracleSigBytes)
	_, err = commands.SettleContract(rpcCon1, cIdx, 12080, oracleSig)
	handleErrorIfNeeded(err)

	mineBlockAndWait()

	lit1Balance, err = commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)
	lit2Balance, err = commands.GetBalance(rpcCon2)
	handleErrorIfNeeded(err)

	fmt.Printf("Concluded tests succesfully. End balances: LIT1 [%d] - LIT2 [%d]\n", lit1Balance.Balances[0].MatureWitty, lit2Balance.Balances[0].MatureWitty)

	for true {
		time.Sleep(time.Minute)
	}
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
