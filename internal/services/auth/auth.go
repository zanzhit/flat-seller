package authservice

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/zanzhit/flat-seller/internal/domain/constants"
	"github.com/zanzhit/flat-seller/internal/domain/errs"
	"github.com/zanzhit/flat-seller/internal/domain/models"
	jwtmid "github.com/zanzhit/flat-seller/internal/lib/jwt"
	"github.com/zanzhit/flat-seller/internal/lib/logger/sl"
)

type AuthService struct {
	secret       string
	tokenTTL     time.Duration
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
}

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, tokenTTL time.Duration, secret string) *AuthService {
	return &AuthService{
		secret:       secret,
		tokenTTL:     tokenTTL,
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
	}
}

type UserSaver interface {
	SaveUser(email, userType string, passHash []byte) (string, error)
}

type UserProvider interface {
	User(userID string) (models.User, error)
}

func (s *AuthService) RegisterNewUser(email, password, userType string) (string, error) {
	const op = "service.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	if userType != constants.User && userType != constants.Admin {
		s.log.Warn("invalid user_type", sl.Err(errs.ErrUserType))
		return "", fmt.Errorf("%s: %w", op, errs.ErrUserType)
	}

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.userSaver.SaveUser(email, userType, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *AuthService) Login(userID, password string) (string, error) {
	const op = "service.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("userID", userID),
	)

	log.Info("attempting to login user")

	user, err := s.userProvider.User(userID)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidCredentials) {
			s.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		}

		s.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		s.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
	}

	log.Info("user logged in successfully")

	token, err := jwtmid.NewToken(user, s.tokenTTL, s.secret)
	if err != nil {
		s.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *AuthService) GenerateToken(userID, email, userType string) (string, error) {
	const op = "service.auth.GenerateToken"

	log := s.log.With(
		slog.String("op", op),
		slog.String("userID", userID),
		slog.String("email", email),
		slog.String("userType", userType),
	)

	claims := jwt.MapClaims{
		"uid":       userID,
		"email":     email,
		"user_type": userType,
		"exp":       time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}
