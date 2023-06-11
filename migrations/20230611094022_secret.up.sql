-- Create a postgresql secrets table for storing secrets
CREATE TABLE secrets (
    id         VARCHAR PRIMARY KEY,
    encrypted_data VARCHAR NOT NULL,
    ttl        INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);