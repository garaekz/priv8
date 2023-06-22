CREATE TABLE album
(
    id         VARCHAR PRIMARY KEY,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE users (
    id         VARCHAR PRIMARY KEY,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

INSERT INTO users (id, name, created_at, updated_at) VALUES ('1', 'user1', '2020-01-01 00:00:00', '2020-01-01 00:00:00');
