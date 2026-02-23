package timestamp

import (
	"strconv"
	"strings"
	"testing"
)

func TestResolveRR_BareRR(t *testing.T) {
	for i := 0; i < 50; i++ {
		result, err := ResolveRR("RR:RR:RR")
		if err != nil {
			t.Fatalf("ResolveRR(bare): %v", err)
		}

		parts := strings.Split(result, ":")
		if len(parts) != 3 {
			t.Fatalf("expected HH:MM:SS, got %q", result)
		}

		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		s, _ := strconv.Atoi(parts[2])

		if h < 0 || h > 23 {
			t.Errorf("hour out of range: %d", h)
		}
		if m < 0 || m > 59 {
			t.Errorf("minute out of range: %d", m)
		}
		if s < 0 || s > 59 {
			t.Errorf("second out of range: %d", s)
		}
	}
}

func TestResolveRR_WithBounds(t *testing.T) {
	for i := 0; i < 50; i++ {
		result, err := ResolveRR("RR(09,17):RR(00,30):00")
		if err != nil {
			t.Fatalf("ResolveRR(bounded): %v", err)
		}

		parts := strings.Split(result, ":")
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])

		if h < 9 || h > 17 {
			t.Errorf("hour %d not in [9,17]", h)
		}
		if m < 0 || m > 30 {
			t.Errorf("minute %d not in [0,30]", m)
		}
		if parts[2] != "00" {
			t.Errorf("seconds should be 00, got %s", parts[2])
		}
	}
}

func TestResolveRR_NoRR(t *testing.T) {
	result, err := ResolveRR("14:30:00")
	if err != nil {
		t.Fatalf("ResolveRR(no-rr): %v", err)
	}
	if result != "14:30:00" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

func TestResolveRR_InvalidFormat(t *testing.T) {
	_, err := ResolveRR("14:30")
	if err == nil {
		t.Error("expected error for HH:MM (missing seconds)")
	}
}

func TestResolveRR_MinGreaterThanMax(t *testing.T) {
	_, err := ResolveRR("RR(20,10):00:00")
	if err == nil {
		t.Error("expected error when min > max")
	}
}

func TestContainsRR(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"RR", true},
		{"RR(08,17)", true},
		{"14:30:00", false},
		{"PREV", false},
		{"RR:RR:00", true},
	}
	for _, tt := range tests {
		if got := ContainsRR(tt.input); got != tt.want {
			t.Errorf("ContainsRR(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
