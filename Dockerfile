FROM golang:1.23-alpine

WORKDIR /app

# Installe air pour le hot-reload en developpement
RUN go install github.com/air-verse/air@latest

# Copie les fichiers de dependances d'abord (meilleur cache Docker)
COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]