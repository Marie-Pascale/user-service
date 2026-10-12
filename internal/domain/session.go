package domain

import (
	"errors"
	"time"
)

var ErrInvalidTTL = errors.New("session TTL must be positive")

type Session struct {
	ID         string
	UserID     string
	TokenHash  string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt time.Time
}

func NewSession(id, userID, ipAddress, userAgent string, ttl time.Duration) (session *Session, plainToken string, err error) {
	if userID == "" {
		return nil, "", ErrEmptyUserID
	}
	if ttl <= 0 {
		return nil, "", ErrInvalidTTL
	}

	plainToken, err = generateSecureToken(32)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	session = &Session{
		ID:         id,
		UserID:     userID,
		TokenHash:  HashToken(plainToken),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		LastUsedAt: now,
	}
	return session, plainToken, nil
}

func (s *Session) VerifyToken(plainToken string) bool {
	return HashToken(plainToken) == s.TokenHash
}

func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

func (s *Session) IsSuspicious(ipAddress, userAgent string) bool {
	return ipAddress != s.IPAddress && userAgent != s.UserAgent
}

func (s *Session) Touch() {
	s.LastUsedAt = time.Now().UTC()
}
