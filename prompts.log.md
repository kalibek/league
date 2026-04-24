### Group view improvements:
1. Sort group by initial seeding
2. Show names, not Player #{id} in group list and matrix
3. Put `did not show` button inline with player name, make icon button

### Score entering form improvements:
1. Show dynamic buttons for possible scores instead of number inputs. E.g. best of 3 → buttons 2-0, 2-1, 0-2, 1-2


# User stories

## Admin role
### Description:
Admin create leagues, assign Maintainers.
### Acceptance criteria:
- Only Admin create leagues
- Admin and Maintainers assign Maintainers to leagues

## Next league event
### Description:
Maintainer create new league event when current one over.
### Acceptance criteria:
- New event created with same settings as previous.
- Same number of players as previous.
- Players moved per advances/recedes from previous event.

## Player profile view
### Description:
Any user view any player profile: league participation, rating delta, match scores ordered newest to oldest.
### Acceptance criteria:
- Load 5 most recent events. Older on scroll.

## Player profile edit
### Description:
Player edit own profile. If profile unfilled, show in-app notification until filled.
### Acceptance criteria:
- Profile editable.
- Editable fields:
  - First name
  - Last name
  - Country - dropdown, all countries from list. Countries in DB migration
  - City - dropdown with ability to add new city
  - Birthdate
  - Grip (penhold, shakehand)
  - Gender
  - Setup:
    - Blade - dropdown, add new blade
    - Forehand Rubber - dropdown, add new rubber (same table as backhand)
    - Backhand Rubber - dropdown, add new rubber (same table as forehand)

## Player profile search
### Description:
Any user search players by any profile field.
### Acceptance criteria:
- Search case insensitive

## Internationalization
### Description:
Any user change app language.
### Acceptance criteria:
- Library: `react-i18next`
- All static texts translated in frontend
- 3 languages:
    - English
    - Russian
    - Kazakh
  
## Leagues, events pagination
### Description:
Show only recent events in leagues/events list. Same for player lists and searches.
### Acceptance criteria:
- Older events on scroll.
- 10 events per page, configurable


## Tie break calculation
### Description:
Tie break points should be only calculated within the players who have the equal number of points. If no tiebreak points are calculated,
on the frontend TB field should be '-'
### Examples:
- In a group of 4 players A, B, C, D have 6, 5, 4, 3 points, then no tie break calculation is required
- In a group of 4 players A, B, C, D have 5, 5, 5, 3 points, then tie break calculated only for A, B, C players
- In a group of 4 players A, B, C, D have 5, 5, 4, 4 points, then tie break calculated for A, B players, and separately for C, D players
