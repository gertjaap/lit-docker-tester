package commands

import (
	"net/rpc"

	"github.com/btcsuite/btcd/wire"
)

type ListOraclesArgs struct {
	// none
}

type ListOraclesReply struct {
	Oracles []*DlcOracle
}

type DlcOracle struct {
	Idx  uint64   // Index of the oracle for refencing in commands
	A    [33]byte // public key of the oracle
	Name string   // Name of the oracle for display purposes
	Url  string   // Base URL of the oracle, if its REST based (optional)
}

type ImportOracleArgs struct {
	Url  string
	Name string
}

type ImportOracleReply struct {
	Oracle *DlcOracle
}

type AddOracleArgs struct {
	Key  string
	Name string
}

type AddOracleReply struct {
	Oracle *DlcOracle
}

type DlcContractStatus int

const (
	ContractStatusDraft        DlcContractStatus = 0
	ContractStatusOfferedByMe  DlcContractStatus = 1
	ContractStatusOfferedToMe  DlcContractStatus = 2
	ContractStatusDeclined     DlcContractStatus = 3
	ContractStatusAccepted     DlcContractStatus = 4
	ContractStatusAcknowledged DlcContractStatus = 5
	ContractStatusActive       DlcContractStatus = 6
	ContractStatusClosed       DlcContractStatus = 7
)

type DlcContractSettlementSignature struct {
	Outcome   int64    // The oracle value for which transaction these are the signatures
	Signature [64]byte // The signature for the transaction
}

type DlcContract struct {
	Idx                                      uint64                           // Index of the contract for referencing in commands
	PeerIdx                                  uint32                           // Index of the peer we've offered the contract to or received the contract from
	PubKey                                   [33]byte                         // Key of the contract
	CoinType                                 uint32                           // Coin type
	OracleA, OracleR                         [33]byte                         // Pub keys of the oracle
	OracleTimestamp                          uint64                           // The time we expect the oracle to publish
	Division                                 []DlcContractDivision            // The payout specification
	OurFundingAmount, TheirFundingAmount     int64                            // The amounts either side are funding
	OurChangePKH, TheirChangePKH             [20]byte                         // PKH to which the contracts funding change should go
	OurFundMultisigPub, TheirFundMultisigPub [33]byte                         // Pubkey used in the funding multisig output
	OurPayoutPub, TheirPayoutPub             [33]byte                         // Pubkey to which the contracts are supposed to pay out
	Status                                   DlcContractStatus                // Status of the contract
	OurFundingInputs, TheirFundingInputs     []DlcContractFundingInput        // Outpoints used to fund the contract
	TheirSettlementSignatures                []DlcContractSettlementSignature // Signatures for the settlement transactions
}

type DlcContractDivision struct {
	OracleValue int64
	ValueOurs   int64
}

type DlcContractFundingInput struct {
	Outpoint wire.OutPoint
	Value    int64
}

type NewContractArgs struct {
	// empty
}

type NewContractReply struct {
	Contract *DlcContract
}

type ListContractsArgs struct {
	// none
}

type ListContractsReply struct {
	Contracts []*DlcContract
}

type GetContractArgs struct {
	Idx uint64
}

type GetContractReply struct {
	Contract *DlcContract
}

type SetContractOracleArgs struct {
	CIdx uint64
	OIdx uint64
}

type SetContractOracleReply struct {
	Success bool
}

type SetContractDatafeedArgs struct {
	CIdx uint64
	Feed uint64
}

type SetContractDatafeedReply struct {
	Success bool
}

type SetContractRPointArgs struct {
	CIdx   uint64
	RPoint [33]byte
}

type SetContractRPointReply struct {
	Success bool
}

type SetContractSettlementTimeArgs struct {
	CIdx uint64
	Time uint64
}

type SetContractSettlementTimeReply struct {
	Success bool
}

type SetContractFundingArgs struct {
	CIdx        uint64
	OurAmount   int64
	TheirAmount int64
}

type SetContractFundingReply struct {
	Success bool
}

type SetContractSettlementDivisionArgs struct {
	CIdx             uint64
	ValueFullyOurs   int64
	ValueFullyTheirs int64
}

type SetContractSettlementDivisionReply struct {
	Success bool
}

type SetContractCoinTypeArgs struct {
	CIdx     uint64
	CoinType uint32
}

type SetContractCoinTypeReply struct {
	Success bool
}

type OfferContractArgs struct {
	CIdx    uint64
	PeerIdx uint32
}

type OfferContractReply struct {
	Success bool
}

type DeclineContractArgs struct {
	CIdx uint64
}

type DeclineContractReply struct {
	Success bool
}

type AcceptContractArgs struct {
	CIdx uint64
}

type AcceptContractReply struct {
	Success bool
}

type SettleContractArgs struct {
	CIdx        uint64
	OracleValue int64
	OracleSig   [32]byte
}

type SettleContractReply struct {
	Success bool
}

