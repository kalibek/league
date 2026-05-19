# How to Create a League

This guide walks you through setting up a new table tennis league from scratch.

## Step 1: Create the League

1. Click **"+ New League"** from the Leagues page
2. Enter a **league name** (e.g., "Tuesday Night Table Tennis")
3. Optionally add a **description** (e.g., "Casual competitive league for club members")
4. Click **Create League**

Your league is now created with default configuration settings.

## Step 2: Configure League Rules

1. Open your new league
2. Click the **Configure** button
3. Set the following rules:

### Group Size
- Determines how many players are in each group
- Common values: 5–7 players per group
- More players = longer events, more round-robin matches

### Advances per Group
- How many top finishers advance to the next higher group
- Common value: 2 players per group
- Set to 0 if you don't want advancement

### Recedes per Group
- How many bottom finishers drop to the next lower group
- Common value: 2 players per group
- Set to 0 if you don't want relegation

### Games to Win
- Best-of-how-many for each match
- **3** = first to 2 games wins
- **5** = first to 3 games wins

### Number of Tables
- How many ping-pong tables you have available
- Used by the scheduler when assigning matches to time slots

4. Click **Save Configuration**

## Step 3: Add League Members

Members are automatically added when they join the league or when an admin adds them during event setup.

## Step 4: Create Your First Event

1. Click **+ Create Event** on the league page
2. Enter an **event title** (e.g., "January Monthly")
3. Click **Create**

Your event is now in **DRAFT** status.

## Step 5: Set Up Groups and Players

1. Click **Setup** on the event
2. Click **+ Add Group** to create a group
3. For each group:
   - Set the **division** and **group number** (e.g., "Division 1, Group 1")
   - Optionally set a **scheduled date/time**
   - Click **Create Group**

4. Add players to groups:
   - Click the group to open its player list
   - Select a player from the dropdown
   - Click **Add**

5. Ensure all groups have the correct number of players and all players are assigned

## Step 6: Start the Event

1. Once all groups are set up, click **Start** on the league page
2. The event status changes to **IN_PROGRESS**
3. The app automatically generates a **round-robin schedule** for each group

## Step 7: Score Matches

1. Click **View** on the event to open the live scoring interface
2. For each match:
   - Click the score area
   - Enter each game's result (11-5, 11-3, etc.)
   - Submit the score

3. Scores update in **real-time** — all viewers see updates instantly via WebSocket

## Step 8: Finish the Event

1. When all groups have completed all their matches:
   - Click **Finish Group** for each completed group
   - Or click **Finish Event** to close the entire event at once

2. The app:
   - Calculates **final standings** in each group
   - Resolves **tiebreaks** (automatic + manual if needed)
   - Updates each player's **Glicko2 rating**
   - Schedules **promotions and relegations** for the next event

3. The draft for the **next event** is automatically created

## Tips for Success

- **Plan group counts:** If you have 24 players and group size is 6, you'll have 4 groups
- **Set realistic match times:** Use the scheduled date/time to indicate when each group will play
- **Verify standings:** Always check tiebreaks and manual placements are correct before finishing
- **Backup scoring:** Consider taking photos of score sheets in case of data issues
- **Communicate:** Let players know when results will be published and ratings updated

## Common Issues

**"Tiebreaker required"** — Two or more players finished level. Drag them into the correct order and confirm.

**"Not enough players in group"** — Add more players or remove the group entirely.

**"Event is locked"** — The event is no longer in DRAFT status. You cannot change groups or players anymore.

---

Questions? Contact your league admin or check the other help pages!
