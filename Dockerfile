FROM golang:1.25-alpine

WORKDIR /app

# Installe air pour le hot-reload en developpement.
# Version epinglee (plutot que @latest) pour eviter qu'une future version
# d'air exige une version de Go plus recente que l'image utilisee ici,
# ce qui casserait le build sans avertissement -- comme ca vient d'arriver.
RUN go install github.com/air-verse/air@v1.65.3

# Installe le CLI golang-migrate, compile uniquement avec le driver
# postgres (via un build tag) pour reduire le temps de build.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copie les fichiers de dependances d'abord (meilleur cache Docker)
COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
