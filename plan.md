# Table tennis leagues

## Description

The application consists from backend and frontend applications.
The application is aimed for conducting the table tennis leagues.


## Business requirements

### High level description

The table tennis leagues are conducted monthly. The leagues are usually named wiht `{LETTER}-{NUMBER}`,
where letters mean division, and number is the group number inside a division. Order of letters is A,B,C,...etc.
where A is a top division. Numbers are ordered in the same manner where 1 is the top group in a division.
And there should be one 'Superleague'(A0) group above A1 group.
Each group in a league is played in a round-robin bracket with best of 5 matches. When umpires
enter scores for all matches in the group then the group can be finished. It means that placements are
calculated and the playes ratings are calculated. Rating system is `glicko2`.
The initial player creation can be done through `.csv` import or manually through form.
For each league to finish all mathes in all divisions and groups should be finished.
Before a league is finished it is possible to do changes in match scoring.
When the current month league is finished the draft for the next month league should  be created.
When the draft is created top `numberOfAdvances` players should be moved to the higher group, exception is
the topmost group where players just stay. When the draft is created bottom `numberOfRecedes` players should 
be moved to the lower group, exception is the lowest group where players just stay. The `numberOfAdvances` and
`numberOfRecedes` should be configurable in the league configuration. Games to win (best of 3 or best of 5 or
best of 7) also should be configured through league configuration. The configurable parameters can be changed
in the next month league, and if so, then the draft should be recreated.
If a player didn't participate in the group, then they automatically moved to the lowest place in the group.
If a player didn't participate in the group, then they can be replaced manually by other player who doesn't 
participate in other groups, but the rating would not be calculated for the newly added player and he should
be skipped in group placements. The player who didn't show by the start of the group automatically gets L:W
scoring for all their matches.

### Roles


The following roles are defined:
1. Maintainer - a person(s) who create leagues, seed players into leagues manually.
2. Umpire - a persons(s) who can enter scores and finish the group.
3. Player - a person(s) who can view the live scores 

Any person can register and by default gets Player role. Player can start a new league. 
A league starter becomes Maintainer only for created league. Maintainer can assign Maintainer/Umpire role to any 
registered person including Maintainers and Umpires of other leagues.

### Scoring and rating

1. The match score is enetered according to configurated `games to win` parameter. There is no draws, 
only one player can win
2. The scoring points for the win is 2, and for the loss is 1. In case the player didn't show then the score 
whould be written  as W:L against the player. Which effectively is 2:0 in scoring points
3. Placement in the group is done by the following rules
    - The players are ordered by scoring points
    - In case of equal points for two players the winner of the match between those two players is placed higher
    - In case of equal points for more than two players the tiebreak points are calculated:
        - tiebreak point example: playerA won two games with 3:0 score and lost 2:3, playerB won 3:1, 3:2 and 
          lost 0:3, playerC won 3:0, 3:1 and lost 0:3. tiebreak points are calculated by delta beteen games won
          and games lost. playerA has 3+3+2-3=7 tiebreak points, playerB has 3+3-1-2=3 tiebreak points,
          playerC has 3+3-1-3=2 tiebreak points. Hence playerA's placement in highest, playerB is in the middle,
          and playerC is the lowest
    - In case of equal tiebreak points between two players the winner of the match between them is placed
      higher
    - In case of equal tiebreak points between more then two players the manual placing by umpire is required and 
      related notification should be shown to the umpire
4. Glicko2 ratings are calculated when the group is finished on per match basis. If the changes are done to finished
group all the ratings must be recalculated

### UI Views

- Player creation form
- Player `csv` import
- Player profile - with rating, leagues and groups, where player participated and rating deltas for each match
- Top players with filters by rating or other parameteres
- List of leagues that are conducted within the app
- League configuration form (when draft is created) - for maintainers
- Live view (all scores groups in a league):
    - for maintainers and umpires also should include entering match scores
    - for maintainers should also allow adding non-calculated players to the group and setting the `did not show` status


## Techincal design

### Backend application tech stack:

