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


func TestAutocomplete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tekst") != "Højgårdsvej 9, 9670" {
			t.Errorf("unexpected query text: %s", r.URL.Query().Get("tekst"))
		}
		if r.URL.Query().Get("token") != "custom-tok" {
			t.Errorf("unexpected token: %s", r.URL.Query().Get("token"))
		}

		response := `{
			"status": "ok",
			"fund": [
				{
					"type": "adresse",
					"id": "addr-1",
					"titel": "Højgårdsvej 9, Mjallerup, 9670 Løgstør",
					"husnummerId": "hus-1"
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	client := NewClient("custom-tok")
	client.AdressevaelgerEndpoint = ts.URL

	suggestions, err := client.Autocomplete(context.Background(), "Højgårdsvej 9, 9670")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if suggestions[0].Text != "Højgårdsvej 9, Mjallerup, 9670 Løgstør" {
		t.Errorf("got text %q, want 'Højgårdsvej 9, Mjallerup, 9670 Løgstør'", suggestions[0].Text)
	}
}

func TestAutocomplete_FallbackToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != DefaultFallbackToken {
			t.Errorf("expected fallback token %s, got %s", DefaultFallbackToken, r.URL.Query().Get("token"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "fund": []}`))
	}))
	defer ts.Close()

	client := NewClient("")
	client.AdressevaelgerEndpoint = ts.URL

	suggestions, err := client.Autocomplete(context.Background(), "Amagertorv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions, got %d", len(suggestions))
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
	// When token is empty, client should fall back to Adressevælgeren and succeed
	client := NewClient("")
	res, err := client.LookupAddress(context.Background(), "Højgårdsvej 9, 9670 Løgstør")
	if err != nil {
		t.Fatalf("unexpected error with fallback: %v", err)
	}
	if res.KommuneCode != "0820" {
		t.Errorf("got KommuneCode %s, want 0820", res.KommuneCode)
	}
	if !res.IsYderkommune {
		t.Errorf("got IsYderkommune false, want true")
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

func TestLookupAddress_DatafordelerenEndpoint_NoAuthHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// When querying Datafordeleren, apiKey is in query params and Authorization header MUST NOT be present
		// to avoid DAF-AUTH-0002 multiple authentication methods error.
		if r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected Authorization header present: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("apiKey") != "my-df-token" {
			t.Errorf("expected apiKey my-df-token, got %s", r.URL.Query().Get("apiKey"))
		}

		response := `{
			"data": {
				"DAR_Husnummer": {
					"nodes": [
						{
							"adgangsadressebetegnelse": "Højgårdsvej 9, Mjallerup, 9670 Løgstør",
							"vejmidte": "0820-0587",
							"status": "3"
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

	client := NewClient("my-df-token")
	client.Endpoint = ts.URL + "/graphql.datafordeler.dk/DAR/v3"

	res, err := client.LookupAddress(context.Background(), "Højgårdsvej 9, Mjallerup, 9670 Løgstør")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.KommuneCode != "0820" {
		t.Errorf("got KommuneCode %s, want 0820", res.KommuneCode)
	}
	if !res.IsYderkommune {
		t.Errorf("expected isYderkommune=true")
	}
}




