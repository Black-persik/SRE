package repository

import (
	"context"
	"errors"
	"fmt"
	"main/internal/entity"
	"main/internal/error"
	"main/pkg/postgres"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ResponseRepository struct {
	db *postgres.Client
}

func NewResponseRepository(db *postgres.Client) *ResponseRepository {
	return &ResponseRepository{db: db}
}

func (r *ResponseRepository) Create(ctx context.Context, response entity.CreateResponse) (entity.Response, error) {
	query := psql.Insert("responses").
		Columns("user_id", "company", "vacancy", "status").
		Values(response.UserID, response.Company, response.Vacancy, response.Status).
		Suffix("RETURNING response_id")

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.Response{}, fmt.Errorf("ResponseRepository.Create: to sql: %w", err)
	}

	var responseID uuid.UUID
	if err = r.db.Pool.QueryRow(ctx, sql, args...).Scan(&responseID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "responses_user_id_fkey" {
			return entity.Response{}, fmt.Errorf("ResponseRepository.Create: %w", apperror.ErrUserNotFound)
		}
		return entity.Response{}, fmt.Errorf("ResponseRepository.Create: query row: %w", err)
	}

	return entity.Response{
		ID:      responseID,
		UserID:  response.UserID,
		Company: response.Company,
		Vacancy: response.Vacancy,
		Status:  response.Status,
	}, nil
}

func (r *ResponseRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Response, error) {
	query := psql.Select("response_id", "user_id", "company", "vacancy", "status").
		From("responses").
		Where(squirrel.Eq{"response_id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.Response{}, fmt.Errorf("ResponseRepository.GetByID: to sql: %w", err)
	}

	var response entity.Response
	if err = r.db.Pool.QueryRow(ctx, sql, args...).
		Scan(&response.ID, &response.UserID, &response.Company, &response.Vacancy, &response.Status); err != nil {
		return entity.Response{}, fmt.Errorf("ResponseRepository.GetByID: query row: %w", err)
	}

	return response, nil
}

func (r *ResponseRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Response, error) {
	query := psql.Select("response_id", "user_id", "company", "vacancy", "status").
		From("responses").
		Where(squirrel.Eq{"user_id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ResponseRepository.GetByUserID: to sql: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ResponseRepository.GetByUserID: query: %w", err)
	}

	defer rows.Close()

	var responses []entity.Response

	for rows.Next() {
		var response entity.Response
		if err := rows.Scan(&response.ID, &response.UserID, &response.Company, &response.Vacancy, &response.Status); err != nil {
			return nil, fmt.Errorf("ResponseRepository.GetByUserID: scan: %w", err)
		}

		responses = append(responses, response)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ResponseRepository.GetByUserID: rows error: %w", err)
	}
	return responses, nil
}

func (r *ResponseRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResponseStatus) error {
	query := psql.Update("responses").
		Set("status", status).
		Where(squirrel.Eq{"response_id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("ResponseRepository.UpdateStatus: to sql: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ResponseRepository.UpdateStatus: exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *ResponseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := psql.Delete("responses").
		Where(squirrel.Eq{"response_id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("ResponseRepository.Delete: to sql: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("ResponseRepository.Delete: exec: %w", err)
	}

	return nil
}
