package daprovider

import (
	"errors"
	"fmt"

	agentcert "github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

type CertifiedBlock struct {
	Certificate []byte `ic:"certificate" json:"certificate"`
	Data        []byte `ic:"data" json:"data"`
	Witness     []byte `ic:"witness" json:"witness"`
}

var (
	DefaultCanister = principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai")
)

func VerifyDataFromIC(certificate []byte, rootKey []byte, canister principal.Principal, witness []byte) error {

	var c agentcert.Certificate
	if err := cbor.Unmarshal(certificate, &c); err != nil {
		return err
	}

	if err := agentcert.VerifyCertificate(c, canister, rootKey); err != nil {
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
func VerifyCertificate(certificate []byte, rootKey []byte, canister principal.Principal) error {

	var c agentcert.Certificate
	if err := cbor.Unmarshal(certificate, &c); err != nil {
		return err
	}

	if err := agentcert.VerifyCertificate(c, canister, rootKey); err != nil {
		return err
	}

	return nil
}
