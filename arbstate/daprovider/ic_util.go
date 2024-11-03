// package daprovider

// import (
// 	"errors"
// 	"fmt"

// 	"github.com/aviate-labs/agent-go"
// 	agentcert "github.com/aviate-labs/agent-go/certification"
// 	"github.com/aviate-labs/agent-go/certification/hashtree"
// 	"github.com/aviate-labs/agent-go/principal"
// 	"github.com/fxamacker/cbor/v2"
// )

// var (
// 	DefaultCanister = principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai")
// )

// type CertifiedBlock struct {
// 	Certificate []byte `ic:"certificate" json:"certificate"`
// 	Data        []byte `ic:"data" json:"data"`
// 	Witness     []byte `ic:"witness" json:"witness"`
// }

// type CertifiedBlockECDSA struct {
// 	Signature []byte `ic:"signature" json:"signature"`
// 	Block     string `ic:"block" json:"block"`
// }

// type Object struct {
// 	Data []byte `ic:"data" json:"data"`
// }

// type StorageReceipt struct {
// 	LeafHash []byte `ic:"leaf_hash" json:"leaf_hash"`
// 	RootHash []byte `ic:"root_hash" json:"root_hash"`
// }

// type StorageReceiptECDSA struct {
// 	Signature []byte `ic:"signature" json:"signature"`
// }

// // Agent is a client for the "test" canister.
// type Agent struct {
// 	*agent.Agent
// 	CanisterId principal.Principal
// }

// func NewAgent(canisterId principal.Principal, config agent.Config) (*Agent, error) {
// 	a, err := agent.New(config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Agent{
// 		Agent:      a,
// 		CanisterId: canisterId,
// 	}, nil
// }

// func (a Agent) Fetch(arg0 string) (*CertifiedBlock, error) {
// 	var r0 CertifiedBlock
// 	if err := a.Agent.Query(
// 		a.CanisterId,
// 		"fetch",
// 		[]any{arg0},
// 		[]any{&r0},
// 	); err != nil {
// 		return nil, err
// 	}
// 	return &r0, nil
// }

// func (a Agent) Store(arg0 string, arg1 []byte) (*StorageReceipt, error) {
// 	var r0 StorageReceipt
// 	if err := a.Agent.Call(
// 		a.CanisterId,
// 		"store",
// 		[]any{arg0, Object{
// 			Data: arg1,
// 		}},
// 		[]any{&r0},
// 	); err != nil {
// 		panic(err)
// 	}
// 	return &r0, nil
// }

// func VerifyDataFromIC(certificate []byte, rootKey []byte, canister principal.Principal, witness []byte) error {

// 	var c agentcert.Certificate
// 	if err := cbor.Unmarshal(certificate, &c); err != nil {
// 		return err
// 	}

// 	if err := agentcert.VerifyCertificate(c, canister, rootKey); err != nil {
// 		return err
// 	}

// 	providedRootHash, err := c.Tree.Lookup(
// 		hashtree.Label("canister"),
// 		canister.Raw,
// 		hashtree.Label("certified_data"))
// 	if err != nil {
// 		return err
// 	}

// 	ht, err := hashtree.Deserialize(witness)
// 	if err != nil {
// 		return err
// 	}

// 	var rootHash [32]byte
// 	copy(rootHash[:], providedRootHash)

// 	witnessHash := ht.Reconstruct()
// 	if witnessHash != rootHash {
// 		return errors.New(fmt.Sprintf("witness hash %x doesn't match known root hash %x", witnessHash, rootHash))
// 	}

// 	return nil
// }

// func VerifyCertificate(certificate []byte, rootKey []byte, canister principal.Principal) error {

// 	var c agentcert.Certificate
// 	if err := cbor.Unmarshal(certificate, &c); err != nil {
// 		return err
// 	}

// 	if err := agentcert.VerifyCertificate(c, canister, rootKey); err != nil {
// 		return err
// 	}

// 	return nil
// }
