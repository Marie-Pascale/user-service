package domain

import (
	"errors"
	"time"
)

var ErrInvalidTTL = errors.New("session TTL must be positive")

// Session represente une connexion active d'un usager. Contrairement a un
// JWT (signe, sans etat, non revocable avant expiration), une Session est
// un jeton OPAQUE stocke en base -- choix fait volontairement pour pouvoir
// revoquer une session immediatement (deconnexion, compromission suspectee)
// sans attendre son expiration naturelle.
//
// Comme pour Service.APIKeyHash, le jeton en clair n'est JAMAIS stocke --
// seul son hash SHA-256 (voir token.go, partage avec Service).
//
// IPAddress et UserAgent sont captures UNE SEULE FOIS, a la creation --
// jamais mis a jour par la suite (voir IsSuspicious). C'est deliberement
// une "ancre" fixe plutot qu'une fenetre glissante : si on mettait a jour
// ces champs a chaque requete reussie, un attaquant qui reussit UNE
// requete avec un fingerprint different "blanchirait" son propre
// fingerprint comme nouvelle reference, et c'est l'USAGER LEGITIME qui se
// mettrait a paraitre suspect par la suite. Ancrer sur la creation evite
// ce probleme.
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

// NewSession cree une nouvelle session pour un usager, valide pour la
// duree ttl. Le jeton en clair (plainToken) n'est retourne qu'UNE SEULE
// FOIS, a la creation -- exactement comme domain.NewService pour les cles
// API. C'est ce jeton que l'usager doit presenter (via l'entete
// Authorization) pour prouver qu'il est bien connecte.
//
// ipAddress et userAgent peuvent etre vides (ex: client sans entete
// User-Agent) -- ce n'est pas une erreur de validation, juste une donnee
// moins riche pour IsSuspicious.
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

// VerifyToken retourne true si le jeton en clair fourni correspond a
// cette session.
func (s *Session) VerifyToken(plainToken string) bool {
	return HashToken(plainToken) == s.TokenHash
}

// IsExpired indique si la session a depasse sa duree de vie.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// IsSuspicious retourne true SEULEMENT si l'IP ET le User-Agent de la
// requete courante different TOUS LES DEUX de ceux enregistres a la
// creation de la session. Un changement d'un seul des deux (usager mobile
// qui change de reseau, navigateur qui se met a jour) reste tolere --
// pour eviter de deconnecter des usagers legitimes sur un simple
// changement de reseau.
//
// LIMITE CONNUE, ASSUMEE : le User-Agent est un simple entete HTTP,
// trivialement falsifiable. Un attaquant qui a capture le jeton ET le
// User-Agent exact de la victime (souvent visible dans les memes logs qui
// auraient fuite le jeton) peut recopier ce User-Agent et passer sous ce
// controle, meme depuis une IP differente. Ce controle attrape le vol
// "naif" (un jeton simplement rejoue depuis un autre appareil, sans
// falsification des entetes), pas un attaquant cible et sophistique.
// C'est un compromis deliberement choisi entre securite et faux positifs
// -- voir le README pour la discussion complete.
func (s *Session) IsSuspicious(ipAddress, userAgent string) bool {
	return ipAddress != s.IPAddress && userAgent != s.UserAgent
}

// Touch met a jour l'horodatage de derniere utilisation. N'etend PAS
// ExpiresAt (pas de "sliding session" pour l'instant) -- une session
// expire toujours a duree fixe depuis sa creation, ce qui reste simple a
// raisonner. Ca pourra evoluer si le besoin se presente.
func (s *Session) Touch() {
	s.LastUsedAt = time.Now().UTC()
}
