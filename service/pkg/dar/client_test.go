package dar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupAddress_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		response := `{
			"data": {
				"DAR_Husnummer": {
					"nodes": [
						{
							"adgangsadressebetegnelse": "Strandvejen 1, 4800 Nykøbing F",
							"vejmidte": "0376-0042",
							"status": "gældende"
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	client := NewClient("test-token")
	client.Endpoint = ts.URL

	res, err := client.LookupAddress(context.Background(), "Strandvejen 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.KommuneCode != "0376" {
		t.Errorf("got KommuneCode %s, want 0376", res.KommuneCode)
	}
	if res.KommuneName != "Guldborgsund" {
		t.Errorf("got KommuneName %s, want Guldborgsund", res.KommuneName)
	}
	if !res.IsYderkommune {
		t.Errorf("got IsYderkommune false, want true")
	}
}

func TestLookupAddress_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"DAR_Husnummer": {
					"nodes": []
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	client := NewClient("test-token")
	client.Endpoint = ts.URL

	_, err := client.LookupAddress(context.Background(), "Nonexistent Address 999")
	if err != ErrAddressNotFound {
		t.Fatalf("got error %v, want ErrAddressNotFound", err)
	}
}

func TestLookupAddress_NoToken(t *testing.T) {
	client := NewClient("")
	_, err := client.LookupAddress(context.Background(), "Strandvejen 1")
	if err != ErrNoToken {
		t.Fatalf("got error %v, want ErrNoToken", err)
	}
}

func TestLookupAddress_SupplerendeBynavn(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock query that checks if it was called with full address or prefix
		var req struct {
			Variables struct {
				Address string `json:"address"`
			} `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var response string
		if req.Variables.Address == "Højgårdsvej 9, 9670 Løgstør" {
			// First attempt fails because DAR has Mjallerup in between
			response = `{"data":{"DAR_Husnummer":{"nodes":[]}}}`
		} else {
			// Second attempt with prefix returns the node with Mjallerup
			response = `{
				"data": {
					"DAR_Husnummer": {
						"nodes": [
							{
								"adgangsadressebetegnelse": "Højgårdsvej 9, 3520 Farum",
								"vejmidte": "0190-0294",
								"status": "gældende"
							},
							{
								"adgangsadressebetegnelse": "Højgårdsvej 9, Mjallerup, 9670 Løgstør",
								"vejmidte": "0820-0587",
								"status": "gældende"
							}
						]
					}
				}
			}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	client := NewClient("test-token")
	client.Endpoint = ts.URL

	res, err := client.LookupAddress(context.Background(), "Højgårdsvej 9, 9670 Løgstør")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.KommuneCode != "0820" {
		t.Errorf("got KommuneCode %s, want 0820", res.KommuneCode)
	}
	if res.KommuneName != "Vesthimmerland" {
		t.Errorf("got KommuneName %s, want Vesthimmerland", res.KommuneName)
	}
	if !res.IsYderkommune {
		t.Errorf("got IsYderkommune false, want true")
	}
}

