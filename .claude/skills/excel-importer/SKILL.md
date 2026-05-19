---
name: excel-importer
description: Use when the league event data should be imported from an excel file.
argument-hint: "<file> <event-link>"
---

## Arguments
- The league event data is stored $0 file and the event link is $1

## Steps:

1. Confirm the arguments with the user before starting other steps
2. Verify that the file $0 is an excel file
3. The file should be in the `./docs` directory
4. Verify that the event is in `DRAFT` status
5. Use API to import the data with the following credentials:
    - username: `admin@tt-league.kz`
    - password (from `./.env` file): ${ADMIN_PASSWORD}
6. Verify that the seeding in the excel file and in the event are the same
7. Start the event
8. Enter scores from the excel file
9. Use and modify when needed `./docs/seed_january.py` and `./docs/enter_results.py`
10. Use search API to find players and create players only if they don't exist


## Excel structure:
### Sheets titles
- Each sheet represents a group in an event.
- Sheet name is Division+GroupNo (some sheets can have underscore and number in their name that could be ignored)
- Some sheets can have a cyrillic division letter instead of latin (e.g А instead of A, В instead of B, etc.)
### Group structure
- Each sheet consists of a group matrix that usually starts row 6 column 7
- Player names are in the column 10 of the group matrix (highlighted in green are the players who shown up)
- The format for player names is `{LastName} {FirstName}`. Sometimes a '+' sign is added to the end of the name that should be ignored
- Underneath the group matrix is a table with inidividual match results in the format `Player1 - Player2 Score`