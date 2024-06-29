CREATE TABLE IF NOT EXISTS status (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL
);

-- Cursed	If it isn’t removed within three days, you will die.
-- Death Cursed	Just like curse, except it can only be removed by the Wizard's Blessing or the Siren's Salt Water Embrace.
-- Frozen	You can’t use any abilities until you thaw out. 1/3 chance you’ll thaw each new day. Thaws out on the 3rd day if not thawed out before.
-- Paralyzed	Activates once you have used an AA, item or a base ability. You have a 48 hour cooldown before you can use another ability or item. Permanent until cured.
-- Burned	Every 24 hours from when you're inflicted, you will lose an item from rarity descending. You cant pass items whilst burning. Permanent until cured.
-- Empowered	You can use any one of your abilities for 2 days, even if you ran out of them. When you use an ability, it's a 12 hour cooldown before you can use another. If you use a killing ability, you use it and then you no longer have Empowered.
-- Drunk	25% chance to target a random person instead when using an item or an ability. Wears off after 48 hours. This effect stacks.
-- Restrained	Cannot use abilities. Permanent until cured.
-- Disabled	You can’t vote at the Elimination Phases. Permanent until cured.
-- Blackmailed	You can’t talk for a day and you can’t vote at that Elimination Phase. Removes after 24 hours.
-- Despaired	You vote for yourself at Elimination Phase. Permanent until cured.
-- Madness	When inflicted with madness you must make efforts to present yourself as the role you've been made mad about. Anything deviating from that will count as breaking madness, and will result in death. This status lasts 12 hours unless otherwise stated.
-- Lucky	2x luck, 1.5x coins, removed on gaining another status
-- Unlucky	0.5x luck, 0.5x coins, removed on gaining another status
INSERT INTO status (name, description) 
VALUES 
('Cursed', 'If it isn’t removed within three days, you will die.'),
('Death Cursed', 'Just like curse, except it can only be removed by the Wizard''s Blessing or the Siren''s Salt Water Embrace.'),
('Frozen', 'You can’t use any abilities until you thaw out. 1/3 chance you’ll thaw each new day. Thaws out on the 3rd day if not thawed out before.'),
('Paralyzed', 'Activates once you have used an AA, item or a base ability. You have a 48 hour cooldown before you can use another ability or item. Permanent until cured.'),
('Burned', 'Every 24 hours from when you''re inflicted, you will lose an item from rarity descending. You cant pass items whilst burning. Permanent until cured.'),
('Empowered', 'You can use any one of your abilities for 2 days, even if you ran out of them. When you use an ability, it''s a 12 hour cooldown before you can use another. If you use a killing ability, you use it and then you no longer have Empowered.'),
('Drunk', '25% chance to target a random person instead when using an item or an ability. Wears off after 48 hours. This effect stacks.'),
('Restrained', 'Cannot use abilities. Permanent until cured.'),
('Disabled', 'You can’t vote at the Elimination Phases. Permanent until cured.'),
('Blackmailed', 'You can’t talk for a day and you can’t vote at that Elimination Phase. Removes after 24 hours.'),
('Despaired', 'You vote for yourself at Elimination Phase. Permanent until cured.'),
('Madness', 'When inflicted with madness you must make efforts to present yourself as the role you''ve been made mad about. Anything deviating from that will count as breaking madness, and will result in death. This status lasts 12 hours unless otherwise stated.'),
('Lucky', '2x luck, 1.5x coins, removed on gaining another status'),
('Unlucky', '0.5x luck, 0.5x coins, removed on gaining another status');

