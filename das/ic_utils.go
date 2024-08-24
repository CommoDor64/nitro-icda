package das

import (
	"errors"
	"fmt"

	"github.com/aviate-labs/agent-go"
	cert "github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

type CertifiedBlock struct {
	Certificate []byte `ic:"certificate" json:"certificate"`
	Data        []byte `ic:"data" json:"data"`
	Witness     []byte `ic:"witness" json:"witness"`
}

type CertifiedBlockECDSA struct {
	Signature []byte `ic:"signature" json:"signature"`
	Block     string `ic:"block" json:"block"`
}

type Object struct {
	Data []byte `ic:"data" json:"data"`
}

type StorageReceipt struct {
	LeafHash []byte `ic:"leaf_hash" json:"leaf_hash"`
	RootHash []byte `ic:"root_hash" json:"root_hash"`
}

type StorageReceiptECDSA struct {
	Signature []byte `ic:"signature" json:"signature"`
}

// Agent is a client for the "test" canister.
type Agent struct {
	*agent.Agent
	CanisterId principal.Principal
}

// NewAgent creates a new agent for the "test" canister.
func NewAgent(canisterId principal.Principal, config agent.Config) (*Agent, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}
	return &Agent{
		Agent:      a,
		CanisterId: canisterId,
	}, nil
}

// Fetch calls the "fetch" method on the "test" canister.
func (a Agent) Fetch(arg0 string) (*CertifiedBlock, error) {
	var r0 CertifiedBlock
	if err := a.Agent.Query(
		a.CanisterId,
		"fetch",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// FetchEcdsa calls the "fetch_ecdsa" method on the "test" canister.
func (a Agent) FetchEcdsa(arg0 string) (**CertifiedBlockECDSA, error) {
	var r0 *CertifiedBlockECDSA
	if err := a.Agent.Query(
		a.CanisterId,
		"fetch_ecdsa",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Store calls the "store" method on the "test" canister.
func (a Agent) Store(arg0 string, arg1 Object) (*StorageReceipt, error) {
	fmt.Println("arg0:", arg0)
	fmt.Println("arg1:", arg1)
	fmt.Println("arg1.Data:", arg1.Data)
	var r0 StorageReceipt
	if err := a.Agent.Call(
		a.CanisterId,
		"store",
		[]any{arg0, arg1},
		[]any{&r0},
	); err != nil {
		panic(err)
	}
	return &r0, nil
}

// StoreEcdsa calls the "store_ecdsa" method on the "test" canister.
func (a Agent) StoreEcdsa(arg0 string, arg1 string) (*StorageReceiptECDSA, error) {
	var r0 StorageReceiptECDSA
	if err := a.Agent.Call(
		a.CanisterId,
		"store_ecdsa",
		[]any{arg0, arg1},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func VerifyDataFromIC(certificate []byte, rootKey []byte, canister principal.Principal, witness []byte) error {

	var c cert.Certificate
	if err := cbor.Unmarshal(certificate, &c); err != nil {
		return err
	}

	if err := cert.VerifyCertificate(c, canister, rootKey); err != nil {
		return err
	}

	providedRootHash, err := c.Tree.Lookup(
		hashtree.Label("canister"),
		canister.Raw,
		hashtree.Label("certified_data"))
	if err != nil {
		return err
	}

	ht, err := hashtree.Deserialize(witness)
	if err != nil {
		return err
	}

	//FIXME! check inclusion!!

	// h := sha256.New()
	// h.Write(cb.Data)
	// dataHash := h.Sum(nil)

	// if _, err := hashtree.Lookup(ht, hashtree.Label(hex.EncodeToString(dataHash))); err != nil {
	// 	panic(errors.New(fmt.Sprintf("couldn't find hash %x in hashtree", dataHash)))
	// }

	var rootHash [32]byte
	copy(rootHash[:], providedRootHash)

	witnessHash := ht.Reconstruct()
	if witnessHash != rootHash {
		return errors.New(fmt.Sprintf("witness hash %x doesn't match known root hash %x", witnessHash, rootHash))
	}

	return nil
}
