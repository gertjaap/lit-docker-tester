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
	fmt.Println("Funding channel between lit1 and lit5")
	FundChannelBetween(rpcConns[0], rpcConns[5], "lit1", 257, 10000000, 5000000)
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

	FundChannels(rpcConns)
	/*HtlcTestMultipleOffChain(rpcConns, 1, false)
	HtlcTestMultipleOffChain(rpcConns, 10, false)
	HtlcTestMultipleOffChain(rpcConns, 3, true)
	HtlcTestMultipleOffChainTimeout(rpcConns, 1, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 1, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 10, false)
	HtlcTestMultipleOffChainTimeout(rpcConns, 3, true)
	*/

	HtlcTestMultipleOnChain(rpcConns, 1, false)

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
	return AddHtlcWithCustomLocktime(rpcConn, chanIdx, amt, hash, 5)
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
	fmt.Println("Breaking channel success: %s", reply.Status)
}

func ClaimHtlc(rpcConn *rpc.Client, preimage [16]byte) {
	reply, err := commands.ClaimHTLC(rpcConn, preimage)
	handleErrorIfNeeded(err)
	for _, txid := range reply.Txids {
		fmt.Printf("Claimed HTLC using TXID: [%x]\n", txid)
	}
}

func HtlcTestMultipleOnChain(rpcConns []*rpc.Client, count int, timeout bool) {
	fmt.Println("Testing multiple concurrent HTLCs on-chain success")

	preimages := make([][16]byte, count)
	hashes := make([][32]byte, count)
	htlcIdxs := make([]uint32, count)
	for i := 0; i < count; i++ {
		fmt.Printf("Adding HTLC to rpcConns[%d] channel 2\n", i)
		preimages[i], hashes[i] = GetPreimageAndHash()
		htlcIdxs[i] = AddHtlc(rpcConns[i], 2, 200000, hashes[i])
		fmt.Printf("Success, idx: %d\n", htlcIdxs[i])
	}

	fmt.Println("Checking channel balances...")

	for i := 0; i < count; i++ {
		CheckChannelBalance(rpcConns[i], 2, 4800000)
	}

	fmt.Println("Breaking channels...")

	for i := 0; i < count; i++ {
		Break(rpcConns[i], 2)
	}

	btc.MineBlocks(5)
	time.Sleep(time.Second * 5)

	if timeout {
		// timeout the HTLCs by mining a couple blocks
		btc.MineBlocks(100)
	} else {
		// claim the HTLCs using the preimages
		for i := 0; i < count; i++ {
			ClaimHtlc(rpcConns[i], preimages[i])
		}
	}

	btc.MineBlocks(5)
	time.Sleep(time.Second * 5)
	fmt.Println("Checking wallet balances...")

	for i := 0; i < count; i++ {
		balances, err := commands.GetBalance(rpcConns[i])
		handleErrorIfNeeded(err)
		for _, b := range balances.Balances {
			fmt.Printf("Node [%d] coin [%d] witconf [%d]\n", i, b.CoinType, b.MatureWitty)
		}
	}

	fmt.Printf("Done - %d concurrent on-chain success test succesful\n", count)
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
