package ingredients

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/core"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

func (r PgRepository) Find(id uuid.UUID) (*Ingredient, error) {
	return r.findBy("id", id.String())
}

func (r PgRepository) FindByName(name string) (*Ingredient, error) {
	return r.findBy("name", name)
}

func (r PgRepository) findBy(field string, value string) (*Ingredient, error) {
	row := r.pool.QueryRow(context.TODO(), "SELECT * from ingredients WHERE "+field+" = $1", value)
	ingredientDB := struct {
		ID   uuid.UUID
		Name string
		Kind *string
	}{}

	err := row.Scan(&ingredientDB.ID, &ingredientDB.Name, &ingredientDB.Kind)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &core.NotFoundError{Field: field, Value: value, Err: err}
		}

		return nil, fmt.Errorf("error scanning ingredient: %w", err)
	}

	ingredient := New(ingredientDB.ID, ingredientDB.Name, ingredientDB.Kind)

	return ingredient, nil
}

func (r PgRepository) Save(ingredient *Ingredient) error {
	values := pgx.NamedArgs{
		"id":   ingredient.ID(),
		"name": ingredient.Name(),
		"kind": ingredient.Kind(),
	}

	_, err := r.pool.Exec(
		context.TODO(),
		"INSERT INTO ingredients (id, name, kind) VALUES (@id,@name,@kind)",
		values,
	)
	if err != nil {
		return core.HandleError(err)
	}

	return nil
}
