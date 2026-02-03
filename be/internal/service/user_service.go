package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
)

type UserService interface {
	CreateUser(ctx context.Context, name string, email string, password string) (sqlc.User, error)
	AuthenticateUser(ctx context.Context, email string, password string) (sqlc.User, error)
	GetUser(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(
	ctx context.Context,
	name string,
	email string,
	password string,
) (sqlc.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return sqlc.User{}, errors.New("email is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return sqlc.User{}, err
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		Name:        pgtype.Text{String: name, Valid: true},
		Email:       email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, ErrUserAlreadyExists
		}
		// Fallback for when errors.As fails (e.g. if error is wrapped differently)
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key value") {
			return sqlc.User{}, ErrUserAlreadyExists
		}
		return sqlc.User{}, err
	}

	return user, nil
}

func (s *userService) AuthenticateUser(ctx context.Context, email string, password string) (sqlc.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc.User{}, ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return sqlc.User{}, ErrInvalidPassword
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return sqlc.User{}, ErrUserNotFound
	}
	return user, nil
}
