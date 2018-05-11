package testscripts

import (
	"fmt"
	"time"

	"github.com/gertjaap/lit-docker-tester/btc"
)

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
