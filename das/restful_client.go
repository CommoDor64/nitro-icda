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

	// // ICDA: verify L2 data coming from ic
	// {
	// 	p := principal.MustDecode(string(DefaultTestStorageConfig.Canister))

	// 	rootKey := []byte{48, 129, 130, 48, 29, 6, 13, 43, 6, 1, 4, 1, 130, 220, 124, 5, 3, 1, 2, 1, 6, 12, 43, 6, 1, 4, 1, 130, 220, 124, 5, 3, 2, 1, 3, 97, 0, 184, 144, 225, 237, 54, 1, 156, 0, 29, 86, 172, 96, 112, 82, 127, 156, 217, 69, 27, 159, 9, 247, 116, 100, 118, 209, 60, 101, 39, 123, 245, 82, 195, 182, 241, 240, 99, 235, 171, 47, 14, 225, 49, 212, 247, 215, 132, 7, 18, 94, 207, 28, 236, 177, 68, 227, 152, 124, 115, 163, 111, 243, 238, 230, 164, 78, 211, 201, 249, 27, 24, 60, 9, 66, 118, 198, 80, 221, 35, 98, 208, 68, 234, 146, 135, 114, 60, 91, 234, 134, 235, 155, 152, 156, 171, 12}

	// 	if err := VerifyDataFromIC(response.Certificate, rootKey, p, response.Witness); err != nil {
	// 		return nil, err
	// 	}
	// }

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
