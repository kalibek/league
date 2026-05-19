# About This App

**League Manager** is a web application for organizing and running table tennis leagues. It is designed for clubs, recreational centers, and competitive communities that want a structured, fair, and transparent competition system.

## Who Is It For

- Table tennis clubs running regular league events
- Tournament organizers who need automated scheduling and standings
- Players who want to track their progress and rating over time

## How It Works

Players are organized into **groups** (S, A1, A2, B1, B2, and so on) for each monthly event. Within every group a full **round-robin** is played — every player faces every other player. At the end of the event, top finishers advance to a higher group and bottom finishers recede to a lower group, creating a self-sorting competitive ladder.

## Key Features

- **Group-based round-robin** scheduling generated automatically
- **Glicko2 rating system** — a statistically rigorous algorithm that tracks each player's skill, uncertainty, and consistency across events
- **Live scoring** — scores update in real time via WebSocket so spectators and players can follow along
- **DNS handling** — Did Not Show players are tracked and their absence is factored into promotions and relegations
- **Tiebreak resolution** — four-stage tiebreak logic ensures fair placement even when players finish level on points
- **Multi-language interface** — available in English, Russian, and Kazakh
- **Admin tools** — league admins can configure group sizes, advancement rules, number of tables, and more

## Rating

Every player carries a **Glicko2 rating** that updates after each group event. New players start at 1500 with high uncertainty; the rating becomes more stable as they play more events. See the [Rating](rating.md) page for a full explanation.

## Group Movement

Players move between groups based on their finish position and configured advance/recede counts. See the [Movements](movements.md) page for details on how promotions, relegations, and tiebreaks work.
