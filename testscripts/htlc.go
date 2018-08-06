package testscripts

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/gertjaap/lit-docker-tester/commands"

	"crypto/rand"
	"crypto/sha256"

	"github.com/gertjaap/lit-docker-tester/btc"
)

type ChannelOperationType int

const (
	OpDeltaSig    ChannelOperationType = 0
	OpHashSig     ChannelOperationType = 1
	OpPreimageSig ChannelOperationType = 2
)

var globalPreImage [16]byte
var globalHash [32]byte

func FundChannels(rpcConns []*rpc.Client) {
	fmt.Println("Funding channel between lit1 and lit2")
	FundChannelBetween(rpcConns[0], rpcConns[1], "lit1", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Funding channel between lit2 and lit3")
	FundChannelBetween(rpcConns[1], rpcConns[2], "lit2", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Funding channel between lit3 and lit4")
	FundChannelBetween(rpcConns[2], rpcConns[3], "lit3", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Funding channel between lit4 and lit5")
	FundChannelBetween(rpcConns[3], rpcConns[4], "lit4", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Funding channel between lit5 and lit6")
	FundChannelBetween(rpcConns[4], rpcConns[5], "lit5", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Funding channel between lit6 and lit1")
	FundChannelBetween(rpcConns[5], rpcConns[0], "lit6", 257, 10000000, 5000000)
	btc.MineBlocks(1)
	fmt.Println("Mining two blocks to confirm channels")
	btc.MineBlocks(1)
	btc.MineBlocks(1)

	time.Sleep(time.Second * 1)
	fmt.Println("Channels funded")
}

func HtlcTest() {
	fmt.Println("LIT HTLC Tester Script")

	fmt.Println("Connecting to LIT nodes..")
	rpcConns, wsConns := ConnectAndFund()
	for _, wsConn := range wsConns {
		defer wsConn.Close()
	}

	globalPreImage, globalHash = GetPreimageAndHash()

	FundChannels(rpcConns)
	/*HtlcTestMultipleOffChain(rpcConns, 1, false)
	HtlcTestMultipleOffChain(rpcConns, 10, false)
	HtlcTestMultipleOffChain(rpcConns, 3, true)
	HtlcTestMultipleOffChainTimeout(rpcConns, 1, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 1, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 10, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 3, true)
	HtlcTestOnChain(rpcConns[0], rpcConns[5], 2, 2, false, false)
	HtlcTestOnChain(rpcConns[1], rpcConns[2], 2, 1, true, false)
	HtlcTestOnChain(rpcConns[3], rpcConns[4], 2, 1, false, true)
	HtlcTestOnChain(rpcConns[4], rpcConns[5], 2, 1, true, true)*/

	//=== HashSig-HashSig collision
	//HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, 0, 0, OpHashSig, OpHashSig)

	//=== HashSig-Deltasig collision
	//HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, 0, 0, OpHashSig, OpDeltaSig)

	//=== HashSig-PreimageSig collision
	//idx := AddHtlc(rpcConns[0], 1, 200000, globalHash)
	//HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, idx, 0, OpPreimageSig, OpHashSig)

	//=== PreimageSig-PreimageSig collision
	//idx1 := AddHtlc(rpcConns[0], 1, 200000, globalHash)
	//idx2 := AddHtlc(rpcConns[1], 1, 200000, globalHash)
	//HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, idx1, idx2, OpPreimageSig, OpPreimageSig)

	//=== PreimageSig-Deltasig collision
	//idx := AddHtlc(rpcConns[0], 1, 200000, globalHash)
	//HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, idx, 0, OpPreimageSig, OpDeltaSig)

	//=== DeltaSig-DeltaSig collision
	HtlcTestCollision(rpcConns[0], rpcConns[1], 1, 1, 0, 0, OpDeltaSig, OpDeltaSig)
}

func HtlcTestCollision(rpcConn1, rpcConn2 *rpc.Client, chanIdx1, chanIdx2, htlcIdx1, htlcIdx2 uint32, op1, op2 ChannelOperationType) {
	done1 := make(chan bool)
	done2 := make(chan bool)
	go testCollision(rpcConn1, chanIdx1, op1, htlcIdx1, done1)
	go testCollision(rpcConn2, chanIdx2, op2, htlcIdx2, done2)

	isDone := <-done1
	fmt.Printf("Collision test done1: %t\n", isDone)
	isDone = <-done2
	fmt.Printf("Collision test done2: %t\n", isDone)

	fmt.Println("Checking if channel is still usable by pushing some funds back and forth...")
	// Check if channel is still in usable state afterwards
	for i := 0; i < 10; i++ {
		fmt.Printf("Push %d\n", i)
		Push(rpcConn1, chanIdx1, 10000)
		Push(rpcConn2, chanIdx2, 10000)
	}

	fmt.Println("Checking if channel is still usable for HTLC by adding and clearing a bunch of HTLC off-chain")
	for i := 0; i < 10; i++ {
		preimage, hash := GetPreimageAndHash()
		idx := AddHtlc(rpcConn1, chanIdx1, 200000, hash)
		fmt.Printf("Added HTLC %d\n", idx)
		ClearHtlc(rpcConn2, chanIdx2, idx, preimage)
		fmt.Printf("Cleared HTLC %d\n", idx)
	}
	fmt.Println("Collision test success")
}

func testCollision(rpcConn *rpc.Client, chanIdx uint32, op ChannelOperationType, htlcIdx uint32, done chan bool) {
	if op == OpHashSig {
		AddHtlc(rpcConn, chanIdx, 200000, globalHash)
	} else if op == OpPreimageSig {
		ClearHtlc(rpcConn, chanIdx, htlcIdx, globalPreImage)
	} else {
		Push(rpcConn, chanIdx, 10000)
	}
	done <- true
}

func Push(rpcConn *rpc.Client, chanIdx uint32, amt int64) {
	_, err := commands.Push(rpcConn, chanIdx, amt, [32]byte{})
	handleErrorIfNeeded(err)
}

func ClearHtlc(rpcConn *rpc.Client, chanIdx, htlcIdx uint32, preimage [16]byte) {
	_, err := commands.ClearHTLC(rpcConn, chanIdx, htlcIdx, preimage, [32]byte{})
	handleErrorIfNeeded(err)
}

func AddHtlc(rpcConn *rpc.Client, chanIdx uint32, amt int64, hash [32]byte) uint32 {
	return AddHtlcWithCustomLocktime(rpcConn, chanIdx, amt, hash, 20)
}

func AddHtlcWithCustomLocktime(rpcConn *rpc.Client, chanIdx uint32, amt int64, hash [32]byte, locktime uint32) uint32 {
	reply, err := commands.AddHTLC(rpcConn, chanIdx, amt, locktime, hash, [32]byte{})
	handleErrorIfNeeded(err)
	return reply.HTLCIndex
}

func CheckChannelBalance(rpcConn *rpc.Client, chanIdx uint32, expectedBalance int64) {
	chanReply, err := commands.ListChannels(rpcConn)
	found := false
	handleErrorIfNeeded(err)
	for _, c := range chanReply.Channels {
		if c.CIdx == chanIdx {
			gotBalance := c.MyBalance
			if gotBalance != expectedBalance {
				handleErrorIfNeeded(fmt.Errorf("Wrong balance - Expected [%d], Got [%d]", expectedBalance, gotBalance))
			}
			found = true
		}
	}

	if !found {
		handleErrorIfNeeded(fmt.Errorf("Channel [%d] not found while checking balance", chanIdx))
	}
}

func GetPreimageAndHash() ([16]byte, [32]byte) {
	var preimage [16]byte
	rand.Read(preimage[:])
	hash := sha256.Sum256(preimage[:])
	return preimage, hash
}

func HtlcTestMultipleOffChain(rpcConns []*rpc.Client, count int, pushes bool) {
	withPushes := ""
	if pushes {
		withPushes = "(with pushes between HTLC operations)"
	}

	fmt.Printf("Testing multiple concurrent HTLCs off-chain success %s\n", withPushes)

	fmt.Printf("Adding %d HTLCs in channel between lit1 and lit2...\n", count)
	preimages := make([][16]byte, count)
	hashes := make([][32]byte, count)
	htlcIdxs := make([]uint32, count)
	for i := 0; i < count; i++ {
		fmt.Printf("Adding HTLC %d\n", i+1)
		preimages[i], hashes[i] = GetPreimageAndHash()
		htlcIdxs[i] = AddHtlc(rpcConns[0], 1, 200000, hashes[i])
		fmt.Printf("Success, idx: %d\n", htlcIdxs[i])
		if pushes {
			Push(rpcConns[1], 1, 10000)
		}
	}

	// Wait for negotiations to finish
	time.Sleep(time.Second * 3)

	fmt.Println("Checking balances after adding HTLCs")
	if pushes {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*190000)
		CheckChannelBalance(rpcConns[1], 1, 5000000-int64(count)*10000)
	} else {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*200000)
		CheckChannelBalance(rpcConns[1], 1, 5000000)
	}

	fmt.Println("Clearing the HTLCs between lit1 and 2 using preimages...")
	for i := 0; i < count; i++ {
		ClearHtlc(rpcConns[0], 1, htlcIdxs[i], preimages[i])
		if pushes {
			Push(rpcConns[1], 1, 10000)
		}
	}
	fmt.Println("Checking channel balances after clearing HTLCs")

	if pushes {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*180000)
		CheckChannelBalance(rpcConns[1], 1, 5000000+int64(count)*180000)
	} else {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*200000)
		CheckChannelBalance(rpcConns[1], 1, 5000000+int64(count)*200000)
	}

	fmt.Println("Rebalancing channel")
	if pushes {
		Push(rpcConns[1], 1, int64(count)*180000)
	} else {
		Push(rpcConns[1], 1, int64(count)*200000)
	}

	fmt.Printf("Done - %d concurrent HTLC Off Chain %s (success) test succesful\n", count, withPushes)
}

func HtlcTestMultipleOffChainTimeout(rpcConns []*rpc.Client, count int, pushes bool) {
	withPushes := ""
	if pushes {
		withPushes = "(with pushes between HTLC operations)"
	}

	fmt.Printf("Testing multiple concurrent HTLCs off-chain timeout %s\n", withPushes)

	fmt.Printf("Adding %d HTLCs in channel between lit1 and lit2...\n", count)
	preimages := make([][16]byte, count)
	hashes := make([][32]byte, count)
	htlcIdxs := make([]uint32, count)
	for i := 0; i < count; i++ {
		fmt.Printf("Adding HTLC %d\n", i+1)
		preimages[i], hashes[i] = GetPreimageAndHash()
		htlcIdxs[i] = AddHtlc(rpcConns[0], 1, 200000, hashes[i])
		fmt.Printf("Success, idx: %d\n", htlcIdxs[i])
		if pushes {
			Push(rpcConns[1], 1, 10000)
		}
	}

	// Wait for negotiations to finish
	time.Sleep(time.Second * 3)

	fmt.Println("Checking balances after adding HTLCs")
	if pushes {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*190000)
		CheckChannelBalance(rpcConns[1], 1, 5000000-int64(count)*10000)
	} else {
		CheckChannelBalance(rpcConns[0], 1, 5000000-int64(count)*200000)
		CheckChannelBalance(rpcConns[1], 1, 5000000)
	}

	fmt.Println("Timing out the HTLCs between lit1 and 2 by mining a few blocks..")
	btc.MineBlocks(uint32(count) + 5)
	time.Sleep(time.Second * 10)

	if pushes {
		CheckChannelBalance(rpcConns[0], 1, 5000000+int64(count)*10000)
		CheckChannelBalance(rpcConns[1], 1, 5000000-int64(count)*10000)
	} else {
		CheckChannelBalance(rpcConns[0], 1, 5000000)
		CheckChannelBalance(rpcConns[1], 1, 5000000)
	}

	if pushes {
		fmt.Println("Rebalancing channel")
		Push(rpcConns[0], 1, int64(count)*10000)
	}

	// Wait for the Push to finish
	time.Sleep(time.Second * 1)

	fmt.Printf("Done - %d concurrent HTLC Off Chain timeouts %s test succesful\n", count, withPushes)
}

