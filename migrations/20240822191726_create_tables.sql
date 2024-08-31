-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
     id serial PRIMARY KEY,
     login varchar(64) UNIQUE NOT NULL CHECK (login SIMILAR TO '[\w\.\-]+'),
     password_hash varchar(256) NOT NULL CHECK (password_hash SIMILAR TO '[\w\.\-]+')
);

CREATE TABLE balance (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id) UNIQUE,
    current numeric NOT NULL DEFAULT 0 CHECK (current >= 0),
    withdrawn numeric NOT NULL DEFAULT 0 CHECK (withdrawn >= 0)
);

CREATE TABLE "order" (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES "user"(id),
    number text UNIQUE NOT NULL CHECK (number SIMILAR TO '[0-9]+'),
    status varchar(16) NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual numeric CHECK (accrual >= 0),
    uploaded_at timestamp NOT NULL DEFAULT now()
);

CREATE TABLE withdrawal (
    id serial PRIMARY KEY,
    user_id integer NOT NULL DEFAULT NULL REFERENCES "user"(id),
    order_number text UNIQUE NOT NULL CHECK (order_number SIMILAR TO '[0-9]+'),
    sum numeric NOT NULL CHECK (sum >= 0),
    processed_at timestamp NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE withdrawal;

DROP TABLE "order";

DROP TABLE balance;

DROP TABLE "user";
-- +goose StatementEnd
