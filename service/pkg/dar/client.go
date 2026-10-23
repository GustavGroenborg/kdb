package dar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gcrg/kdb/service/pkg/kommune"
)

var (
	ErrNoToken         = errors.New("DAR API token not configured")
	ErrAddressNotFound = errors.New("address not found in DAR")
)

const DefaultEndpoint = "https://services.datafordeler.dk/DAR/DAR/1/rest/graphql"

// LoadDARToken retrieves the Datafordeleren DAR API token according to security standards:
// 1. systemd credentials directory ($CREDENTIALS_DIRECTORY/dar_token)
// 2. DAR_API_TOKEN environment variable (or .env file)
func LoadDARToken() string {
	if credDir := os.Getenv("CREDENTIALS_DIRECTORY"); credDir != "" {
		if data, err := os.ReadFile(filepath.Join(credDir, "dar_token")); err == nil {
			token := strings.TrimSpace(string(data))
			if token != "" {
				return token
			}
		}
	}
	return strings.TrimSpace(os.Getenv("DAR_API_TOKEN"))
}

type Client struct {
	Endpoint   string
	Token      string
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Endpoint: DefaultEndpoint,
		Token:    token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AddressResult struct {
	ResolvedAddress string `json:"resolved_address"`
	KommuneCode     string `json:"kommune_code"`
	KommuneName     string `json:"kommune_name"`
	IsYderkommune   bool   `json:"is_yderkommune"`
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphQLResponse struct {
	Data struct {
		DARHusnummer struct {
			Nodes []struct {
				Adgangsadressebetegnelse string `json:"adgangsadressebetegnelse"`
				Vejmidte                 string `json:"vejmidte"`
				Status                   string `json:"status"`
			} `json:"nodes"`
		} `json:"DAR_Husnummer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

const lookupQuery = `
query LookupAddress($address: String!, $now: DafDateTime!) {
  DAR_Husnummer(
    virkningstid: $now
    first: 5
    where: {
      adgangsadressebetegnelse: { startsWith: $address }
      status: { eq: "gældende" }
    }
  ) {
    nodes {
      adgangsadressebetegnelse
      vejmidte
      status
    }
  }
}
`

// LookupAddress queries the DAR GraphQL API to resolve the given address to a Danish municipality.
func (c *Client) LookupAddress(ctx context.Context, address string) (*AddressResult, error) {
	if strings.TrimSpace(c.Token) == "" {
		return nil, ErrNoToken
	}

	trimmedAddress := strings.TrimSpace(address)
	if trimmedAddress == "" {
		return nil, ErrAddressNotFound
	}

	nowISO := time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")

	reqPayload := graphQLRequest{
		Query: lookupQuery,
		Variables: map[string]interface{}{
			"address": trimmedAddress,
			"now":     nowISO,
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DAR request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DAR HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode DAR response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("DAR GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	nodes := gqlResp.Data.DARHusnummer.Nodes
	if len(nodes) == 0 {
		return nil, ErrAddressNotFound
	}

	// Extract Kommunekode from vejmidte (formatted as "kommunekode-vejkode", e.g. "0376-0042")
	var kommuneCode string
	var resolvedAddr string

	for _, node := range nodes {
		vm := strings.TrimSpace(node.Vejmidte)
		if vm != "" {
			parts := strings.Split(vm, "-")
			if len(parts) >= 1 && len(parts[0]) >= 3 {
				kommuneCode = parts[0]
				resolvedAddr = node.Adgangsadressebetegnelse
				break
			}
		}
	}

	if kommuneCode == "" {
		return nil, ErrAddressNotFound
	}

	name, isYder, found := kommune.LookupKommune(kommuneCode)
	if !found {
		name = fmt.Sprintf("Kommune %s", kommuneCode)
	}

	return &AddressResult{
		ResolvedAddress: resolvedAddr,
		KommuneCode:     kommuneCode,
		KommuneName:     name,
		IsYderkommune:   isYder,
	}, nil
}