func Break(rpcConn *rpc.Client, chanIdx uint32) {
	reply, err := commands.Break(rpcConn, chanIdx)
	handleErrorIfNeeded(err)
	fmt.Printf("Breaking channel success: %s\n", reply.Status)
}

func ClaimHtlc(rpcConn *rpc.Client, preimage [16]byte) {
	reply, err := commands.ClaimHTLC(rpcConn, preimage)
	handleErrorIfNeeded(err)
	for _, txid := range reply.Txids {
		fmt.Printf("Claimed HTLC using TXID: [%s]\n", txid)
	}
}

func CheckUtxos(rpcConn *rpc.Client, amountsExpected []int64) {
	utxos, err := commands.ListUtxos(rpcConn)
	handleErrorIfNeeded(err)
	foundUtxos := make([]bool, len(amountsExpected))
	for _, t := range utxos.Txos {
		for i, amt := range amountsExpected {
			if t.CoinType == "regtest" && t.Amt == amt && foundUtxos[i] == false {
				foundUtxos[i] = true
				break
			}
		}
	}

	for i, found := range foundUtxos {
		if !found {
			for _, t := range utxos.Txos {
				fmt.Printf("Have TXO [%d]\n", t.Amt)
			}
			handleErrorIfNeeded(fmt.Errorf("Did not find expected utxos at index %d worth %d", i, amountsExpected[i]))
		}
	}
}

