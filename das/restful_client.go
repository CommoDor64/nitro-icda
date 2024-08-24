// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package das

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fxamacker/cbor/v2"
	"github.com/offchainlabs/nitro/arbstate/daprovider"
	"github.com/offchainlabs/nitro/das/dastree"
)

// RestfulDasClient implements daprovider.DASReader
type RestfulDasClient struct {
	url string
}

func NewRestfulDasClient(protocol string, host string, port int) *RestfulDasClient {
	return &RestfulDasClient{
		url: fmt.Sprintf("%s://%s:%d", protocol, host, port),
	}
}

func NewRestfulDasClientFromURL(url string) (*RestfulDasClient, error) {
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return nil, fmt.Errorf("protocol prefix 'http://' or 'https://' must be specified for RestfulDasClient; got '%s'", url)

	}
	return &RestfulDasClient{
		url: url,
	}, nil
}

func (c *RestfulDasClient) GetByHash(ctx context.Context, hash common.Hash) ([]byte, error) {
	fmt.Println(c.url + getByHashRequestPath + hash.Hex())
	prefixHash := hash.Hex()
	if len(prefixHash) == 64 {
		prefixHash = "0x" + prefixHash
	}

	res, err := http.Get(c.url + getByHashRequestPath + prefixHash)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error with status %d returned by server: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response RestfulDasServerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(response.Data)))
	decodedBytes, err := io.ReadAll(decoder)
	if err != nil {
		return nil, err
	}

	if !dastree.ValidHash(hash, decodedBytes) {
		return nil, daprovider.ErrHashMismatch
	}

	// This block is IC verification specific
	{

		u, err := url.Parse(DefaultTestStorageConfig.Network)
		if err != nil {
			return nil, err
		}

		aconfig := agent.Config{
			ClientConfig:                   &agent.ClientConfig{Host: u},
			FetchRootKey:                   true,
			DisableSignedQueryVerification: true,
		}

		p := principal.MustDecode(string(DefaultTestStorageConfig.Canister))

		a, err := NewAgent(p, aconfig)
		if err != nil {
			return nil, err
		}

		rootKey := a.GetRootKey()

		var cb CertifiedBlock
		if err := cbor.Unmarshal(decodedBytes, &c); err != nil {
			return nil, err
		}

		if err := VerifyDataFromIC(cb.Certificate, rootKey, p, cb.Witness); err != nil {
			return nil, err
		}
	}

	return decodedBytes, nil
}

func (c *RestfulDasClient) HealthCheck(ctx context.Context) error {
	res, err := http.Get(c.url + healthRequestPath)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error with status %d returned by server: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return nil
}

func (c *RestfulDasClient) ExpirationPolicy(ctx context.Context) (daprovider.ExpirationPolicy, error) {
	res, err := http.Get(c.url + expirationPolicyRequestPath)
	if err != nil {
		return -1, err
	}
	if res.StatusCode != http.StatusOK {
		return -1, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, fmt.Errorf("HTTP error with status %d returned by server: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	var response RestfulDasServerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return -1, err
	}

	return daprovider.StringToExpirationPolicy(response.ExpirationPolicy)
}
