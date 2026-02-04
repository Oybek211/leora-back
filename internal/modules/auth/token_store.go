package auth

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

type refreshRecord struct {
	userID    string
	expiresAt time.Time
}

type TokenStore interface {
	SaveRefreshToken(token string, record refreshRecord)
	ConsumeRefreshToken(token string) (string, error)
	RevokeRefreshTokensForUser(userID string)
	BlacklistAccessToken(token string, expiresAt time.Time)
	IsAccessTokenBlacklisted(token string) bool
}

// InMemoryTokenStore keeps signed tokens in memory for simulation.
type InMemoryTokenStore struct {
	mu           sync.RWMutex
	refreshStore map[string]refreshRecord
	blacklist    map[string]time.Time
}

func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		refreshStore: make(map[string]refreshRecord),
		blacklist:    make(map[string]time.Time),
	}
}

func (s *InMemoryTokenStore) SaveRefreshToken(token string, record refreshRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshStore[token] = record
}

func (s *InMemoryTokenStore) ConsumeRefreshToken(token string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.refreshStore[token]
	if !ok {
		return "", ErrRefreshTokenNotFound
	}
	if time.Now().After(record.expiresAt) {
		delete(s.refreshStore, token)
		return "", ErrRefreshTokenExpired
	}
	delete(s.refreshStore, token)
	return record.userID, nil
}

func (s *InMemoryTokenStore) RevokeRefreshTokensForUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, record := range s.refreshStore {
		if record.userID == userID {
			delete(s.refreshStore, token)
		}
	}
}

func (s *InMemoryTokenStore) BlacklistAccessToken(token string, expiresAt time.Time) {
	if token == "" {
		return
	}
	tail := token
	if len(tail) > 8 {
		tail = tail[len(tail)-8:]
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blacklist[token] = expiresAt
	log.Printf("[tokenStore] BlacklistAccessToken: added ...%s (total blacklisted: %d)", tail, len(s.blacklist))
}

func (s *InMemoryTokenStore) IsAccessTokenBlacklisted(token string) bool {
	if token == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	tail := token
	if len(tail) > 8 {
		tail = tail[len(tail)-8:]
	}

	if expiry, ok := s.blacklist[token]; ok {
		if time.Now().After(expiry) {
			delete(s.blacklist, token)
			return false
		}
		log.Printf("[tokenStore] token ...%s IS in blacklist (expires %v)", tail, expiry)
		return true
	}

	// Log all blacklisted tokens for debugging
	for blToken := range s.blacklist {
		blTail := blToken
		if len(blTail) > 8 {
			blTail = blTail[len(blTail)-8:]
		}
		log.Printf("[tokenStore] blacklist entry: ...%s (checked ...%s, match=%v)", blTail, tail, blToken == token)
	}

	return false
}
