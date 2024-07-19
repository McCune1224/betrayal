CREATE TABLE IF NOT EXISTS role_perk (
  role_id INT NOT NULL,
  perk_id INT NOT NULL,
  PRIMARY KEY (role_id, perk_id)
);
