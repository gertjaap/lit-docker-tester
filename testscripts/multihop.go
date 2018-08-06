package testscripts

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/gertjaap/lit-docker-tester/btc"

	"github.com/gertjaap/lit-docker-tester/commands"
)

func MultihopTest() {
	fmt.Println("LIT Multihop Tester Script")

	rpcConns, wsConns := ConnectAndFund()
	for _, wsConn := range wsConns {
		defer wsConn.Close()
	}

	fmt.Println("Connecting node path and funding channels")
	for i := 0; i < len(rpcConns)-1; i++ {
		fmt.Printf("Funding channel between lit%d and lit%d\n", i+1, i+2)
		FundChannelBetween(rpcConns[i], rpcConns[i+1], fmt.Sprintf("lit%d", i+1), 257, 1000000, 500000)
		btc.MineBlocks(1)
		time.Sleep(time.Second * 1)
		if i < 7 {
			fmt.Printf("Funding channel between lit%d and lit%d\n", i+1, i+4)
			FundChannelBetween(rpcConns[i], rpcConns[i+3], fmt.Sprintf("lit%d", i+1), 257, 1000000, 500000)
		}
		btc.MineBlocks(1)
		time.Sleep(time.Second * 1)
		if i > 2 {
			fmt.Printf("Funding channel between lit%d and lit%d\n", i+1, i-1)
			FundChannelBetween(rpcConns[i], rpcConns[i-2], fmt.Sprintf("lit%d", i+1), 257, 1000000, 500000)
		}
		btc.MineBlocks(1)
		time.Sleep(time.Second * 1)
	}

	addresses := make([]string, len(rpcConns))
	for i := 0; i < len(rpcConns); i++ {
		reply, err := commands.GetListeningPorts(rpcConns[i])
		handleErrorIfNeeded(err)
		addresses[i] = reply.Adr
	}

	btc.MineBlocks(1)
	fmt.Println("Waiting for 10 seconds for all the channels to propagate")
	time.Sleep(time.Second * 5)
	btc.MineBlocks(1)
	time.Sleep(time.Second * 5)
	btc.MineBlocks(1)
	fmt.Println("Paying the last node from the first")
	reply, err := commands.PayMultihop(rpcConns[0], addresses[len(addresses)-1], 257, 100000)
	handleErrorIfNeeded(err)
	fmt.Printf("Paying result: %s", reply.Status)
	fmt.Println("Done.")
}

func FundChannelBetween(rpcCon1, rpcCon2 *rpc.Client, hostName1 string, coinType uint32, amount, initialSend int64) {
	peers1, err := commands.ListConnections(rpcCon1)
	handleErrorIfNeeded(err)

	ConnectTogether(rpcCon1, rpcCon2, hostName1)
	peerIdx := uint32(0)
	for peerIdx == 0 {
		peers2, err := commands.ListConnections(rpcCon1)
		handleErrorIfNeeded(err)
		for _, peer := range peers2.Connections {
			found := false
			for _, existingPeer := range peers1.Connections {
				if existingPeer.PeerNumber == peer.PeerNumber && existingPeer.RemoteHost == peer.RemoteHost {
					found = true
				}
			}
			if !found {
				peerIdx = peer.PeerNumber
				break
			}
		}
		time.Sleep(time.Millisecond * 500)
	}

	reply, err := commands.Fund(rpcCon1, peerIdx, 257, amount, initialSend)
	handleErrorIfNeeded(err)
	fmt.Printf("%s\n", reply.Status)

}
