FROM golang:1.25-alpine

WORKDIR /app

# Installe air pour le hot-reload en developpement.
# Version epinglee (plutot que @latest) pour eviter qu'une future version
# d'air exige une version de Go plus recente que l'image utilisee ici,
# ce qui casserait le build sans avertissement -- comme ca vient d'arriver.
RUN go install github.com/air-verse/air@v1.65.3

# Copie les fichiers de dependances d'abord (meilleur cache Docker)
COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
