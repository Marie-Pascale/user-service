# Tests de charge

## Pourquoi un conteneur dedie plutot qu'une installation locale ?

[Vegeta](https://github.com/tsenart/vegeta) est un outil de test de charge HTTP ecrit en Go. Plutot que de l'installer localement (`go install`), on le fait tourner dans un conteneur Docker construit specifiquement pour ca (`loadtest/Dockerfile`), avec une **version epinglee**. Deux avantages :

- **Pas de probleme de versionnage** : le conteneur encapsule sa propre version de Go et de vegeta, independamment de ce qui est installe (ou pas) sur ta machine.
- **Portable vers un pipeline CI** : `docker compose run` fonctionne de la meme facon sur ta machine et sur un runner CI -- pas besoin d'installer quoi que ce soit sur l'environnement d'execution des tests automatises.

## Comment ca fonctionne

Le service `loadtest` dans `docker-compose.yml` a un `profiles: ["tools"]` -- ca l'exclut de `docker compose up` (et donc de `make up`). Il ne demarre jamais comme service de fond, seulement a la demande via `docker compose run --rm`, comme un outil en ligne de commande ponctuel.

La cible pointe vers `http://api:8080/health` (pas `localhost`) parce que la requete part maintenant *depuis l'interieur* du reseau Docker du projet (`usagers-net`) -- `api` est le nom du service, resolu automatiquement par le DNS interne de Docker Compose.

**Pourquoi un fichier de resultats plutot qu'un tube (`|`) entre deux conteneurs ?** C'est la premiere approche qu'on avait essayee, mais relier deux `docker compose run` par un tube shell s'est revele fragile : les messages de statut de Docker Compose (construction de l'image, creation du conteneur) peuvent se glisser dans le flux binaire attendu par `report`, le corrompant (`encode: can't detect encoding of "stdin"`). Ecrire dans un fichier intermediaire, sur le volume monte, evite completement ce probleme -- et c'est aussi plus proche de ce qu'on ferait dans un pipeline CI (garder le fichier de resultats comme artefact).

## Utilisation

```bash
make load-test-health
```

Ce qui execute, en coulisse :

```bash
docker compose run --rm -T loadtest attack -targets=/loadtest/health.http -rate=50 -duration=30s -output=/loadtest/results.bin
docker compose run --rm -T loadtest report /loadtest/results.bin
```

Le fichier `results.bin` atterrit directement dans `loadtest/` sur ton disque (grace au volume monte) -- il est ignore par Git (voir `.gitignore`), donc pas besoin de le nettoyer manuellement entre deux executions ; chaque run l'ecrase simplement.

## Cibles disponibles

| Fichier | Endpoint testé |
|---|---|
| `health.http` | `GET /health` (via le nom de service `api`, dans le reseau Docker) |

## Commandes vegeta avancees

Le fichier `results.bin` genere par `make load-test-health` peut etre reutilise pour d'autres formats de sortie, sans relancer l'attaque :

```bash
docker compose run --rm -T loadtest plot /loadtest/results.bin > loadtest/plot.html
```

## Prochaines étapes

Au fur et a mesure qu'on ajoute des endpoints critiques (ex: `GET /users`,
`POST /users`), on ajoutera un fichier `.http` correspondant ici (en
pointant vers `api:8080`, pas `localhost`), avec une cible dans le
`Makefile`. Pour les endpoints qui necessitent une authentification,
vegeta supporte l'ajout d'en-tetes personnalises via un fichier de cibles
au format JSON -- utile pour simuler des requetes avec `Authorization:
Bearer ...` ou `X-API-Key: ...`.
