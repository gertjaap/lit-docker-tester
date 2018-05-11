package testscripts

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gertjaap/lit-docker-tester/commands"
)

func DlcTest() {
	fmt.Println("LIT DLC Tester Script")

	rpcCon1, rpcCon2, wsConn1, wsConn2 := ConnectAndFund()
	defer wsConn1.Close()
	defer wsConn2.Close()

	oIdx, _ := ImportOracle(rpcCon1, rpcCon2)

	fmt.Println("Creating contract on LIT1")
	contract, err := commands.NewContract(rpcCon1)
	handleErrorIfNeeded(err)

	cIdx := contract.Contract.Idx

	_, err = commands.SetContractOracle(rpcCon1, cIdx, oIdx)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractSettlementTime(rpcCon1, cIdx, 1524156300)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractDatafeed(rpcCon1, cIdx, 1)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractFunding(rpcCon1, cIdx, 5000000, 5000000)
	handleErrorIfNeeded(err)
	_, err = commands.SetContractDivision(rpcCon1, cIdx, 11950, 12100)
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
		if contract.TheirIdx == cIdx {
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
	time.Sleep(time.Millisecond * 2000)

	// After accept the contract "TheirIdx" should be properly filled on LIT1
	found := false
	contracts, err := commands.ListContracts(rpcCon1)
	for _, contract := range contracts.Contracts {
		if contract.TheirIdx == cIdx2 {
			found = true
			break
		}
	}

	if !found {
		handleErrorIfNeeded(fmt.Errorf("Could not find a contract with TheirIdx=%d on LIT1", cIdx2))
	}

	time.Sleep(time.Millisecond * 2000)

	fmt.Println("Checking if contract status on LIT1 is 'Active'")
	contract1, err := commands.GetContract(rpcCon1, cIdx)
	handleErrorIfNeeded(err)

	if contract1.Contract.Status != commands.ContractStatusActive {
		handleErrorIfNeeded(fmt.Errorf("Contract status on LIT1 not Active"))
	}

	mineBlockAndWait()

	fmt.Println("Checking if balances reflect the contract funding")

	lit1Balance, err := commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)

	if lit1Balance.Balances[0].CoinType == 257 && lit1Balance.Balances[0].MatureWitty == initialFundingAmount-5000500 {
		fmt.Println("LIT1 shows correct funds")
	} else {
		handleErrorIfNeeded(fmt.Errorf("LIT1 did not have the correct balance or cointype: %d / %d", lit1Balance.Balances[0].CoinType, lit1Balance.Balances[0].MatureWitty))
	}

	lit2Balance, err := commands.GetBalance(rpcCon2)
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

	// This makes LIT1 update its balance with the settlement amount claimed. It will trigger LIT2 to claim its part
	mineBlockAndWait()

	// This makes LIT2 update its balance
	mineBlockAndWait()

	lit1Balance, err = commands.GetBalance(rpcCon1)
	handleErrorIfNeeded(err)
	lit2Balance, err = commands.GetBalance(rpcCon2)
	handleErrorIfNeeded(err)

	if lit1Balance.Balances[0].MatureWitty == 16331834 && lit2Balance.Balances[0].MatureWitty == 23665166 {
		fmt.Println("Concluded tests succesfully.")
	} else {
		fmt.Printf("Wrong balances. LIT1 [%d] - LIT2 [%d]\n", lit1Balance.Balances[0].MatureWitty, lit2Balance.Balances[0].MatureWitty)
	}
}
