FROM golang:1.25-alpine

WORKDIR /app

# gcc + musl-dev : requis par cgo, lui-meme requis par "go test -race"
# (le detecteur de race conditions s'appuie sur une bibliotheque C).
# Absent par defaut des images alpine (contrairement aux images golang
# standards, non-alpine) -- voir make race dans le Makefile, qui active
# cgo explicitement pour cette seule commande.
RUN apk add --no-cache gcc musl-dev

# Installe air pour le hot-reload en developpement.
# Version epinglee (plutot que @latest) pour eviter qu'une future version
# d'air exige une version de Go plus recente que l'image utilisee ici,
# ce qui casserait le build sans avertissement -- comme ca vient d'arriver.
RUN go install github.com/air-verse/air@v1.65.3

# Installe le CLI golang-migrate, compile uniquement avec le driver
# postgres (via un build tag) pour reduire le temps de build.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copie les fichiers de dependances d'abord (meilleur cache Docker).
# Le "go.sum*" (avec l'etoile) est volontaire : ca copie go.sum s'il
# existe, sans faire echouer le build s'il n'existe pas encore (ex: tout
# premier clone du projet, avant le premier "make mod-tidy").
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
