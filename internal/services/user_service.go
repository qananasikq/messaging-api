package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"messaging-api/internal/models"
	"messaging-api/internal/repositories"
	jwtpkg "messaging-api/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repositories.UserRepo
	jwt  *jwtpkg.JWT
}

func NewUserService(repo *repositories.UserRepo, jwt *jwtpkg.JWT) *UserService {
	return &UserService{repo: repo, jwt: jwt}
}

type AuthResult struct {
	User  models.User
	Token string
}

func (s *UserService) Register(ctx context.Context, username, password string) (AuthResult, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return AuthResult{}, ErrValidation
	}
	if len(password) < 8 || len(password) > 128 {
		return AuthResult{}, ErrValidation
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("bcrypt: %w", err)
	}

	u := models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		if err == repositories.ErrConflict {
			return AuthResult{}, ErrConflict
		}
		return AuthResult{}, err
	}

	token, _, err := s.jwt.IssueAccessToken(created.ID, created.Username)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token: %w", err)
	}

	return AuthResult{User: created, Token: token}, nil
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (AuthResult, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return AuthResult{}, ErrValidation
	}
	if len(password) < 8 || len(password) > 128 {
		return AuthResult{}, ErrValidation
	}

	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if err == repositories.ErrNotFound {
			return AuthResult{}, ErrUnauthorized
		}
		return AuthResult{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, ErrUnauthorized
	}

	token, _, err := s.jwt.IssueAccessToken(u.ID, u.Username)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token: %w", err)
	}

	u.PasswordHash = ""
	return AuthResult{User: u, Token: token}, nil
}
