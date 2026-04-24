package glicko2

import (
	"math"
	"testing"
)

// tolerance for floating-point comparisons.
const tol = 0.5

func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// TestCalculate_GlickmanExample validates against the worked example from Glickman's 2012 paper.
// Player: r=1500, RD=200, σ=0.06
// Opponents: (1400,30), (1550,100), (1700,300)
// Scores:     win,        loss,         loss
// Expected:   r'≈1464.1, RD'≈151.5, σ'≈0.06 (approx)
func TestCalculate_GlickmanExample(t *testing.T) {
	player := Player{Rating: 1500, Deviation: 200, Volatility: 0.06}
	results := []MatchResult{
		{Opponent: Player{Rating: 1400, Deviation: 30, Volatility: 0.06}, Score: 1.0},
		{Opponent: Player{Rating: 1550, Deviation: 100, Volatility: 0.06}, Score: 0.0},
		{Opponent: Player{Rating: 1700, Deviation: 300, Volatility: 0.06}, Score: 0.0},
	}

	got := Calculate(player, results)

	// Glickman paper gives r'≈1464.1, RD'≈151.5
	if !approxEqual(got.Rating, 1464.1, 2.0) {
		t.Errorf("rating: expected ≈1464.1, got %.2f", got.Rating)
	}
	if !approxEqual(got.Deviation, 151.5, 2.0) {
		t.Errorf("deviation: expected ≈151.5, got %.2f", got.Deviation)
	}
	if got.Volatility <= 0 || got.Volatility > 0.1 {
		t.Errorf("volatility out of range: got %f", got.Volatility)
	}
}

// TestCalculate_NoResults ensures deviation increases when player doesn't compete.
func TestCalculate_NoResults(t *testing.T) {
	player := Player{Rating: 1500, Deviation: 200, Volatility: 0.06}
	got := Calculate(player, nil)

	if got.Rating != player.Rating {
		t.Errorf("rating should be unchanged: got %.2f", got.Rating)
	}
	if got.Deviation <= player.Deviation {
		t.Errorf("deviation should increase with no play: got %.2f", got.Deviation)
	}
}

// TestCalculate_AllWins ensures rating increases after all wins.
func TestCalculate_AllWins(t *testing.T) {
	player := Player{Rating: 1500, Deviation: 200, Volatility: 0.06}
	results := []MatchResult{
		{Opponent: Player{Rating: 1500, Deviation: 200, Volatility: 0.06}, Score: 1.0},
		{Opponent: Player{Rating: 1500, Deviation: 200, Volatility: 0.06}, Score: 1.0},
		{Opponent: Player{Rating: 1500, Deviation: 200, Volatility: 0.06}, Score: 1.0},
	}
	got := Calculate(player, results)
	if got.Rating <= player.Rating {
		t.Errorf("rating should increase after all wins: got %.2f", got.Rating)
	}
}

// TestCalculate_AllLosses ensures rating decreases after all losses.
func TestCalculate_AllLosses(t *testing.T) {
	player := Player{Rating: 1500, Deviation: 200, Volatility: 0.06}
	results := []MatchResult{
		{Opponent: Player{Rating: 1500, Deviation: 200, Volatility: 0.06}, Score: 0.0},
		{Opponent: Player{Rating: 1500, Deviation: 200, Volatility: 0.06}, Score: 0.0},
	}
	got := Calculate(player, results)
	if got.Rating >= player.Rating {
		t.Errorf("rating should decrease after all losses: got %.2f", got.Rating)
	}
}
