CREATE TABLE transaction_history(
    transaction_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    transaction_type INTEGER NOT NULL,
    amount INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user_balance(user_id)
);
