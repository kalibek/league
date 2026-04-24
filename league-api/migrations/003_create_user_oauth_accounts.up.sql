CREATE TABLE user_oauth_accounts (
    account_id   BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider     VARCHAR(20) NOT NULL,
    provider_sub VARCHAR(255) NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_sub)
);
