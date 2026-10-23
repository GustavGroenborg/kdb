package dar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gcrg/kdb/service/pkg/kommune"
)

var (
	ErrNoToken         = errors.New("DAR API token not configured")
	ErrAddressNotFound = errors.New("address not found in DAR")
)

const (
	DefaultEndpoint               = "https://graphql.datafordeler.dk/DAR/v3"
	DefaultAdressevaelgerEndpoint = "https://adressevaelger.dk/adresser/soeg"
	DefaultFallbackToken          = "adressevaelger123"
)

var postalCodeRegex = regexp.MustCompile(`\b([1-9][0-9]{3})\b`)

// LoadDARToken retrieves the Datafordeleren DAR API token according to security standards:
// 1. systemd credentials directory ($CREDENTIALS_DIRECTORY/dar_token)
// 2. DAR_API_TOKEN environment variable
// 3. .env file in the current working directory
func LoadDARToken() string {
	if credDir := os.Getenv("CREDENTIALS_DIRECTORY"); credDir != "" {
		if data, err := os.ReadFile(filepath.Join(credDir, "dar_token")); err == nil {
			token := strings.TrimSpace(string(data))
			if token != "" {
				return token
			}
		}
	}
	if token := strings.TrimSpace(os.Getenv("DAR_API_TOKEN")); token != "" {
		return token
	}
	if data, err := os.ReadFile(".env"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DAR_API_TOKEN=") {
				val := strings.TrimPrefix(line, "DAR_API_TOKEN=")
				val = strings.TrimSpace(val)
				val = strings.Trim(val, `"'`)
				if val != "" {
					return val
				}
			}
		}
	}
	return ""
}

type Client struct {
	Endpoint               string
	AdressevaelgerEndpoint string
	Token                  string
	HTTPClient             *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Endpoint:               DefaultEndpoint,
		AdressevaelgerEndpoint: DefaultAdressevaelgerEndpoint,
		Token:                  token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AutocompleteSuggestion struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type adressevaelgerResponse struct {
	Status      string `json:"status"`
	Beskrivelse string `json:"beskrivelse,omitempty"`
	Fund        []struct {
		ID    string `json:"id"`
		Titel string `json:"titel"`
	} `json:"fund"`
}

type AddressResult struct {
	ResolvedAddress string `json:"resolved_address"`
	KommuneCode     string `json:"kommune_code"`
	KommuneName     string `json:"kommune_name"`
	IsYderkommune   bool   `json:"is_yderkommune"`
}

type nodeItem struct {
	Adgangsadressebetegnelse string `json:"adgangsadressebetegnelse"`
	Vejmidte                 string `json:"vejmidte"`
	Status                   string `json:"status"`
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphQLResponse struct {
	Data struct {
		DARHusnummer struct {
			Nodes []nodeItem `json:"nodes"`
		} `json:"DAR_Husnummer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

const lookupQuery = `
query LookupAddress($address: String!, $now: DafDateTime!, $first: Int!) {
  DAR_Husnummer(
    virkningstid: $now
    first: $first
    where: {
      adgangsadressebetegnelse: { startsWith: $address }
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

// executeQuery sends a query to the DAR GraphQL API.
func (c *Client) executeQuery(ctx context.Context, search string, first int) ([]nodeItem, error) {
	nowISO := time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")

	reqPayload := graphQLRequest{
		Query: lookupQuery,
		Variables: map[string]interface{}{
			"address": search,
			"now":     nowISO,
			"first":   first,
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	endpoint := c.Endpoint
	if strings.Contains(endpoint, "graphql.datafordeler.dk") && !strings.Contains(endpoint, "apiKey=") {
		if strings.Contains(endpoint, "?") {
			endpoint += "&apiKey=" + c.Token
		} else {
			endpoint += "?apiKey=" + c.Token
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBytes))
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

	return gqlResp.Data.DARHusnummer.Nodes, nil
}

// LookupAddress queries the DAR GraphQL API to resolve the given address to a Danish municipality.
// It supports standard addresses as well as addresses containing a supplerende bynavn (e.g. Mjallerup in 9670 Løgstør).
func (c *Client) LookupAddress(ctx context.Context, address string) (*AddressResult, error) {
	if strings.TrimSpace(c.Token) == "" {
		return nil, ErrNoToken
	}

	trimmedAddress := strings.TrimSpace(address)
	if trimmedAddress == "" {
		return nil, ErrAddressNotFound
	}

	// 1. First attempt: exact startsWith match with full input
	nodes, err := c.executeQuery(ctx, trimmedAddress, 10)
	if err != nil {
		return nil, err
	}

	// 2. Second attempt: address might have an omitted supplerende bynavn in user input.
	// E.g. "Højgårdsvej 9, 9670 Løgstør" where official DAR address is "Højgårdsvej 9, Mjallerup, 9670 Løgstør".
	if len(nodes) == 0 {
		postcodeMatch := postalCodeRegex.FindString(trimmedAddress)

		var streetPrefix string
		if commaIdx := strings.Index(trimmedAddress, ","); commaIdx > 0 {
			streetPrefix = strings.TrimSpace(trimmedAddress[:commaIdx])
		} else if postcodeMatch != "" {
			idx := strings.Index(trimmedAddress, postcodeMatch)
			streetPrefix = strings.TrimSpace(trimmedAddress[:idx])
		} else {
			streetPrefix = trimmedAddress
		}

		if streetPrefix != "" {
			// Query with comma e.g. "Højgårdsvej 9, "
			candidates, err := c.executeQuery(ctx, streetPrefix+", ", 500)
			if err == nil && len(candidates) > 0 {
				if postcodeMatch != "" {
					for _, cand := range candidates {
						if strings.Contains(cand.Adgangsadressebetegnelse, postcodeMatch) {
							nodes = []nodeItem{cand}
							break
						}
					}
				} else {
					nodes = candidates
				}
			}
		}
	}

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

// Autocomplete queries Klimadatastyrelsen's Adressevælgeren to provide fast, typo-tolerant type-ahead address suggestions.
func (c *Client) Autocomplete(ctx context.Context, query string) ([]AutocompleteSuggestion, error) {
	trimmed := strings.TrimSpace(query)
	if len(trimmed) < 2 {
		return []AutocompleteSuggestion{}, nil
	}

	token := strings.TrimSpace(c.Token)
	if token == "" {
		token = DefaultFallbackToken
	}

	endpoint := c.AdressevaelgerEndpoint
	if endpoint == "" {
		endpoint = DefaultAdressevaelgerEndpoint
	}

	reqURL := fmt.Sprintf("%s?tekst=%s&token=%s", endpoint, url.QueryEscape(trimmed), url.QueryEscape(token))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create autocomplete request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("autocomplete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("autocomplete HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var res adressevaelgerResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode autocomplete response: %w", err)
	}

	suggestions := make([]AutocompleteSuggestion, 0, len(res.Fund))
	for _, item := range res.Fund {
		if item.Titel != "" {
			suggestions = append(suggestions, AutocompleteSuggestion{
				ID:   item.ID,
				Text: item.Titel,
			})
		}
	}

	return suggestions, nil
}
