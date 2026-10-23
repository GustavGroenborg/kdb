package kommune

import "testing"

func TestLookupKommune(t *testing.T) {
	tests := []struct {
		code          string
		expectedName  string
		expectedYder  bool
		expectedFound bool
	}{
		{"0101", "København", false, true},
		{"0376", "Guldborgsund", true, true},
		{"376", "Guldborgsund", true, true},
		{"0860", "Hjørring", true, true},
		{"0400", "Bornholm", true, true},
		{"0751", "Aarhus", false, true},
		{"9999", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			name, isYder, found := LookupKommune(tt.code)
			if found != tt.expectedFound {
				t.Fatalf("LookupKommune(%s) found=%v; want %v", tt.code, found, tt.expectedFound)
			}
			if name != tt.expectedName {
				t.Errorf("LookupKommune(%s) name=%s; want %s", tt.code, name, tt.expectedName)
			}
			if isYder != tt.expectedYder {
				t.Errorf("LookupKommune(%s) isYder=%v; want %v", tt.code, isYder, tt.expectedYder)
			}
		})
	}
}