func ImportOracle(c *rpc.Client, url, name string) (*ImportOracleReply, error) {
	args := new(ImportOracleArgs)
	args.Url = url
	args.Name = name

	reply := new(ImportOracleReply)
	err := c.Call("LitRPC.ImportOracle", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func NewContract(c *rpc.Client) (*NewContractReply, error) {
	args := new(NewContractArgs)

	reply := new(NewContractReply)
	err := c.Call("LitRPC.NewContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func ListOracles(c *rpc.Client) (*ListOraclesReply, error) {
	args := new(ListOraclesArgs)

	reply := new(ListOraclesReply)
	err := c.Call("LitRPC.ListOracles", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func AddOracle(c *rpc.Client, key, name string) (*AddOracleReply, error) {
	args := new(AddOracleArgs)
	args.Key = key
	args.Name = name
	reply := new(AddOracleReply)
	err := c.Call("LitRPC.AddOracle", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func ListContracts(c *rpc.Client) (*ListContractsReply, error) {
	args := new(ListContractsArgs)

	reply := new(ListContractsReply)
	err := c.Call("LitRPC.ListContracts", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func GetContract(c *rpc.Client, idx uint64) (*GetContractReply, error) {
	args := new(GetContractArgs)
	args.Idx = idx
	reply := new(GetContractReply)
	err := c.Call("LitRPC.GetContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractOracle(c *rpc.Client, cIdx, oIdx uint64) (*SetContractOracleReply, error) {
	args := new(SetContractOracleArgs)
	args.CIdx = cIdx
	args.OIdx = oIdx
	reply := new(SetContractOracleReply)
	err := c.Call("LitRPC.SetContractOracle", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractDatafeed(c *rpc.Client, cIdx, feed uint64) (*SetContractDatafeedReply, error) {
	args := new(SetContractDatafeedArgs)
	args.CIdx = cIdx
	args.Feed = feed
	reply := new(SetContractDatafeedReply)
	err := c.Call("LitRPC.SetContractDatafeed", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractRPoint(c *rpc.Client, cIdx uint64, rPoint [33]byte) (*SetContractRPointReply, error) {
	args := new(SetContractRPointArgs)
	args.CIdx = cIdx
	args.RPoint = rPoint
	reply := new(SetContractRPointReply)
	err := c.Call("LitRPC.SetContractRPoint", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractSettlementTime(c *rpc.Client, cIdx, time uint64) (*SetContractSettlementTimeReply, error) {
	args := new(SetContractSettlementTimeArgs)
	args.CIdx = cIdx
	args.Time = time
	reply := new(SetContractSettlementTimeReply)
	err := c.Call("LitRPC.SetContractSettlementTime", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractFunding(c *rpc.Client, cIdx uint64, ours, theirs int64) (*SetContractFundingReply, error) {
	args := new(SetContractFundingArgs)
	args.CIdx = cIdx
	args.OurAmount = ours
	args.TheirAmount = theirs
	reply := new(SetContractFundingReply)
	err := c.Call("LitRPC.SetContractFunding", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractSettlementDivision(c *rpc.Client, cIdx uint64, allOurs, allTheirs int64) (*SetContractSettlementDivisionReply, error) {
	args := new(SetContractSettlementDivisionArgs)
	args.CIdx = cIdx
	args.ValueFullyOurs = allOurs
	args.ValueFullyTheirs = allTheirs
	reply := new(SetContractSettlementDivisionReply)
	err := c.Call("LitRPC.SetContractSettlementDivision", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SetContractCoinType(c *rpc.Client, cIdx uint64, coinType uint32) (*SetContractCoinTypeReply, error) {
	args := new(SetContractCoinTypeArgs)
	args.CIdx = cIdx
	args.CoinType = coinType
	reply := new(SetContractCoinTypeReply)
	err := c.Call("LitRPC.SetContractCoinType", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func OfferContract(c *rpc.Client, cIdx uint64, peerIdx uint32) (*OfferContractReply, error) {
	args := new(OfferContractArgs)
	args.CIdx = cIdx
	args.PeerIdx = peerIdx
	reply := new(OfferContractReply)
	err := c.Call("LitRPC.OfferContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func AcceptContract(c *rpc.Client, cIdx uint64) (*AcceptContractReply, error) {
	args := new(AcceptContractArgs)
	args.CIdx = cIdx
	reply := new(AcceptContractReply)
	err := c.Call("LitRPC.AcceptContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func DeclineContract(c *rpc.Client, cIdx uint64) (*DeclineContractReply, error) {
	args := new(DeclineContractArgs)
	args.CIdx = cIdx
	reply := new(DeclineContractReply)
	err := c.Call("LitRPC.DeclineContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func SettleContract(c *rpc.Client, cIdx uint64, settleValue int64, oracleSig [32]byte) (*SettleContractReply, error) {
	args := new(SettleContractArgs)
	args.CIdx = cIdx
	args.OracleValue = settleValue
	args.OracleSig = oracleSig
	reply := new(SettleContractReply)
	err := c.Call("LitRPC.SettleContract", args, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
