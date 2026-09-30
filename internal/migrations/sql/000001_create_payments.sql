CREATE TABLE IF NOT EXISTS payments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    status TEXT NOT NULL CHECK (status IN ('Authorized', 'Declined')),
    card_last_four CHAR(4) NOT NULL,
    expiry_month SMALLINT NOT NULL CHECK (expiry_month BETWEEN 1 AND 12),
    expiry_year SMALLINT NOT NULL,
    currency CHAR(3) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    authorization_code TEXT NOT NULL DEFAULT ''
);
