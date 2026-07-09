package application

import "errors"

// ErrNotFound est retournee par les repositories quand la ressource
// demandee n'existe pas. Elle vit ici (pas dans le domaine, pas dans
// l'infrastructure) parce qu'elle decrit un resultat de requete, pas
// une regle metier -- toutes les implementations de repository
// (Postgres aujourd'hui, autre chose demain) doivent la retourner
// de la meme facon, pour que le code appelant n'ait pas a connaitre
// les details de l'implementation (ex: pgx.ErrNoRows).
var ErrNotFound = errors.New("resource not found")
