CREATE TABLE IF NOT EXISTS item_category (
  item_id INT NOT NULL,
  category_id INT NOT NULL,
  PRIMARY KEY (item_id, category_id)
);
