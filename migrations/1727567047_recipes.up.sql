CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS ingredients(
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR (100) UNIQUE NOT NULL,
  kind VARCHAR (50),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_your_table_updated_at
BEFORE UPDATE ON ingredients 
FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS recipes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title VARCHAR (100) NOT NULL,
  headline TEXT,
  description TEXT,
  steps TEXT,
  servings INTEGER,
  url VARCHAR (255),
  tags TEXT[],
  slug TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE recipes ADD CONSTRAINT unique_title UNIQUE (title);
ALTER TABLE recipes ADD CONSTRAINT unique_slug UNIQUE (slug);
CREATE INDEX idx_recipes_slug ON recipes(slug);

CREATE TRIGGER update_your_table_updated_at
BEFORE UPDATE ON recipes
FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS recipe_ingredients (
  recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
  ingredient_id UUID REFERENCES ingredients(id),
  unit VARCHAR (50) NOT NULL, 
  quantity FLOAT NOT NULL,
  PRIMARY KEY (recipe_id, ingredient_id)
);
