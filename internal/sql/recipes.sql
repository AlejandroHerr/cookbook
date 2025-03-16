-- name: List :many
SELECT * FROM recipes ORDER BY created_at DESC;

-- name: GetByID :one
SELECT * FROM recipes
WHERE id = $1 LIMIT 1; 

-- name: GetBySlug :one
SELECT * FROM recipes
WHERE slug = $1 LIMIT 1;

-- name: GetSlugs :many
SELECT slug FROM recipes 
WHERE slug = $1 
OR slug ~ ($1 || '-[0-9]+$');

-- name: GetRecipeIngredients :many
SELECT 
  i.id, i.name, i.kind, ri.unit, ri.quantity
FROM 
recipe_ingredients ri 
LEFT JOIN 
  ingredients i ON i.id = ri.ingredient_id 
WHERE 
ri.recipe_id = $1;

-- name: CreateRecipe :one
INSERT INTO
  recipes (id, title, slug, headline, description, steps, prep_time, servings, url, tags)
VALUES 
  (@id, @title, @slug, @headline, @description, @steps, @prep_time, @servings, @url, @tags)
RETURNING
  *;

-- name: UpdateRecipe :one
UPDATE 
  recipes 
SET
  title = @title,
  slug = @slug,
  headline = @headline,
  description = @description,
  steps = @steps,
  prep_time = @prep_time,
  servings = @servings,
  url = @url,
  tags = @tags
WHERE 
  id = @id
RETURNING
  *;

-- name: DeleteRecipe :execrows
DELETE FROM recipes
WHERE id = $1;

-- name: CreateIngredient :one
INSERT INTO
  ingredients (name, kind)
VALUES 
  (@name, @kind)
RETURNING
  *;
-- name: UpsertIngredient :one
INSERT INTO ingredients (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE 
SET name = EXCLUDED.name
RETURNING *;

-- name: CreateRecipeIngredient :execrows
INSERT INTO
recipe_ingredients (recipe_id, ingredient_id, unit, quantity) 
VALUES 
(@recipe_id, @ingredient_id, @unit, @quantity);

-- name: DeleteRecipeIngredients :execrows
DELETE FROM recipe_ingredients
WHERE recipe_id = $1;

