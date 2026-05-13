# features
### call to table
During the league setup there should be a parameter for number of tables. e.g 10, means that there can be maximum 10 matches in the league simultaneously.
The match has to have additional field "table" with number of table.
When umpire or owner sets the table number to the match then the match is automatically trasitions to the "IN PROGRESS" status.
The table number should be a modal with a drop down list of tables. The table numbers that are currently assigned to any match "IN PROGRESS" should not be available for choosing.
The table number should be visible in the match details.
The match with in progres status should be highlighted in yellow in GroupStandings and MatchGrid components on UI

The table is calculated according to the following rules:

- tie break calc according to federation rules
- adding players during the run
- W-L and DNS is different status
- all DNS should be receded, if recedes more that planned, then more advances from the group below
- export to excel by group


in the `docs` folder there are the excel files with results of leagues starting from 2026 January, each league has its own sheet.
extract the players into csv file with the following format: `first_name, last_name, email, initial_rating`,
where:
    - `first_name` and `last_name` are the players names
    - `email` is the just first_name_last_name lowercased and translitterated and with @tt-league.kz appended
    - `initial_rating` is the rating by the league . top league is a superleague with 2500 rating and 
    the rating step to the next league is 50, e.g. A1 league players should have rating 2450, A2 2400, ..., A8 2100, B1 2050, B2 2000, etc.
Notes: 
    - the files are in cyrillic, so the A1 league, for example, can be in cyrillic А1, etc. some leagues have `_2` in their names which should be ignored 
    - if the player appears in multiple files, then the rating should be from earliest appearance.
Put results into the `docs/players.csv` file.


## League setup features
### Description:
Following user stories should be implemented:
- As an owner of the league, I want to be able to remove empty groups from the event when the league is in 'DRAFT' or 'IN PROGRESS' status.
- As an owner of the league, I want to see names of the players in the groups instead of their player IDs.
- As an owner or umpire of the league, I want to collapse not only by group but also by divisions, so that I could quickly navigate between specific divisions and groups

## Group View
### Description
As an anonymous user or a player, I want to have a separate page for a group in the event, so that I could track progress and view calls for specific group
### Acceptance Criteria
- The group page should be accessible by a link from the group list
- The group page should contain the same information as an event page but only for the selected group

## Excel export
### Description
As an owner of the league, I want to be able to export the league results to an Excel file.
### Acceptance Criteria
- The league results should be exported to an Excel file in the same format as the live view page
- The file should take header and footer from the template
- The file should be generated on the fly and available for download, no permanent storage required

## Registration feature flag
### Description
As an owner of the system, I want to enable or disable the registration feature.
### Acceptance Criteria
- The registration feature should be disabled by default
- The registration feature should be enabled by the admin