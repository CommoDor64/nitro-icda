package das

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/offchainlabs/nitro/arbstate/daprovider"
	"github.com/offchainlabs/nitro/das/dastree"
	flag "github.com/spf13/pflag"
)

type ExpirationPolicy struct{}

type ICStorageConfig struct {
	Enable bool `koanf:"enable"`
}

func ICStorageConfigAddOptions(prefix string, f *flag.FlagSet) {
	f.Bool(prefix+".enable", DefaultLocalDBStorageConfig.Enable, "enable storage/retrieval of sequencer batch data from a database on the local filesystem")
}

type ICStorageService struct{}

func NewICStorageService(config ICStorageConfig) (StorageService, error) {
	return ICStorageService{}, nil
}

func (s *ICStorageService) Read(ctx context.Context) error {
	return nil
}

type (
	Account struct {
		Account string `ic:"account"`
	}

	Balance struct {
		E8S uint64 `ic:"e8s"`
	}
)

func (s ICStorageService) Put(ctx context.Context, data []byte, expirationTime uint64) error {
	return nil
}

func (s ICStorageService) Sync(ctx context.Context) error {
	return nil
}

func (s ICStorageService) GetByHash(ctx context.Context, hash common.Hash) ([]byte, error) {
	u, _ := url.Parse("http://host.docker.internal:40363")
	config := agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	}
	a, _ := agent.New(config)

	var balance Balance
	if err := a.Query(
		ic.LEDGER_PRINCIPAL, "account_balance_dfx",
		[]any{Account{"21e7a49f00b40e83c2eb17468070566d2d6fc8cf4ae2cf3f54dee335de10745b"}},
		[]any{&balance},
	); err != nil {
		log.Fatal(err)
	}

	_ = balance
	return nil, nil
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
