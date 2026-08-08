CREATE TABLE IF NOT EXISTS ability_category (
  ability_id INT NOT NULL,
  category_id INT NOT NULL,
  PRIMARY KEY (ability_id, category_id)
);
