CREATE TABLE games (
  id VARCHAR(16),
  date DATE,
  status VARCHAR(16),
  type SMALLINT DEFAULT 0,
  PRIMARY KEY (id)
);
CREATE INDEX games_date ON games (date);
CREATE INDEX games_status ON games (status);
CREATE INDEX games_type ON games (type);

CREATE TABLE homeruns (
  game VARCHAR(16),
  batter VARCHAR(16),
  number VARCHAR(8),
  scenario VARCHAR(8),
  pitcher VARCHAR(16),
  PRIMARY KEY (game, batter, number)
);
