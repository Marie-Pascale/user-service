package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON ecrit une reponse JSON avec le status donne. Centralise
// l'entete Content-Type pour ne pas l'oublier dans chaque handler.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// decodeJSON decode le corps d'une requete JSON dans dst.
// DisallowUnknownFields rejette les champs inattendus plutot que de les
// ignorer silencieusement -- ca attrape les fautes de frappe du client
// (ex: "emial" au lieu de "email") au lieu de les laisser passer inapercues.
func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