- `GoLang`
- `Gin` Framework for api and websocket updates of live scores
- `Postgres` for storing data
- `sqlx` library for communcation with database
- `go-migrations` for schema changes
- test-tables approach for unit tests
- for testing data access layer the `sqlite` in mem should be used

### Backend principles
- All schema changes are done only through migrations
- API follows the REST approach
- The authentication is done with OAuth protocol

### Frontend application tech stack:

1. Libraries:
    - `React` SPA with typescript with vite
    - `React router` in data mode for routing
    - `Tailwind` for component styling
    - `Vitest` for testing
    - `esling` for linting
    - `prettier` for code formatting
    - `cypress` for integration testing

2. React Components:
   - Each component should be atomic and is a pure function of its props and state. 
   - Each component should be self-contained and reusable.
   - Each component should be covered with unit and integration tests.
   - Compoments are stored in `src/components`
   - Each component should be stored in a separate directory with its tests and styles
   - When a new component is created, first search for existing components to reuse.
3. API calls:
   - API calls are made using `axios`
   - API calls are stored in `src/api`
   - When a new API call is created, related hooks should be added to `src/hooks`
4. Testing:
   - Unit tests are written using `vitest`
   - Integration tests are written using `cypress` 


### Data model

```pluntuml

@startuml
hide circle
skinparam linetype ortho
skinparam entity {
  BackgroundColor #FAFAFA
  BorderColor #555
  FontSize 12
}

entity roles {
    * role_id: serial
    --
    role_name: varchar
}

entity users {
    * user_id: bigserial
    --
    first_name: varchar not null
    last_name: varchar not null
    created: timestamp not null
    last_updated: timestamp not null

    current_rating: double not null
    deviation : double not null
    volatility : double not null
}

entity leagues {
    * league_id: bigserial
    --
    title: varchar not null
    description: text not null
    created: timestamp not null
    last_updated: timestamp not null
    configration: jsonb not null
}

entity league_events {
    * event_id: bigserial
    league_id: bigint <<FK>>
    --
    status: varchar not null [DRAFT, IN_PROGRESS, DONE]
    title: varchar not null
    startdate: date not null
    enddate: date not null
    created: timestamp not null
    last_updated: timestamp not null
}

entity groups {
    * group_id: bigserial
    event_id: bigint <<FK>>
    --
    status: varchar not null [DRAFT, IN_PROGRESS, DONE]
    division: varchar not null
    group_no: int no null
    scheduled: datetime not null
    created: timestamp not null
    last_updated: timestamp not null
}

entity group_players {
    * group_player_id: bigserial
    league_id: bigint <<FK>>
    user_id: bigint <<FK>>
    --
    seed: smallint not null
    place: smallint not null default 0
    points: smallint not null default 0
    tiebreak_points: smallint not null default 0
    advances: boolean not null default false
    recedes: boolean not null default false

    created: timestamp not null
    last_updated: timestamp not null
}

entity user_roles {
    user_id: bigint <<FK>>
    role_id: int <<FK>>
    league_id: bigint <<FK>>
    --
    created: timestamp not null
    last_updated: timestamp not null
}

entity matches {
    * match_id: bigserial
    group_id: bigint <<FK>>
    --
    group_player1_id: bigint <<nullable>>
    group_player2_id: bigint <<nullable>>
    score1: smallint
    score1: smallint
    withdraw1: boolean not null default false
    withdraw2: boolean not null default false

    status: varchar not null [DRAFT, IN_PROGRESS, DONE]

    created: timestamp not null
    last_updated: timestamp not null
}

entity rating_history {
    * history_id: bigserial
    user_id: bigint <<FK>>
    match_id: bigint <<FK>>
    --
    delta: double not null
    rating: double not null
    deviation : double not null
    volatility : double not null
}

roles        ||--o{ user_roles
users        ||--o{ user_roles
leagues      ||--o{ user_roles
leagues      ||--o{ league_events
league_events     ||--o{ groups
groups      ||--o{ group_players
users      ||--o{ group_players
groups    ||--o{ matches
matches      ||--o{ rating_history
users      ||--o{ rating_history

@enduml

```

