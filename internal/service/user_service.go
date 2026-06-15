package service

import (
	"errors"
	"notification-api/internal/domain"
	"notification-api/internal/repository"
	"notification-api/internal/validation"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(email, password string) (*domain.User, error) {
	if err := validation.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return nil, err
	}

	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *UserService) GetByID(userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *UserService) GetAll() ([]domain.User, error) {
	return s.userRepo.GetAll()
}
