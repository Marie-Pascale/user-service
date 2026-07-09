// Package security fournit les implementations concretes des contrats
// de securite definis dans le domaine (voir domain.PasswordHasher).
package security

import "golang.org/x/crypto/bcrypt"

// BcryptHasher implemente domain.PasswordHasher avec bcrypt.
//
// Note : ce type n'importe pas le package domain, il implemente juste
// la meme interface implicitement (c'est ainsi que fonctionne le
// duck-typing des interfaces en Go -- aucune declaration explicite
// d'implementation n'est necessaire).
type BcryptHasher struct {
	// Cost controle le nombre d'iterations de l'algorithme. Plus il est
	// eleve, plus le hachage est lent (donc plus resistant au brute-force,
	// mais plus couteux en CPU). bcrypt.DefaultCost (10) est un bon point
	// de depart.
	Cost int
}

// NewBcryptHasher cree un BcryptHasher avec le cout par defaut recommande.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{Cost: bcrypt.DefaultCost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
