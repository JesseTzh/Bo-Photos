package asset

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		{StatusProcessing, StatusReady, true},
		{StatusProcessing, StatusFailed, true},
		{StatusProcessing, StatusDeleted, true},
		{StatusProcessing, StatusPurged, false},
		{StatusReady, StatusProcessing, true},
		{StatusReady, StatusDeleted, true},
		{StatusReady, StatusFailed, false},
		{StatusReady, StatusPurged, false},
		{StatusFailed, StatusProcessing, true},
		{StatusFailed, StatusDeleted, true},
		{StatusFailed, StatusReady, false},
		{StatusFailed, StatusPurged, false},
		{StatusDeleted, StatusReady, true},
		{StatusDeleted, StatusPurged, true},
		{StatusDeleted, StatusFailed, false},
		{StatusPurged, StatusDeleted, false},
		{StatusPurged, StatusReady, false},
	}
	for _, test := range tests {
		got := CanTransition(test.from, test.to)
		if got != test.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", test.from, test.to, got, test.want)
		}
	}
}
