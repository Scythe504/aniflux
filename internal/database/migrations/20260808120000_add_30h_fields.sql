-- +goose Up
ALTER TABLE airing_schedule 
ADD COLUMN schedule_day TEXT NOT NULL DEFAULT '',
ADD COLUMN timing TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE airing_schedule 
DROP COLUMN schedule_day,
DROP COLUMN timing;
