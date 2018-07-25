package testscripts

import (
	"bytes"
	"fmt"

	"crypto/rand"

	"github.com/gertjaap/lit-docker-tester/commands"
	"github.com/gertjaap/lit/btcutil/btcec"
	"github.com/gertjaap/lit/sig64"
	"github.com/gertjaap/lit/wire"
	"github.com/mit-dci/lit/crypto/fastsha256"
)

func RemoteControlTest() {
	fmt.Println("LIT Remote control Tester Script")

	fmt.Println("Connecting to LIT nodes..")
	rpcConns, wsConns := ConnectAndFund()
	for _, wsConn := range wsConns {
		defer wsConn.Close()
	}

	ConnectTogether(rpcConns[0], rpcConns[1], "lit1")

	var pkBytes [32]byte
	rand.Read(pkBytes[:])
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes[:])

	pubBytes := pubKey.SerializeCompressed()
	var pubKeyBytes [33]byte
	copy(pubKeyBytes[:], pubBytes)

	fmt.Println("Authorizing pubkey for remote control")
	commands.RCAuth(rpcConns[0], pubKeyBytes, true)

	msg := []byte("{\"method\":\"Say\", \"args\":{\"Peer\":1, \"Message\":\"hello\"}}")
	hash := fastsha256.Sum256(msg)
	signature, _ := privKey.Sign(hash[:])
	sig, _ := sig64.SigCompress(signature.Serialize())
	var buf bytes.Buffer
	buf.WriteByte(0xB0)
	buf.Write(pubKeyBytes[:])
	buf.Write(sig[:])
	wire.WriteVarInt(&buf, 0, uint64(len(msg)))
	buf.Write(msg)
	request := buf.Bytes()
	fmt.Printf("Sending remote control command from node 2 to node 1: %x\n", request)
	commands.RCSend(rpcConns[1], 1, request)

	// alter pubkey to make it invalid
	request2 := make([]byte, len(request))
	copy(request2[:], request)
	request2[3] = request2[3] + 1

	fmt.Println("Sending invalid remote control command from node 2 to node 1")
	commands.RCSend(rpcConns[1], 1, request2)

	request3 := make([]byte, len(request))
	copy(request3[:], request)
	request3[66] = request2[3] + 1

	fmt.Println("Sending invalid signature remote control command from node 2 to node 1")
	commands.RCSend(rpcConns[1], 1, request3)

	fmt.Println("Done!")
}
