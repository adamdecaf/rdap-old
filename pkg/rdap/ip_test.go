package rdap

import (
	"testing"
)

func TestClient__IPFail(t *testing.T) {
	failures := []string{
		"",
		"    ",
		"999.999.999.999",
		"123/40",
		"198.51.100.1/ZZZZff00",
	}

	client := Client{}
	for i := range failures {
		_, err := client.IP(failures[i])
		if err == nil {
			t.Errorf("expected failure with %q", failures[i])
		}
	}
}