func GetLatestWitAddress(rpcConn *rpc.Client) string {
	reply, err := commands.GetAddresses(rpcConn)
	handleErrorIfNeeded(err)
	if len(reply.WitAddresses) == 0 {
		handleErrorIfNeeded(fmt.Errorf("No addresses returned"))
	}
	return reply.WitAddresses[len(reply.WitAddresses)-1]
}

func HtlcTestOnChain(rpcConn, rpcConn2 *rpc.Client, chanIdx, chanIdx2 uint32, timeout, theyBreak bool) {
	whoBreaks := "offerer breaks"
	claimType := "success"
	if timeout {
		claimType = "timeout"
	}

	if theyBreak {
		whoBreaks = "recipient breaks"
	}

	fmt.Printf("Testing on chain HTLC %s, %s\n", claimType, whoBreaks)
	fee := int64(80000)

	preimage, hash := GetPreimageAndHash()
	AddHtlc(rpcConn, chanIdx, 200000, hash)

	fmt.Println("Checking channel balance...")

	CheckChannelBalance(rpcConn, 2, 4800000)

	fmt.Println("Breaking channel...")

	if theyBreak {
		Break(rpcConn2, chanIdx2)
	} else {
		Break(rpcConn, chanIdx)
	}

	btc.MineBlocks(1)
	time.Sleep(time.Second * 1)

	if timeout {
		// timeout the HTLC by mining a couple blocks
		btc.MineBlocks(100)
	} else {
		// claim the HTLC using the preimage
		ClaimHtlc(rpcConn2, preimage)
	}

	for i := 0; i < 10; i++ {
		btc.MineBlocks(1)
		time.Sleep(time.Second * 1)
	}

	fmt.Println("Checking UTXOs...")

	amountsExpected := []int64{
		4800000 - fee,
	}
	if timeout {
		amountsExpected = append(amountsExpected, 200000-fee)
	}
	CheckUtxos(rpcConn, amountsExpected)

	amountsExpected = []int64{
		5000000 - fee,
	}
	if !timeout {
		amountsExpected = append(amountsExpected, 200000-fee)
	}
	CheckUtxos(rpcConn2, amountsExpected)

	fmt.Println("Trying spending all UTXOs to own address")

	sendAmount := int64(0)
	sendAmount2 := int64(0)
	balances, err := commands.GetBalance(rpcConn)
	handleErrorIfNeeded(err)
	sendAmount = balances.Balances[0].MatureWitty

	balances, err = commands.GetBalance(rpcConn2)
	handleErrorIfNeeded(err)
	sendAmount2 = balances.Balances[0].MatureWitty

	adr := GetLatestWitAddress(rpcConn)
	txids, err := commands.Send(rpcConn, adr, sendAmount-fee)
	handleErrorIfNeeded(err)
	if len(txids.Txids) != 1 {
		handleErrorIfNeeded(fmt.Errorf("Unexpected number of TXs in TXIDReply: %d", len(txids.Txids)))
	}

	fmt.Printf("Spent using TXID %s\n", txids.Txids[0])

	adr = GetLatestWitAddress(rpcConn2)
	txids2, err := commands.Send(rpcConn2, adr, sendAmount2-fee)
	handleErrorIfNeeded(err)
	if len(txids2.Txids) != 1 {
		handleErrorIfNeeded(fmt.Errorf("Unexpected number of TXs in TXIDReply: %d", len(txids2.Txids)))
	}

	fmt.Printf("Spent using TXID %s\n", txids.Txids[0])

	fmt.Println("Mining some blocks...")
	btc.MineBlocks(2)

	fmt.Println("Looking up the spends on the blockchain...")

	success, err := btc.CheckTx(txids.Txids[0])
	handleErrorIfNeeded(err)
	if !success {
		handleErrorIfNeeded(fmt.Errorf("Did not find transaction %s on the blockchain", txids.Txids[0]))
	}

	success2, err := btc.CheckTx(txids2.Txids[0])
	handleErrorIfNeeded(err)
	if !success2 {
		handleErrorIfNeeded(fmt.Errorf("Did not find transaction %s on the blockchain", txids.Txids[0]))
	}

	fmt.Printf("Done - Testing on chain HTLC %s, %s succesful\n", claimType, whoBreaks)
}

