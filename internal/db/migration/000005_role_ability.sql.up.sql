CREATE TABLE IF NOT EXISTS role_ability (
  role_id INT NOT NULL,
  ability_id INT NOT NULL,
  PRIMARY KEY (role_id, ability_id)
);
