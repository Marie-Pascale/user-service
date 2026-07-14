FROM golang:1.25-alpine

# git est requis par "go install" pour resoudre certaines dependances via
# le controle de version -- absent par defaut des images alpine, contrairement
# aux images golang standards (non-alpine).
RUN apk add --no-cache git

# Version epinglee (pas @latest) pour un environnement reproductible --
# evite les problemes de versionnage rencontres avec une installation
# locale, et donne une base stable pour integrer ces tests a un pipeline
# CI plus tard.
RUN go install github.com/tsenart/vegeta/v12@v12.12.0

ENTRYPOINT ["vegeta"]
