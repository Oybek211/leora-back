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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blacklist[token] = expiresAt
	log.Printf("[tokenStore] BLACKLIST ADD len=%d tail=...%s total=%d", len(token), last8(token), len(s.blacklist))
}

func (s *InMemoryTokenStore) IsAccessTokenBlacklisted(token string) bool {
	if token == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, ok := s.blacklist[token]
	if !ok {
		log.Printf("[tokenStore] CHECK len=%d tail=...%s → NOT in blacklist (size=%d)", len(token), last8(token), len(s.blacklist))
		for bl := range s.blacklist {
			log.Printf("[tokenStore]   entry: len=%d tail=...%s same=%v", len(bl), last8(bl), bl == token)
		}
		return false
	}
	if time.Now().After(expiry) {
		delete(s.blacklist, token)
		return false
	}
	log.Printf("[tokenStore] CHECK len=%d tail=...%s → BLACKLISTED", len(token), last8(token))
	return true
}

func last8(s string) string {
	if len(s) > 8 {
		return s[len(s)-8:]
	}
	return s
}
