-- name: FindAllIngredients :many
SELECT name AS value FROM ingredients ORDER BY name ASC;

-- name: FindMatchingIngredients :many
SELECT value, score
FROM (
SELECT
  name AS value, similarity(name, sqlc.arg(search)::text) as score
FROM
ingredients
ORDER BY 
score DESC
) AS matches
WHERE
score >= sqlc.arg(threshold)::float;

-- name: FindAllTags :many
SELECT DISTINCT
tag::text AS value
FROM
  recipe_tags
ORDER BY
  tag ASC;

-- name: FindMatchingTags :many
SELECT DISTINCT value::text, score
FROM (
  SELECT
tag AS value, similarity(tag, sqlc.arg(search)::text) as score
FROM
  recipe_tags
  ORDER BY
  score DESC
) AS matches
WHERE
  score >= sqlc.arg(threshold)::float;
  
