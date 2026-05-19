# Rating System

This app uses the **Glicko2** algorithm to rate players. Glicko2 improves on the classic Elo system by explicitly tracking how confident the system is in each rating.

## The Three Parameters

Every player has three values:

| Parameter | Symbol | Starting Value | Meaning |
|---|---|---|---|
| **Rating** | μ (mu) | 1500 | Estimated skill level |
| **Rating Deviation** | RD (σ) | 350 | Uncertainty in the rating |
| **Volatility** | φ (phi) | 0.06 | How erratic the player's performance is |

- **Rating** is the number displayed on your profile. A higher number means a stronger player.
- **Rating Deviation** is a confidence interval. A rating of 1600 ±50 RD is much more reliable than 1600 ±300 RD. New players start with RD 350 because nothing is known about them yet.
- **Volatility** captures inconsistency. A player who sometimes crushes strong opponents and sometimes loses to weaker ones will have a higher volatility than someone who performs predictably.

## How Ratings Update

After every group event, ratings are recalculated using all the results within the group. The size of the change depends on:

1. **Who you beat or lost to.** Beating a highly-rated opponent earns more rating points than beating a weaker one. Losing to a strong opponent costs fewer points than losing to a weaker one.
2. **Your current RD.** A high RD means the system is uncertain about you, so it adjusts your rating more aggressively in both directions until it has more data.
3. **Your opponent's RD.** Results against opponents with high RD carry less weight because those ratings are themselves uncertain.

## The Expected Score Formula

The probability that player A (rating r) beats player B (rating r_j, deviation RD_j) is:

```
E = 1 / (1 + exp(-g(RD_j) × (r − r_j) / 400))
```

where `g(RD_j)` is a scaling factor that reduces the impact of opponents with high RD. In plain terms: **the further apart two players are in rating, the more predictable the outcome, and the less rating changes hands**.

## Practical Examples

- You are rated **1400** and beat a player rated **1700**. This is a big upset — your rating rises significantly, perhaps by 60–80 points.
- You are rated **1700** and beat a player rated **1400**. This was expected — your rating barely moves, perhaps +5 points.
- You are a **new player** (RD 350). After your first event your rating might swing ±150 points because the system is still learning where you belong.
- After playing **10 events**, your RD might drop to around 60–80. Your rating changes by smaller amounts each event, reflecting that your skill level is now well-established.

## Inactive Players

If a player misses events, their **RD slowly increases** over time. This means the system becomes less certain about their rating again. When they return, their first event back can produce a larger rating change — both up and down — until certainty is restored.
