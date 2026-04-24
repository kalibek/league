
### Group view improvements:
1. Group should be sorted by initial seeding
2. There should names and not Player #{id} in the group list and matrix
3. Put `did not show` button inline with the player name and make it icon button

### Score entering form improvements:
1. Show dynamic buttons for possible scores instead of number inputs. E.g. in case of best of 3 the buttons should be 2-0, 2-1, 0-2, 1-2


# User stories

## Admin role
### Description:
As an admin, I want to be able to create leagues and assign Maintainers to them.
### Acceptance criteria:
- Only Admin can create leagues
- Admin and Maintainers can assign Maintainers to leagues

## Next league event
### Description:
As a maintainer, I want to be able to create a new league event when the current one is over.
### Acceptance criteria:
- When the current league event is over, a new one should be created with the same settings as the previous one.
- The new league event should be filled with same number of players as the previous one.
- The new league event should properly move players from the previous one to the new one according to the advances/recedes from the previous event.

## Player profile view
### Description:
As any user, I want to be able to see any players profile and see player's participation in leagues with rating delta changes and match scores ordered by date from newest to oldest
### Acceptance criteria:
- The player's history should load 5 most recent events. Older events should be shown on scroll.

## Player profile edit
### Description:
As a player, I want to be able to edit my profile. If I didn't fill my profile yet, I would like to see in app notification about it until I fill it.
### Acceptance criteria:
- The player's profile should be editable.
- The editable fields are the following:
  - First name
  - Last name
  - Country - should be a dropdown with all countries from the list. Countries should be created in database migration
  - City - should be a dropdown list with the ability to add a new city
  - Birthdate
  - Grip (penhold, shakehand)
  - Gender
  - Setup:
    - Blade - should be a dropdown list with the ability to add a new blade
    - Forehand Rubber - should be a dropdown list with the ability to add a new rubber (use the same table as for backhand)
    - Backhand Rubber - should be a dropdown list with the ability to add a new rubber (use the same table as for forehand)

## Player profile search
### Description:
As any user, I want to be able to search for players by all fields in the profile
### Acceptance criteria:
- The search should be case insensitive

## Internationalization
### Description:
As any user, I want to be able to change the language of the app.
### Acceptance criteria:
- The library should be `react-i18next`
- All the static texts should be translated in the frontend app
- The app should support 3 languages:
    - English
    - Russian
    - Kazakh
  
## Leagues, events pagination
### Description:
As any user, I want to see only recent events in the leagues and events list. Same for players lists and player searches
### Acceptance criteria:
- The older events should be shown on scroll.
- Size should be 10 events per page and configurable