CREATE TABLE IF NOT EXISTS status (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  hour_duration INT NOT NULL DEFAULT 0
);

INSERT INTO status (name, description, hour_duration)
VALUES 
('Cursed', 'If it isn’t removed within three days, you will die.', 36),
('Death Cursed', 'Just like curse, except it can only be removed by the Wizard''s Blessing or the Siren''s Salt Water Embrace.', 36),
('Frozen', 'You can’t use any abilities until you thaw out. 1/3 chance you’ll thaw each new day. Thaws out on the 3rd day if not thawed out before.', 36),
-- WARNING: It's technically 48 hours, however it does not go into effect until an item or ability has been used, so by default keeping it at 0.
('Paralyzed', 'Activates once you have used an AA, item or a base ability. You have a 48 hour cooldown before you can use another ability or item. Permanent until cured.', 0),
('Burned', 'Every 24 hours from when you''re inflicted, you will lose an item from rarity descending. You cant pass items whilst burning. Permanent until cured.', 0),
('Empowered', 'You can use any one of your abilities for 2 days, even if you ran out of them. When you use an ability, it''s a 12 hour cooldown before you can use another. If you use a killing ability, you use it and then you no longer have Empowered.', 48),
('Drunk', '25% chance to target a random person instead when using an item or an ability. Wears off after 48 hours. This effect stacks.', 48),
('Restrained', 'Cannot use abilities. Permanent until cured.', 0),
('Disabled', 'You can’t vote at the Elimination Phases. Permanent until cured.', 0),
('Blackmailed', 'You can’t talk for a day and you can’t vote at that Elimination Phase. Removes after 24 hours.', 24),
('Despaired', 'You vote for yourself at Elimination Phase. Permanent until cured.', 0),
('Madness', 'When inflicted with madness you must make efforts to present yourself as the role you''ve been made mad about. Anything deviating from that will count as breaking madness, and will result in death. This status lasts 12 hours unless otherwise stated.', 12),
('Lucky', '2x luck, 1.5x coins, removed on gaining another status', 0),
('Unlucky', '0.5x luck, 0.5x coins, removed on gaining another status', 0);

