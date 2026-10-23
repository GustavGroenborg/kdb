package dar

import (
	"context"
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
