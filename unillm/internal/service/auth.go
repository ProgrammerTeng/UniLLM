package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	infrajwt "github.com/unillm/unillm/infra/jwt"
	"github.com/unillm/unillm/infra/persistence"
	"github.com/unillm/unillm/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *persistence.UserRepo
	keyRepo    *persistence.APIKeyRepo
	jwtIssuer  *infrajwt.Issuer
	bcryptCost int
}

func NewAuthService(userRepo *persistence.UserRepo, keyRepo *persistence.APIKeyRepo, jwtIssuer *infrajwt.Issuer, bcryptCost int) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		keyRepo:    keyRepo,
		jwtIssuer:  jwtIssuer,
		bcryptCost: bcryptCost,
	}
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(input.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Role:         "user",
		Balance:      1.0,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.jwtIssuer.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.jwtIssuer.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) CreateAPIKey(userID int64, name, scope string) (string, *model.APIKey, error) {
	raw := generateAPIKey()
	hash := infrajwt.HashAPIKey(raw)

	key := &model.APIKey{
		UserID:    userID,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: raw[:12],
		Scope:     scope,
		IsActive:  true,
	}
	if err := s.keyRepo.Create(key); err != nil {
		return "", nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return raw, key, nil
}

func (s *AuthService) ListAPIKeys(userID int64) ([]model.APIKey, error) {
	return s.keyRepo.ListByUserID(userID)
}

func (s *AuthService) DeleteAPIKey(id, userID int64) error {
	return s.keyRepo.Deactivate(id, userID)
}

func (s *AuthService) GetUser(id int64) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *AuthService) ResolveAPIKey(keyHash string) (userID int64, keyID int64, ok bool) {
	key, err := s.keyRepo.FindByHash(keyHash)
	if err != nil {
		return 0, 0, false
	}
	go func() { _ = s.keyRepo.UpdateLastUsed(key.ID) }()
	return key.UserID, key.ID, true
}

func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.userRepo.UpdatePassword(userID, string(hash))
}

func generateAPIKey() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed")
	}
	return "sk-" + hex.EncodeToString(b)
}
