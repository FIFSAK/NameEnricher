CREATE TABLE IF NOT EXISTS genders
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(10) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS nationalities
(
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS persons
(
    id             SERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    surname        TEXT NOT NULL,
    patronymic     TEXT,
    age            INT  NOT NULL,
    gender_id      INT  NOT NULL REFERENCES genders (id),
    nationality_id INT  NOT NULL REFERENCES nationalities (id)
);

CREATE INDEX idx_persons_gender_id ON persons (gender_id);
CREATE INDEX idx_persons_nationality_id ON persons (nationality_id);

CREATE INDEX idx_persons_name ON persons (name);
CREATE INDEX idx_persons_surname ON persons (surname);
CREATE INDEX idx_persons_age ON persons (age);
