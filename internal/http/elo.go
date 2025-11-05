package http

import "math"

// CalculateExpectedScore calculates the expected score for a player
// using the standard ELO formula:
// Expected score = 1 / (1 + 10^((opponent_rating - player_rating) / 400))
func CalculateExpectedScore(playerRating, opponentRating float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (opponentRating-playerRating)/400))
}

// CalculateNewRating calculates the new rating after a match
// Parameters:
// - currentRating: Player's rating before the match
// - expectedScore: Expected probability of winning (0.0 to 1.0)
// - actualScore: Actual result (1.0 for win, 0.0 for loss)
// - kFactor: K-factor controlling rating volatility
func CalculateNewRating(currentRating, expectedScore, actualScore, kFactor float64) float64 {
	return currentRating + kFactor*(actualScore-expectedScore)
}

// UpdateRatings calculates new ratings for both players after a match
// Returns (newRating1, newRating2)
func UpdateRatings(rating1, rating2 float64, player1Won bool, kFactor float64) (float64, float64) {
	exp1 := CalculateExpectedScore(rating1, rating2)
	exp2 := CalculateExpectedScore(rating2, rating1)

	var actualScore1, actualScore2 float64
	if player1Won {
		actualScore1, actualScore2 = 1.0, 0.0
	} else {
		actualScore1, actualScore2 = 0.0, 1.0
	}

	newRating1 := CalculateNewRating(rating1, exp1, actualScore1, kFactor)
	newRating2 := CalculateNewRating(rating2, exp2, actualScore2, kFactor)

	return newRating1, newRating2
}
