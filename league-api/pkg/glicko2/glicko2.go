// Package glicko2 implements the Glicko2 rating algorithm as described in
// Mark Glickman's 2012 paper "Example of the Glicko-2 system".
package glicko2

import "math"

const (
	// tau is the system constant; constrains volatility change.
	tau = 0.5
	// epsilon is the convergence tolerance for the Illinois algorithm.
	epsilon = 0.000001
	// scale is the conversion factor between Glicko1 and Glicko2 scales.
	scale = 173.7178
)

// Player holds the current Glicko2 rating parameters.
type Player struct {
	Rating     float64 // r  (Glicko1 scale, default 1500)
	Deviation  float64 // RD (Glicko1 scale, default 350)
	Volatility float64 // σ  (default 0.06)
}

// MatchResult represents one game in the rating period.
type MatchResult struct {
	Opponent Player
	Score    float64 // 1.0 = win, 0.5 = draw, 0.0 = loss
}

// Calculate returns new rating parameters for one rating period.
// If results is empty the function updates only the deviation (period without play).
func Calculate(player Player, results []MatchResult) Player {
	// Step 1: Convert to Glicko2 scale.
	mu := (player.Rating - 1500) / scale
	phi := player.Deviation / scale
	sigma := player.Volatility

	if len(results) == 0 {
		// Player did not compete; only deviation increases.
		phiStar := math.Sqrt(phi*phi + sigma*sigma)
		return Player{
			Rating:     scale*mu + 1500,
			Deviation:  scale * phiStar,
			Volatility: sigma,
		}
	}

	// Step 2: Compute auxiliary values for each game.
	type gameAux struct {
		gj float64 // g(φj)
		Ej float64 // E(μ, μj, φj)
		sj float64 // score
	}
	games := make([]gameAux, len(results))
	for i, res := range results {
		muj := (res.Opponent.Rating - 1500) / scale
		phij := res.Opponent.Deviation / scale
		gj := gFunc(phij)
		Ej := eFunc(mu, muj, phij)
		games[i] = gameAux{gj: gj, Ej: Ej, sj: res.Score}
	}

	// Step 3: Compute v (variance inverse).
	var vInv float64
	for _, g := range games {
		vInv += g.gj * g.gj * g.Ej * (1 - g.Ej)
	}
	v := 1.0 / vInv

	// Step 4: Compute Δ (improvement estimate).
	var deltaSum float64
	for _, g := range games {
		deltaSum += g.gj * (g.sj - g.Ej)
	}
	delta := v * deltaSum

	// Step 5: Determine new volatility σ' via Illinois algorithm (regula falsi).
	sigmaPrime := newVolatility(phi, sigma, delta, v)

	// Step 6: Update φ*.
	phiStar := math.Sqrt(phi*phi + sigmaPrime*sigmaPrime)

	// Step 7: Update φ'.
	phiPrime := 1 / math.Sqrt(1/phiStar/phiStar+1/v)

	// Step 8: Update μ'.
	muPrime := mu + phiPrime*phiPrime*deltaSum

	// Step 9: Convert back to Glicko1 scale.
	return Player{
		Rating:     scale*muPrime + 1500,
		Deviation:  scale * phiPrime,
		Volatility: sigmaPrime,
	}
}

// gFunc computes g(φ) = 1 / sqrt(1 + 3φ²/π²).
func gFunc(phi float64) float64 {
	return 1 / math.Sqrt(1+3*phi*phi/(math.Pi*math.Pi))
}

// eFunc computes E(μ, μj, φj) = 1 / (1 + exp(-g(φj)(μ - μj))).
func eFunc(mu, muj, phij float64) float64 {
	return 1 / (1 + math.Exp(-gFunc(phij)*(mu-muj)))
}

// newVolatility finds σ' using the Illinois algorithm from Glickman's paper.
func newVolatility(phi, sigma, delta, v float64) float64 {
	a := math.Log(sigma * sigma)
	phi2 := phi * phi

	f := func(x float64) float64 {
		ex := math.Exp(x)
		d := phi2 + v + ex
		num := ex * (delta*delta - phi2 - v - ex)
		den := 2 * d * d
		return num/den - (x-a)/(tau*tau)
	}

	// Initial bracket.
	A := a
	var B float64
	if delta*delta > phi2+v {
		B = math.Log(delta*delta - phi2 - v)
	} else {
		k := 1.0
		for f(a-k*tau) < 0 {
			k++
		}
		B = a - k*tau
	}

	fA := f(A)
	fB := f(B)

	for math.Abs(B-A) > epsilon {
		C := A + (A-B)*fA/(fB-fA)
		fC := f(C)
		if fC*fB <= 0 {
			A = B
			fA = fB
		} else {
			fA /= 2
		}
		B = C
		fB = fC
	}

	return math.Exp(A / 2)
}
