# Tests de charge

## Pourquoi vegeta ?

[Vegeta](https://github.com/tsenart/vegeta) est un outil de test de charge HTTP écrit en Go.
On aurait aussi pu utiliser `k6` (JavaScript) ou `hey`, mais vegeta a l'avantage
de rester dans l'écosystème Go et de produire des rapports détaillés
(latence p50/p95/p99, taux de succès, débit) directement en ligne de commande.

## Installation

```bash
go install github.com/tsenart/vegeta@latest
```

## Utilisation

L'API doit être en cours d'exécution (`make up` ou `docker compose up`).

Chaque fichier `.http` dans ce dossier définit une ou plusieurs cibles au format :
```
MÉTHODE URL
```

Pour lancer un test de charge sur `/health` (50 requêtes/seconde pendant 30 secondes) :

```bash
make load-test-health
```

Ou directement :

```bash
vegeta attack -targets=loadtest/health.http -rate=50 -duration=30s | vegeta report
```

### Générer un graphique de latence

```bash
vegeta attack -targets=loadtest/health.http -rate=50 -duration=30s | vegeta encode > results.json
cat results.json | vegeta plot > plot.html
```

## Cibles disponibles

| Fichier | Endpoint testé |
|---|---|
| `health.http` | `GET /health` |

## Prochaines étapes

Au fur et à mesure qu'on ajoute des endpoints critiques (ex: `GET /users`,
`POST /users`), on ajoutera un fichier `.http` correspondant ici, avec une
cible dans le `Makefile`. Pour les endpoints qui nécessitent une
authentification, vegeta supporte l'ajout d'en-têtes personnalisés via
un fichier de cibles au format JSON — on ajustera quand on y sera rendus.
