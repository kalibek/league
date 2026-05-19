# League Movements

Players move between groups based on their finishing position in each event. This creates a **self-sorting competitive ladder** where players settle into groups that match their actual skill level over time.

## Group Structure

Groups are arranged in a hierarchy: **S → A1 → A2 → B1 → B2 → C1 → C2 → …**

S is the top (elite) group. Each division letter represents a tier. Within a tier, groups are numbered — A1 and A2 are parallel groups at the A level. Promotion and relegation always moves one step up or down this chain.

## How Movements Work

Each league has two configured numbers:

- **Advances per group** — how many top finishers move up
- **Recedes per group** — the target number of bottom finishers who move down

After a group finishes, players are ranked and movements are applied:

- **Top N finishers** → advance to the group above
- **Bottom M finishers** → recede to the group below (see DNS rules below)
- **Everyone else** → stays in the same group next event

## DNS Players and Recede Slots

A player who **did not show (DNS)** always recedes regardless of their place. DNS players count toward the configured number of recedes.

**Delta rule:**
```
delta = NumberOfRecedes − count(DNS players)
```
- If **delta > 0**: the bottom `delta` active (non-DNS) players also recede
- If **delta ≤ 0**: no active player recedes — the DNS players already fill all recede slots

**Example:** configured recedes = 3, group has 2 DNS players → delta = 1, so only the last-place active player also recedes. Total receding = 3.

If there are 4 DNS players and configured recedes = 3: all 4 DNS recede (all DNS always recede), and no active player recedes.

## Tiebreak Resolution

When players finish with equal points, placement is resolved through four steps in order:

1. **Tiebreak points** — game score differential (games won minus games lost) calculated *only* against the other tied players
2. **W/L ratio** — total wins divided by total losses across all group matches. A player with 4W/1L (ratio 4.0) ranks above 2W/2L (ratio 1.0). Division by zero: a player with wins and no losses ranks highest
3. **Head-to-head** — used for two-player sub-ties after step 1: the winner of the direct match takes the higher place
4. **Manual placement** — if still unresolvable (e.g. three players with equal W/L), an admin uses the manual placement tool to drag-and-drop the final order

## Concrete Example

Group B1 has 6 players. Configured: advances = 1, recedes = 2. One player was DNS.

- DNS player → recedes (delta = 2 − 1 = 1, so 1 active player also recedes)
- Players 1st and 2nd finish tied on points → tiebreak points checked → still tied → W/L ratio breaks the tie
- 1st place → advances to A2
- 2nd–4th place → stay in B1
- 5th place (active) → recedes to C1
- DNS player → recedes to C1

## Rating vs. Movements

Group movements are determined by **match points** within the event. **Glicko2 ratings** are calculated separately from match scores and are not directly used for placement. However, advancing to a higher group means tougher opponents, which affects how your rating changes in future events.
