package http

import (
	"math"
	"testing"
)

func TestCalculateExpectedScore(t *testing.T) {
	tests := []struct {
		name           string
		playerRating   float64
		opponentRating float64
		expectedScore  float64
		tolerance      float64
	}{
		{
			name:           "Equal ratings",
			playerRating:   1500,
			opponentRating: 1500,
			expectedScore:  0.5,
			tolerance:      0.001,
		},
		{
			name:           "Player rated 400 points higher",
			playerRating:   1900,
			opponentRating: 1500,
			expectedScore:  0.909, // ~91% win probability
			tolerance:      0.001,
		},
		{
			name:           "Player rated 400 points lower",
			playerRating:   1500,
			opponentRating: 1900,
			expectedScore:  0.091, // ~9% win probability
			tolerance:      0.001,
		},
		{
			name:           "Player rated 200 points higher",
			playerRating:   1700,
			opponentRating: 1500,
			expectedScore:  0.76, // ~76% win probability
			tolerance:      0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateExpectedScore(tt.playerRating, tt.opponentRating)
			if math.Abs(result-tt.expectedScore) > tt.tolerance {
				t.Errorf("CalculateExpectedScore() = %v, want %v (±%v)",
					result, tt.expectedScore, tt.tolerance)
			}
		})
	}
}

func TestCalculateNewRating(t *testing.T) {
	tests := []struct {
		name           string
		currentRating  float64
		expectedScore  float64
		actualScore    float64
		kFactor        float64
		expectedRating float64
	}{
		{
			name:           "Expected win becomes actual win",
			currentRating:  1500,
			expectedScore:  0.76,
			actualScore:    1.0,
			kFactor:        32,
			expectedRating: 1500 + 32*(1.0-0.76), // +7.68
		},
		{
			name:           "Unexpected loss",
			currentRating:  1500,
			expectedScore:  0.76,
			actualScore:    0.0,
			kFactor:        32,
			expectedRating: 1500 + 32*(0.0-0.76), // -24.32
		},
		{
			name:           "50-50 match ends in win",
			currentRating:  1500,
			expectedScore:  0.5,
			actualScore:    1.0,
			kFactor:        32,
			expectedRating: 1500 + 32*(1.0-0.5), // +16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNewRating(tt.currentRating, tt.expectedScore, tt.actualScore, tt.kFactor)
			if math.Abs(result-tt.expectedRating) > 0.01 {
				t.Errorf("CalculateNewRating() = %v, want %v",
					result, tt.expectedRating)
			}
		})
	}
}

func TestUpdateRatings(t *testing.T) {
	tests := []struct {
		name           string
		rating1        float64
		rating2        float64
		player1Won     bool
		kFactor        float64
		expectedGain   float64 // Expected rating change for winner
		expectedLoss   float64 // Expected rating change for loser (negative)
		tolerance      float64
	}{
		{
			name:         "Equal players, player 1 wins",
			rating1:      1500,
			rating2:      1500,
			player1Won:   true,
			kFactor:      32,
			expectedGain: 16,   // Winner gains 16
			expectedLoss: -16,  // Loser loses 16
			tolerance:    0.01,
		},
		{
			name:         "Equal players, player 2 wins",
			rating1:      1500,
			rating2:      1500,
			player1Won:   false,
			kFactor:      32,
			expectedGain: 16,   // Player 2 gains 16
			expectedLoss: -16,  // Player 1 loses 16
			tolerance:    0.01,
		},
		{
			name:         "Underdog wins (player 1 rated 200 lower)",
			rating1:      1300,
			rating2:      1500,
			player1Won:   true,
			kFactor:      42,
			expectedGain: 31.5, // Underdog gains ~31.5
			expectedLoss: -31.5, // Favorite loses ~31.5
			tolerance:    0.5,
		},
		{
			name:         "Favorite wins (player 1 rated 200 higher)",
			rating1:      1700,
			rating2:      1500,
			player1Won:   true,
			kFactor:      42,
			expectedGain: 10.5, // Favorite gains ~10.5
			expectedLoss: -10.5, // Underdog loses ~10.5
			tolerance:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRating1, newRating2 := UpdateRatings(tt.rating1, tt.rating2, tt.player1Won, tt.kFactor)

			var winnerGain, loserLoss float64
			if tt.player1Won {
				winnerGain = newRating1 - tt.rating1
				loserLoss = newRating2 - tt.rating2
			} else {
				winnerGain = newRating2 - tt.rating2
				loserLoss = newRating1 - tt.rating1
			}

			// Check winner's gain
			if math.Abs(winnerGain-tt.expectedGain) > tt.tolerance {
				t.Errorf("Winner's rating change = %v, want %v (±%v)",
					winnerGain, tt.expectedGain, tt.tolerance)
			}

			// Check loser's loss
			if math.Abs(loserLoss-tt.expectedLoss) > tt.tolerance {
				t.Errorf("Loser's rating change = %v, want %v (±%v)",
					loserLoss, tt.expectedLoss, tt.tolerance)
			}
		})
	}
}

// TestRatingSymmetry verifies that rating changes are symmetric
// (what one player gains, the other loses)
func TestRatingSymmetry(t *testing.T) {
	tests := []struct {
		name       string
		rating1    float64
		rating2    float64
		player1Won bool
		kFactor    float64
	}{
		{"Equal ratings, player 1 wins", 1500, 1500, true, 32},
		{"Equal ratings, player 2 wins", 1500, 1500, false, 32},
		{"Different ratings, player 1 wins", 1700, 1300, true, 42},
		{"Different ratings, player 2 wins", 1700, 1300, false, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRating1, newRating2 := UpdateRatings(tt.rating1, tt.rating2, tt.player1Won, tt.kFactor)

			change1 := newRating1 - tt.rating1
			change2 := newRating2 - tt.rating2
			totalChange := change1 + change2

			// Total rating change should be close to zero (conservation of rating points)
			if math.Abs(totalChange) > 0.001 {
				t.Errorf("Rating changes not symmetric: player1=%v, player2=%v, total=%v",
					change1, change2, totalChange)
			}
		})
	}
}

// TestExpectedScoreSymmetry verifies that expected scores sum to 1.0
func TestExpectedScoreSymmetry(t *testing.T) {
	tests := []struct {
		rating1 float64
		rating2 float64
	}{
		{1500, 1500},
		{1700, 1300},
		{2000, 1200},
		{1450, 1550},
	}

	for _, tt := range tests {
		exp1 := CalculateExpectedScore(tt.rating1, tt.rating2)
		exp2 := CalculateExpectedScore(tt.rating2, tt.rating1)
		sum := exp1 + exp2

		if math.Abs(sum-1.0) > 0.001 {
			t.Errorf("Expected scores don't sum to 1.0: %v + %v = %v (ratings: %v vs %v)",
				exp1, exp2, sum, tt.rating1, tt.rating2)
		}
	}
}
