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
