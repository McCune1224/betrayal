CREATE TABLE IF NOT EXISTS category (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);

INSERT INTO category 
(name) 
VALUES 
('POSITIVE'),
('NEGATIVE'),
('NEUTRAL'),
('NON-VISITING'),
('NON VISITING'),
('INSTANT'),
('NIGHT'),
('ALTERATION'),
('REACTIVE'),
('REDIRECTION'),
('VISIT REDIRECTION'),
('VOTE AVOIDING'),
('VOTE CHANGE'),
('VOTE BLOCKING'),
('VOTE REDIRECTION'),
('INVESTIGATION'),
('PROTECTION'),
('VISIT'), 
('VISITING'), 
('VISIT BLOCKING'),
('BLOCKING'),
('VOTE'), 
('IMMUNITY'),
('VOTE IMMUNITY'),
('VOTE MANIPULATION'),
('SUPPORT'),
('DEBUFF'),
('THEFT'),
('HEALING'),
('DESTRUCTION'),
('KILLING');
