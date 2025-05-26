-- Drop indexes
DROP INDEX IF EXISTS idx_persons_age;
DROP INDEX IF EXISTS idx_persons_surname;
DROP INDEX IF EXISTS idx_persons_name;
DROP INDEX IF EXISTS idx_persons_nationality_id;
DROP INDEX IF EXISTS idx_persons_gender_id;

-- Drop tables
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS nationalities;
DROP TABLE IF EXISTS genders;