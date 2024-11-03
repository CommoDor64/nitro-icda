package das

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"icdaserver/icutils"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/offchainlabs/nitro/arbstate/daprovider"
	"github.com/offchainlabs/nitro/das/dastree"
	flag "github.com/spf13/pflag"
)

type ExpirationPolicy struct{}
type CanisterId string

type ICStorageConfig struct {
	Enable   bool       `koanf:"enable"`
	Network  string     `konaf:"network"`
	Canister CanisterId `konaf:"canister"`
}

var DefaultTestStorageConfig = ICStorageConfig{
	Enable:   true,
	Network:  "http://172.17.0.1:4943/",
	Canister: "bkyz2-fmaaa-aaaaa-qaaaq-cai",
}

func ICStorageConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", false, "enable the internet computer(ic) as a da layer")
	f.String(prefix+".network", DefaultTestStorageConfig.Network, "url of the ic network")
	f.String(prefix+".canister", string(DefaultTestStorageConfig.Canister), "readable canister ids")
}

type ICStorageService struct {
	Agent    *icutils.Agent
	Canister principal.Principal
	Cache    map[string]string
}

func NewICStorageService(config ICStorageConfig) (StorageService, error) {
	u, err := url.Parse(config.Network)
	if err != nil {
		return nil, err
	}

	aconfig := agent.Config{
		ClientConfig:                   &agent.ClientConfig{Host: u},
		FetchRootKey:                   true,
		DisableSignedQueryVerification: true,
	}

	p := principal.MustDecode(string(config.Canister))

	a, err := icutils.NewAgent(p, aconfig)
	if err != nil {
		return nil, err
	}

	return ICStorageService{
		Cache:    map[string]string{},
		Canister: p,
		Agent:    a,
	}, nil
}

func (s *ICStorageService) Read(ctx context.Context) error {
	return nil
}

func (s ICStorageService) Put(ctx context.Context, data []byte, expirationTime uint64) error {
	_, err := s.Agent.Store(
		dastree.Hash(data).Hex(),
		data)

	if err != nil {
		return nil
	}

	return nil
}

func (s ICStorageService) Sync(ctx context.Context) error {
	return nil
}

func (s ICStorageService) GetByHash(ctx context.Context, hash common.Hash) ([]byte, error) {

	blockHash := hash.Hex()
	if hash == [32]byte{} {
		return nil, errors.New(fmt.Sprintf("expected well formed hash, got %v", blockHash))
	}

	cb, err := s.Agent.Fetch(blockHash)
	if err != nil {
		panic(err)
	}

	if _, err := icutils.VerifyDataFromIC(cb.Certificate, s.Agent.GetRootKey(), s.Canister, cb.Witness, cb.Data); err != nil {
		return nil, err
	}

	b, err := json.Marshal(cb)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s ICStorageService) ExpirationPolicy(ctx context.Context) (daprovider.ExpirationPolicy, error) {
	return daprovider.KeepForever, nil
}

func (s ICStorageService) Close(ctx context.Context) error {
	return nil
}

func (s ICStorageService) String() string {
	return "ICStorageService"
}

func (s ICStorageService) HealthCheck(ctx context.Context) error {
	testData := []byte("Test-Data")
	err := s.Put(ctx, testData, uint64(time.Now().Add(time.Minute).Unix()))
	if err != nil {
		return err
	}
	res, err := s.GetByHash(ctx, dastree.Hash(testData))
	if err != nil {
		return err
	}
	if !bytes.Equal(res, testData) {
		return errors.New("invalid GetByHash result")
	}
	return nil
}
