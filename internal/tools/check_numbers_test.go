package tools

import (
	"testing"
)

func TestCalcRank(t *testing.T) {
	tests := []struct {
		name         string
		matchCount   int
		bonusMatched bool
		wantRank     int
	}{
		{"1등: 6개 일치", 6, false, 1},
		{"1등: 6개 일치 (보너스도)", 6, true, 1},
		{"2등: 5개 + 보너스", 5, true, 2},
		{"3등: 5개 일치", 5, false, 3},
		{"4등: 4개 일치", 4, false, 4},
		{"5등: 3개 일치", 3, false, 5},
		{"미당첨: 2개 일치", 2, false, 0},
		{"미당첨: 1개 일치", 1, false, 0},
		{"미당첨: 0개 일치", 0, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcRank(tt.matchCount, tt.bonusMatched)
			if got != tt.wantRank {
				t.Errorf("calcRank(%d, %v) = %d, want %d", tt.matchCount, tt.bonusMatched, got, tt.wantRank)
			}
		})
	}
}
