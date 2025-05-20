package services

import (
	"context"
	"github.com/MukizuL/diploma-1/internal/errs"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser Creates a user with given login and password. Returns a JWT and an error.
func (s *Services) CreateUser(ctx context.Context, login, password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errs.ErrInternalServerError
	}

	userID, err := s.storage.CreateNewUser(ctx, login, string(passwordHash))
	if err != nil {
		return "", err
	}

	accessTokenSigned, err := s.CreateToken(userID)
	if err != nil {
		return "", err
	}

	return accessTokenSigned, nil
}

// LoginUser Logs in a user with given login and password. Returns a JWT and an error.
func (s *Services) LoginUser(ctx context.Context, login, password string) (string, error) {
	user, passwordHash, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return "", errs.ErrNotAuthorized
	}

	accessTokenSigned, err := s.CreateToken(user.ID)
	if err != nil {
		return "", err
	}

	return accessTokenSigned, nil
}
