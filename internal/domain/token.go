package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Ce fichier centralise la generation et le hachage de jetons aleatoires a
// haute entropie -- utilise a la fois par Service (cle API, voir service.go)
// et par Session (jeton de connexion usager, voir session.go). Les deux
// ont exactement le meme besoin de securite : un secret aleatoire dont on
// ne stocke jamais la version en clair, seulement son hash.
//
// SHA-256 (sans sel) est approprie ici -- contrairement a un mot de passe
// choisi par un humain (voir domain.PasswordHasher, qui utilise bcrypt),
// un jeton genere par generateSecureToken a deja 256 bits d'entropie : un
// hash lent concu pour resister au brute-force n'apporte rien de plus.

// generateSecureToken produit n octets aleatoires cryptographiquement surs,
// encodes en hexadecimal (2n caracteres).
func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken hache un jeton en clair (cle API ou jeton de session) avec
// SHA-256. Exportee parce que la couche infrastructure (middlewares HTTP)
// doit hacher le jeton recu AVANT de le chercher en base -- la recherche
// se fait toujours par hash, jamais par jeton en clair.
func HashToken(plainToken string) string {
	sum := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(sum[:])
}
