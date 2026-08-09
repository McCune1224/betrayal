-- Reconcile pre-existing duplicate catalog names before enforcing uniqueness.
-- Keep the lowest id as canonical and preserve all relationship rows by
-- remapping them first. The migration runs in one transaction, so an error
-- does not leave a partially reconciled catalog.
CREATE TEMP TABLE _role_name_map ON COMMIT DROP AS
SELECT id, min(id) OVER (PARTITION BY name) AS canonical_id
FROM role;
UPDATE player p
SET role_id = m.canonical_id
FROM _role_name_map m
WHERE p.role_id = m.id AND m.id <> m.canonical_id;
UPDATE ability_info a
SET role_specific_id = m.canonical_id
FROM _role_name_map m
WHERE a.role_specific_id = m.id AND m.id <> m.canonical_id;
INSERT INTO role_ability (role_id, ability_id)
SELECT m.canonical_id, ra.ability_id
FROM role_ability ra
JOIN _role_name_map m ON m.id = ra.role_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
INSERT INTO role_perk (role_id, perk_id)
SELECT m.canonical_id, rp.perk_id
FROM role_perk rp
JOIN _role_name_map m ON m.id = rp.role_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
DELETE FROM role_ability ra
USING _role_name_map m
WHERE ra.role_id = m.id AND m.id <> m.canonical_id;
DELETE FROM role_perk rp
USING _role_name_map m
WHERE rp.role_id = m.id AND m.id <> m.canonical_id;

CREATE TEMP TABLE _item_name_map ON COMMIT DROP AS
SELECT id, min(id) OVER (PARTITION BY name) AS canonical_id
FROM item;
INSERT INTO item_category (item_id, category_id)
SELECT m.canonical_id, ic.category_id
FROM item_category ic
JOIN _item_name_map m ON m.id = ic.item_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
INSERT INTO player_item (player_id, item_id, quantity)
SELECT pi.player_id, m.canonical_id, pi.quantity
FROM player_item pi
JOIN _item_name_map m ON m.id = pi.item_id AND m.id <> m.canonical_id
ON CONFLICT (player_id, item_id)
DO UPDATE SET quantity = GREATEST(player_item.quantity, EXCLUDED.quantity);
DELETE FROM item_category ic
USING _item_name_map m
WHERE ic.item_id = m.id AND m.id <> m.canonical_id;
DELETE FROM player_item pi
USING _item_name_map m
WHERE pi.item_id = m.id AND m.id <> m.canonical_id;

CREATE TEMP TABLE _ability_name_map ON COMMIT DROP AS
SELECT id, min(id) OVER (PARTITION BY name) AS canonical_id
FROM ability_info;
INSERT INTO role_ability (role_id, ability_id)
SELECT ra.role_id, m.canonical_id
FROM role_ability ra
JOIN _ability_name_map m ON m.id = ra.ability_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
INSERT INTO player_ability (player_id, ability_id, quantity)
SELECT pa.player_id, m.canonical_id, pa.quantity
FROM player_ability pa
JOIN _ability_name_map m ON m.id = pa.ability_id AND m.id <> m.canonical_id
ON CONFLICT (player_id, ability_id)
DO UPDATE SET quantity = GREATEST(player_ability.quantity, EXCLUDED.quantity);
DELETE FROM role_ability ra
USING _ability_name_map m
WHERE ra.ability_id = m.id AND m.id <> m.canonical_id;
DELETE FROM player_ability pa
USING _ability_name_map m
WHERE pa.ability_id = m.id AND m.id <> m.canonical_id;
DELETE FROM ability_info a
USING _ability_name_map m
WHERE a.id = m.id AND m.id <> m.canonical_id;

CREATE TEMP TABLE _perk_name_map ON COMMIT DROP AS
SELECT id, min(id) OVER (PARTITION BY name) AS canonical_id
FROM perk_info;
INSERT INTO role_perk (role_id, perk_id)
SELECT rp.role_id, m.canonical_id
FROM role_perk rp
JOIN _perk_name_map m ON m.id = rp.perk_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
INSERT INTO player_perk (player_id, perk_id)
SELECT pp.player_id, m.canonical_id
FROM player_perk pp
JOIN _perk_name_map m ON m.id = pp.perk_id AND m.id <> m.canonical_id
ON CONFLICT DO NOTHING;
DELETE FROM role_perk rp
USING _perk_name_map m
WHERE rp.perk_id = m.id AND m.id <> m.canonical_id;
DELETE FROM player_perk pp
USING _perk_name_map m
WHERE pp.perk_id = m.id AND m.id <> m.canonical_id;
DELETE FROM perk_info p
USING _perk_name_map m
WHERE p.id = m.id AND m.id <> m.canonical_id;

DELETE FROM role r
USING _role_name_map m
WHERE r.id = m.id AND m.id <> m.canonical_id;
DELETE FROM item i
USING _item_name_map m
WHERE i.id = m.id AND m.id <> m.canonical_id;

CREATE UNIQUE INDEX IF NOT EXISTS uq_role_name ON role (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_item_name ON item (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_ability_info_name ON ability_info (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_perk_info_name ON perk_info (name);
