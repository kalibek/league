ALTER TABLE group_players
    ADD COLUMN IF NOT EXISTS player_status VARCHAR(10) NOT NULL DEFAULT 'active'
        CHECK (player_status IN ('active', 'dns'));