/*
	fmt.Println("Funding channel between lit1 and lit2")
	FundChannelBetween(rpcConns[0], rpcConns[1], "lit1", 257, 1000000, 500000)
	fmt.Println("Funding channel between lit2 and lit3")
	FundChannelBetween(rpcConns[1], rpcConns[2], "lit2", 257, 1000000, 500000)

	fmt.Println("Mining a block")
	fmt.Println("Mining a block")
	btc.MineBlocks(1)
	time.Sleep(time.Second * 1)

	var preimage [16]byte
	rand.Read(preimage[:])
	hash := sha256.Sum256(preimage[:])

	fmt.Println("Adding the HTLC between lit1 and 2...")
	reply, err := commands.AddHTLC(rpcConns[0], 1, 300000, 5, hash, [32]byte{})
	handleErrorIfNeeded(err)
	fmt.Println("Adding the HTLC between lit2 and 3...")
	reply2, err := commands.AddHTLC(rpcConns[1], 2, 300000, 5, hash, [32]byte{})
	handleErrorIfNeeded(err)
	fmt.Println("Mining a block")
	btc.MineBlocks(1)
	time.Sleep(time.Second * 1)
	handleErrorIfNeeded(err)

	fmt.Printf("State index is %d - HTLC Index is %d\n", reply.StateIndex, reply.HTLCIndex)
	fmt.Printf("State index is %d - HTLC Index is %d\n", reply2.StateIndex, reply2.HTLCIndex)
	// success! Now  break the channel. The output will be released. We need to capture
	// it in the wallit and keep our eyes open for the preimage

	/*time.Sleep(time.Second * 10)

	for i := 1; i <= 10; i++ {
		fmt.Printf("Doing push [%d] to generate older states...\n", i)

		_, err := commands.Push(rpcConns[1], 2, 1000, [32]byte{})
		handleErrorIfNeeded(err)
		time.Sleep(time.Millisecond * 200)
	}

	_, err = commands.Break(rpcConns[0], 1)
	btc.MineBlocks(1)
	time.Sleep(time.Second * 2)
	_, err = commands.Break(rpcConns[1], 2)
	btc.MineBlocks(1)
	time.Sleep(time.Second * 2)
	handleErrorIfNeeded(err)

	//commands.ClaimHTLC(rpcConns[2], preimage)
	btc.MineBlocks(1)

	// Timeout!
	btc.MineBlocks(10)
	time.Sleep(time.Second * 5)

	btc.MineBlocks(1)
	time.Sleep(time.Second * 2)

	fmt.Println("Done!")
}
*/
