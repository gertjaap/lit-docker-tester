package testscripts

import (
	"fmt"
	"net/rpc"

	"github.com/gertjaap/lit-docker-tester/commands"
)

func ImportOracle(rpcCon1 *rpc.Client, rpcCon2 *rpc.Client) (uint64, uint64) {
	fmt.Println("Importing oracle on LIT1")
	importResult, err := commands.ImportOracle(rpcCon1, "https://oracle.gertjaap.org/", "Gert-Jaap")
	handleErrorIfNeeded(err)

	if importResult.Oracle.Name != "Gert-Jaap" {
		handleErrorIfNeeded(fmt.Errorf("Oracle import failed. Name %s unexpected", importResult.Oracle.Name))
	}

	oIdx1 := importResult.Oracle.Idx

	fmt.Println("Importing oracle on LIT2")
	importResult, err = commands.ImportOracle(rpcCon2, "https://oracle.gertjaap.org/", "Gert-Jaap")
	handleErrorIfNeeded(err)

	if importResult.Oracle.Name != "Gert-Jaap" {
		handleErrorIfNeeded(fmt.Errorf("Oracle import failed. Name %s unexpected", importResult.Oracle.Name))
	}

	return oIdx1, importResult.Oracle.Idx
}
