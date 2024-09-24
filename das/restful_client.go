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
	"strings"

	"github.com/aviate-labs/agent-go/principal"
	"github.com/ethereum/go-ethereum/common"
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

	// ICDA: verify L2 data coming from ic
	{
		p := principal.MustDecode(string(DefaultTestStorageConfig.Canister))

		rootKey := []byte{48, 129, 130, 48, 29, 6, 13, 43, 6, 1, 4, 1, 130, 220, 124, 5, 3, 1, 2, 1, 6, 12, 43, 6, 1, 4, 1, 130, 220, 124, 5, 3, 2, 1, 3, 97, 0, 151, 44, 207, 171, 16, 137, 198, 63, 87, 184, 84, 51, 254, 212, 167, 141, 232, 147, 119, 62, 104, 240, 46, 216, 20, 142, 37, 69, 85, 100, 94, 42, 170, 62, 155, 81, 217, 221, 1, 191, 15, 5, 36, 241, 199, 156, 13, 92, 18, 244, 75, 103, 202, 230, 240, 73, 21, 207, 253, 164, 169, 220, 115, 119, 235, 209, 55, 211, 41, 75, 219, 209, 247, 25, 252, 215, 179, 180, 52, 119, 219, 248, 96, 53, 215, 237, 203, 183, 179, 101, 31, 26, 10, 222, 191, 56}

		if err := VerifyDataFromIC(response.Certificate, rootKey, p, response.Witness); err != nil {
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
