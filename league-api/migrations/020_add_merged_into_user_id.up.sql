ALTER TABLE users ADD COLUMN merged_into_user_id BIGINT REFERENCES users(user_id);
