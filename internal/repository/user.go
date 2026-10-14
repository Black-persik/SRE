package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"main/internal/entity"
	"main/internal/error"
	"main/pkg/postgres"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type UserRepository struct {
	db *postgres.Client
}

func NewUserRepository(db *postgres.Client) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user entity.CreateUser) (entity.User, error) {
	query := psql.Insert("users").
		Columns("username", "email").
		Values(user.Username, user.Email).
		Suffix("RETURNING user_id")

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepository.Create: to sql: %w", err)
	}

	var id uuid.UUID
	if err = r.db.Pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return entity.User{}, fmt.Errorf("UserRepository.Create: %w", apperror.ErrUsernameTaken)
			case "users_email_key":
				return entity.User{}, fmt.Errorf("UserRepository.Create: %w", apperror.ErrEmailTaken)
			}
		}
		return entity.User{}, fmt.Errorf("UserRepository.Create: query row: %w", err)
	}

	return entity.User{
		ID:       id,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	query := psql.Select("user_id", "username", "email").
		From("users").
		Where(squirrel.Eq{"user_id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepository.GetByID: to sql: %w", err)
	}

	var user entity.User
	if err = r.db.Pool.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Username, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("UserRepository.GetByID: %w", apperror.ErrNotFound)
		}
		return entity.User{}, fmt.Errorf("UserRepository.GetByID: query row: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	query := psql.Select("user_id", "username", "email").
		From("users").
		Where(squirrel.Eq{"username": username})

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepository.GetByUsername: to sql: %w", err)
	}

	var user entity.User
	if err = r.db.Pool.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Username, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("UserRepository.GetByUsername: %w", apperror.ErrNotFound)
		}
		return entity.User{}, fmt.Errorf("UserRepository.GetByUsername: query row: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	query := psql.Select("user_id", "username", "email").
		From("users").
		Where(squirrel.Eq{"email": email})

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepository.GetByEmail: to sql: %w", err)
	}

	var user entity.User
	if err = r.db.Pool.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Username, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, fmt.Errorf("UserRepository.GetByEmail: %w", apperror.ErrNotFound)
		}
		return entity.User{}, fmt.Errorf("UserRepository.GetByEmail: query row: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	query := psql.Update("users").
		Set("email", email).
		Where(squirrel.Eq{"user_id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UserRepository.UpdateEmail: to sql: %w", err)
	}

	if _, err = r.db.Pool.Exec(ctx, sql, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
			return fmt.Errorf("UserRepository.UpdateEmail: %w", apperror.ErrEmailTaken)
		}
		return fmt.Errorf("UserRepository.UpdateEmail: exec: %w", err)
	}

	return nil
}
