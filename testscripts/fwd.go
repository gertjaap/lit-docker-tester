package testscripts

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gertjaap/lit-docker-tester/commands"
)

func ForwardTest() {
	fmt.Println("LIT Forward contract tester script")

	rpcConns, wsConns := ConnectAndFund()
	for _, wsConn := range wsConns {
		defer wsConn.Close()
	}

	rpcCon1 := rpcConns[0]
	rpcCon2 := rpcConns[1]
	ConnectTogether(rpcCon1, rpcCon2, "lit1")

	ImportOracle(rpcCon1, rpcCon2)

	oracles, err := commands.ListOracles(rpcCon1)
	handleErrorIfNeeded(err)

	oracle := oracles.Oracles[0]

	// Set up a forward buy for 500 USD in 5 minutes.
	fmt.Println("Creating new forward offer on LIT1, sending it to LIT2")

	offer := new(commands.DlcFwdOffer)
	offer.ImBuyer = true
	offer.AssetQuantity = 5000
	offer.OracleA = oracle.A
	settleTime := uint64(time.Now().Unix())

	// Round down to the last 5 minute interval so we can settle
	// immediately.
	settleTime -= (settleTime % 300)

	offer.SettlementTime = settleTime

	// Fetch R-point
	offer.OracleR, err = oracle.FetchRPoint(1, settleTime)
	handleErrorIfNeeded(err)

	offer.FundAmt = 50000000
	offer.PeerIdx = 1
	offer.CoinType = 257

	offerReply, err := commands.NewForwardOffer(rpcCon1, offer)
	handleErrorIfNeeded(err)

	// Check if offer is received on LIT2
	offers, err := commands.ListOffers(rpcCon2)
	handleErrorIfNeeded(err)

	offerOnLit2 := new(commands.DlcFwdOffer)
	found := false
	for _, o := range offers.Offers {
		if o.TheirOIdx == offerReply.Offer.OIdx &&
			bytes.Equal(o.OracleA[:], offer.OracleA[:]) &&
			bytes.Equal(o.OracleR[:], offer.OracleR[:]) &&
			o.SettlementTime == offer.SettlementTime {
			offerOnLit2 = o
			found = true
		}
	}

	if !found {
		handleErrorIfNeeded(fmt.Errorf("Did not find offer on LIT2"))
	}

	// Accept the offer on LIT2
	commands.AcceptOffer(rpcCon2, offerOnLit2.OIdx)

	// Wait for a while - the nodes are now exchanging the actual contract
	time.Sleep(time.Second * 10)

	// Check if there is a contract in "Accepted" state on both nodes
	contracts1, err := commands.ListContracts(rpcCon1)
	handleErrorIfNeeded(err)
	contracts2, err := commands.ListContracts(rpcCon2)
	handleErrorIfNeeded(err)

	contractOne := new(commands.DlcContract)
	found = false
	for _, c := range contracts1.Contracts {
		if bytes.Equal(c.OracleA[:], offer.OracleA[:]) &&
			bytes.Equal(c.OracleR[:], offer.OracleR[:]) &&
			c.OracleTimestamp == offer.SettlementTime &&
			c.Status == commands.ContractStatusActive {
			found = true
			contractOne = c
		} else {
			fmt.Println("Found unmatching contract on LIT1:")
			commands.PrintContract(c)
		}
	}

	if !found {
		handleErrorIfNeeded(fmt.Errorf("Did not find contract on LIT1"))
	}

	found = false
	for _, c := range contracts2.Contracts {
		if bytes.Equal(c.OracleA[:], offer.OracleA[:]) &&
			bytes.Equal(c.OracleR[:], offer.OracleR[:]) &&
			c.OracleTimestamp == offer.SettlementTime &&
			c.Status == commands.ContractStatusActive {
			found = true
		} else {
			fmt.Println("Found unmatching contract on LIT2:")
			commands.PrintContract(c)
		}
	}

	if !found {
		handleErrorIfNeeded(fmt.Errorf("Did not find contract on LIT2"))
	}

	fmt.Println("Settling contract on LIT1")
	// Fetching sig
	oracleValue, oracleSig, err := oracle.FetchSignature(contractOne.OracleR)
	handleErrorIfNeeded(err)

	fmt.Printf("Settling contract on [%d] (sig: %x)\n", oracleValue, oracleSig)

	_, err = commands.SettleContract(rpcCon1, contractOne.Idx, oracleValue, oracleSig)
	handleErrorIfNeeded(err)

	// This makes LIT1 update its balance with the settlement amount claimed. It will trigger LIT2 to claim its part
	mineBlockAndWait()

	// This makes LIT2 update its balance
	mineBlockAndWait()

	lit1Balance, err := commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)
	lit2Balance, err := commands.GetBalance(rpcCon2)
	handleErrorIfNeeded(err)

	fmt.Printf("Tests concluded. LIT1 [%d] - LIT2 [%d]\n", lit1Balance.Balances[0].MatureWitty, lit2Balance.Balances[0].MatureWitty)
}
