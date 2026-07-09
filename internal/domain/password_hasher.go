package domain

// PasswordHasher definit le contrat pour hacher et verifier des mots de
// passe. L'interface vit dans le domaine (Dependency Inversion Principle),
// mais son implementation concrete (bcrypt, argon2, ou plus tard un
// mecanisme type Vault) vit dans la couche infrastructure -- voir
// internal/infrastructure/security.
//
// Ca permet de tester la logique du domaine (AuthMethod) sans dependre
// d'une vraie bibliotheque de hachage, et de changer d'algorithme plus
// tard sans toucher au domaine.
type PasswordHasher interface {
	// Hash retourne le hash d'un mot de passe en clair.
	Hash(password string) (string, error)

	// Verify retourne true si le mot de passe en clair correspond au hash.
	Verify(password, hash string) bool
}
