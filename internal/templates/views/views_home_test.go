package views

import "testing"

func TestFormatTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		want    int
	}{
		{name: "Disabled when zero", timeout: 0, want: 0},
		{name: "Disabled when negative", timeout: -1, want: 0},
		{name: "Default http_timeout", timeout: 20, want: 80000},
		{name: "One second", timeout: 1, want: 4000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeout(tt.timeout)
			if got != tt.want {
				t.Errorf("formatTimeout(%d) = %d, want %d", tt.timeout, got, tt.want)
			}
		})
	}
}
