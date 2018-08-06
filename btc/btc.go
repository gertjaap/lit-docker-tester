package btc

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

func GetRpcClient() (*rpcclient.Client, error) {
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "litbtcregtest:19001",
		User:         "lit",
		Pass:         "lit",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return client, nil

}

func WaitReady() error {
	fmt.Print("Waiting for BTC RPC to be ready...")
	client, err := GetRpcClient()
	if err != nil {
		return err
	}

	for true {
		blocks, err := client.GetBlockCount()
		if err == nil {
			fmt.Printf("ready (%d blocks)\n", blocks)
			break
		}
		time.Sleep(time.Second * 5)
	}

	return nil
}

func MineBlocks(num uint32) error {
	client, err := GetRpcClient()
	if err != nil {
		return err
	}

	_, err = client.Generate(num)
	if err != nil {
		return err
	}

	return nil
}

func SendCoins(addr string, amt int64) error {
	client, err := GetRpcClient()
	if err != nil {
		return err
	}
	btcAddr, err := btcutil.DecodeAddress(addr, &chaincfg.RegressionNetParams)
	if err != nil {
		return err
	}

	_, err = client.SendFrom("", btcAddr, btcutil.Amount(amt))
	if err != nil {
		return err
	}

	return nil
}

func CheckTx(txidHex string) (bool, error) {
	client, err := GetRpcClient()
	if err != nil {
		return false, err
	}
	txHash, err := chainhash.NewHashFromStr(txidHex)
	if err != nil {
		return false, err
	}
	tx, err := client.GetRawTransaction(txHash)
	if err != nil {
		return false, err
	}
	return (tx.Hash().String() == txidHex), nil
}
