-- +goose Up
ALTER TABLE refuges ADD COLUMN state jsonb NOT NULL DEFAULT '{}';

UPDATE refuges SET state = jsonb_build_object(
    'defense', round(defense * 6 / 10.0)::int,
    'morale', round(morale * 6 / 10.0)::int,
    'belligerence', round(belligerence * 6 / 10.0)::int,
    'mobility', mobility,
    'population', jsonb_build_object('level', 0, 'max', 10),
    'reservesMax', greatest(10, water + food + wood),
    'hasWaterSource', false,
    'resources', jsonb_build_object('water', water, 'food', food, 'wood', wood),
    'region', jsonb_build_object('size', 0, 'danger', 0, 'habitation', 0, 'naturalResources', 0,
        'contamination', 0, 'travel', 0, 'abundance', '{}'::jsonb),
    'week', 1,
    'buildings', '[]'::jsonb, 'crises', '[]'::jsonb, 'residents', '[]'::jsonb);

ALTER TABLE refuges
    DROP COLUMN defense, DROP COLUMN morale, DROP COLUMN mobility, DROP COLUMN belligerence,
    DROP COLUMN water, DROP COLUMN food, DROP COLUMN wood;

-- +goose Down
ALTER TABLE refuges
    ADD COLUMN defense int NOT NULL DEFAULT 0, ADD COLUMN morale int NOT NULL DEFAULT 0,
    ADD COLUMN mobility int NOT NULL DEFAULT 0, ADD COLUMN belligerence int NOT NULL DEFAULT 0,
    ADD COLUMN water int NOT NULL DEFAULT 0, ADD COLUMN food int NOT NULL DEFAULT 0,
    ADD COLUMN wood int NOT NULL DEFAULT 0;
UPDATE refuges SET defense = least(10, round((state->>'defense')::int * 10 / 6.0)::int),
    morale = least(10, round((state->>'morale')::int * 10 / 6.0)::int),
    belligerence = least(10, round((state->>'belligerence')::int * 10 / 6.0)::int),
    mobility = least(10, (state->>'mobility')::int),
    water = coalesce((state->'resources'->>'water')::int, 0),
    food = coalesce((state->'resources'->>'food')::int, 0),
    wood = coalesce((state->'resources'->>'wood')::int, 0);
ALTER TABLE refuges DROP COLUMN state;
