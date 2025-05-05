package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/shared"
)

type InMemoryAuthStorage struct {
	mu           sync.RWMutex
	users        map[uuid.UUID]*shared.User
	usersByEmail map[string]*shared.User
	userAccounts map[string]*shared.UserAccount // key: userId_provider
	tokens       map[string]*shared.Token
}

func NewInMemoryAuthStorage() *InMemoryAuthStorage {
	return &InMemoryAuthStorage{
		users:        make(map[uuid.UUID]*shared.User),
		usersByEmail: make(map[string]*shared.User),
		userAccounts: make(map[string]*shared.UserAccount),
		tokens:       make(map[string]*shared.Token),
	}
}

func (s *InMemoryAuthStorage) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return &shared.UserInfo{
		User: *user,
	}, nil
}

func (s *InMemoryAuthStorage) CreateUser(ctx context.Context, user *shared.User) (*shared.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[user.Email]; exists {
		return nil, errors.New("user already exists")
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user
	return user, nil
}

func (s *InMemoryAuthStorage) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[userId]
	if !exists {
		return errors.New("user not found")
	}
	// user.Roles = append(user.Roles, roleNames...)
	return nil
}

func (s *InMemoryAuthStorage) FindUserByEmail(ctx context.Context, email string) (*shared.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *InMemoryAuthStorage) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", userId, provider)
	account, exists := s.userAccounts[key]
	if !exists {
		return nil, errors.New("account not found")
	}
	return account, nil
}

func (s *InMemoryAuthStorage) UpdateUser(ctx context.Context, user *shared.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user
	return nil
}

func (s *InMemoryAuthStorage) UpdateUserAccount(ctx context.Context, account *shared.UserAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", account.UserID, account.Provider)
	if _, exists := s.userAccounts[key]; !exists {
		return errors.New("account not found")
	}
	s.userAccounts[key] = account
	return nil
}

func (s *InMemoryAuthStorage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(s.usersByEmail, user.Email)
	delete(s.users, id)
	return nil
}

func (s *InMemoryAuthStorage) LinkAccount(ctx context.Context, account *shared.UserAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", account.UserID, account.Provider)
	if _, exists := s.userAccounts[key]; exists {
		return errors.New("account already linked")
	}
	s.userAccounts[key] = account
	return nil
}

func (s *InMemoryAuthStorage) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s_%s", userId, provider)
	delete(s.userAccounts, key)
	return nil
}

func (s *InMemoryAuthStorage) VerifyTokenStorage(ctx context.Context, token string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.tokens[token]
	if !exists {
		return errors.New("token not found")
	}
	return nil
}

func (s *InMemoryAuthStorage) GetToken(ctx context.Context, token string) (*shared.Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.tokens[token]
	if !exists {
		return nil, errors.New("token not found")
	}
	if t.Expires.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return t, nil
}

func (s *InMemoryAuthStorage) SaveToken(ctx context.Context, tokenDTO *shared.CreateTokenDTO) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[tokenDTO.Token] = &shared.Token{
		Token:      tokenDTO.Token,
		UserID:     tokenDTO.UserID,
		Expires:    tokenDTO.Expires,
		Type:       tokenDTO.Type,
		Otp:        tokenDTO.Otp,
		Identifier: tokenDTO.Identifier,
	}
	return nil
}

func (s *InMemoryAuthStorage) DeleteToken(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, token)
	return nil
}
