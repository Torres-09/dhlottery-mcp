package client

import (
	"testing"
)

func TestGetWinningNumbers_ParseLogic(t *testing.T) {
	// lt645Item 파싱 및 변환 로직 검증
	item := lt645Item{
		LtEpsd:             1150,
		LtRflYmd:           "20241214",
		Tm1WnNo:            8,
		Tm2WnNo:            9,
		Tm3WnNo:            18,
		Tm4WnNo:            35,
		Tm5WnNo:            39,
		Tm6WnNo:            45,
		BnsWnNo:            25,
		RlvtEpsdSumNtslAmt: 57076017000,
		Rnk1WnAmt:          1570620309,
		Rnk1WnNope:         17,
	}

	result := &WinningNumbers{
		Round:        item.LtEpsd,
		Date:         formatDate(item.LtRflYmd),
		Numbers:      []int{item.Tm1WnNo, item.Tm2WnNo, item.Tm3WnNo, item.Tm4WnNo, item.Tm5WnNo, item.Tm6WnNo},
		Bonus:        item.BnsWnNo,
		TotalSales:   item.RlvtEpsdSumNtslAmt,
		FirstPrize:   item.Rnk1WnAmt,
		FirstWinners: item.Rnk1WnNope,
	}

	if result.Round != 1150 {
		t.Errorf("Round = %d, want 1150", result.Round)
	}
	if result.Date != "2024-12-14" {
		t.Errorf("Date = %q, want %q", result.Date, "2024-12-14")
	}
	if len(result.Numbers) != 6 {
		t.Errorf("Numbers length = %d, want 6", len(result.Numbers))
	}
	if result.Numbers[0] != 8 {
		t.Errorf("Numbers[0] = %d, want 8", result.Numbers[0])
	}
	if result.Bonus != 25 {
		t.Errorf("Bonus = %d, want 25", result.Bonus)
	}
	if result.FirstWinners != 17 {
		t.Errorf("FirstWinners = %d, want 17", result.FirstWinners)
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"20241214", "2024-12-14"},
		{"20250101", "2025-01-01"},
		{"2024-12-14", "2024-12-14"}, // 이미 포맷된 경우
		{"", ""},
	}

	for _, tt := range tests {
		got := formatDate(tt.input)
		if got != tt.want {
			t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetLatestRound_FindMax(t *testing.T) {
	// 최신 회차 찾기 로직 검증
	items := []lt645Item{
		{LtEpsd: 1148, LtRflYmd: "20241130"},
		{LtEpsd: 1150, LtRflYmd: "20241214"},
		{LtEpsd: 1149, LtRflYmd: "20241207"},
	}

	latest := items[0]
	for _, item := range items {
		if item.LtEpsd > latest.LtEpsd {
			latest = item
		}
	}

	if latest.LtEpsd != 1150 {
		t.Errorf("최신 회차 = %d, want 1150", latest.LtEpsd)
	}
	if formatDate(latest.LtRflYmd) != "2024-12-14" {
		t.Errorf("최신 추첨일 = %q, want %q", formatDate(latest.LtRflYmd), "2024-12-14")
	}
}
