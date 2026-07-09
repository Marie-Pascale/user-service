// Package repository contient les implementations Postgres des interfaces
// de repository definies dans internal/application.
package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config contient les parametres de connexion a Postgres.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN construit la chaine de connexion au format attendu par pgx.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// ConfigFromEnv lit la configuration depuis les variables d'environnement
// definies dans docker-compose.yml pour le service api. Des valeurs par
// defaut raisonnables sont utilisees si une variable est absente (utile
// pour lancer des tests hors conteneur, contre un Postgres local).
func ConfigFromEnv() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "postgres"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewPool cree un pool de connexions Postgres et verifie qu'il est
// joignable (Ping) avant de le retourner -- on prefere echouer
// immediatement au demarrage plutot que de decouvrir un probleme
// de connexion a la premiere requete.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("creation du pool de connexions: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping de la base de donnees: %w", err)
	}

	return pool, nil
}
