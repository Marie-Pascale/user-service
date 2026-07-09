package domain

import (
	"crypto/rand"
	"fmt"
)

// NewID genere un identifiant unique au format UUID v4, en utilisant
// uniquement la bibliotheque standard (crypto/rand) -- pas de dependance
// vers github.com/google/uuid. Suffisant pour ce projet ; a remplacer par
// une librairie dediee si tu as besoin d'UUIDs v7 (ordonnes temporellement,
// plus performants pour les index B-tree sur de gros volumes).
func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read ne peut echouer que si /dev/urandom est inaccessible --
		// une situation anormale sur laquelle on ne peut pas raisonnablement
		// continuer.
		panic(err)
	}

	// Version 4, variant RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
