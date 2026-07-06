package domain

import "testing"

func TestTopClassIdFromVotes(t *testing.T) {
	tests := []struct {
		name      string
		votes     map[string]int
		wantId    string
		wantFound bool
	}{
		{
			name:      "empty map has no favorite",
			votes:     map[string]int{},
			wantFound: false,
		},
		{
			name:      "single entry is always a clear favorite",
			votes:     map[string]int{"3": 5},
			wantId:    "3",
			wantFound: true,
		},
		{
			name:      "one clear winner among several",
			votes:     map[string]int{"1": 4, "2": 30, "3": 12, "4": 1},
			wantId:    "2",
			wantFound: true,
		},
		{
			name:      "exact tie between two is not a clear favorite",
			votes:     map[string]int{"1": 10, "2": 10},
			wantFound: false,
		},
		{
			name:      "exact tie among three is not a clear favorite",
			votes:     map[string]int{"1": 7, "2": 7, "3": 7, "4": 1},
			wantFound: false,
		},
		{
			name:      "a later strictly-greater value resets the tie count",
			votes:     map[string]int{"1": 7, "3": 7, "2": 9},
			wantId:    "2",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotFound := TopClassIdFromVotes(tt.votes)
			if gotFound != tt.wantFound {
				t.Fatalf("TopClassIdFromVotes(%v) found = %v, want %v", tt.votes, gotFound, tt.wantFound)
			}
			if gotFound && gotId != tt.wantId {
				t.Errorf("TopClassIdFromVotes(%v) id = %q, want %q", tt.votes, gotId, tt.wantId)
			}
		})
	}
}
