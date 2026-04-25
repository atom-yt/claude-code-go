package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "atom_ai",
		SSLMode:         "disable",
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}

// Database represents the database connection pool
type Database struct {
	pool *pgxpool.Pool
}

// New creates a new database connection pool
func New(ctx context.Context, cfg *Config) (*Database, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{pool: pool}, nil
}

// NewFromURL creates a database connection from a connection URL
func NewFromURL(ctx context.Context, url string) (*Database, error) {
	pool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{pool: pool}, nil
}

// Pool returns the underlying connection pool
func (d *Database) Pool() *pgxpool.Pool {
	return d.pool
}

// Close closes the database connection pool
func (d *Database) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}

// Health checks if the database is healthy
func (d *Database) Health(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (d *Database) Stats() *pgxpool.Stat {
	return d.pool.Stat()
}

// Ping pings the database
func Ping(ctx context.Context, conn *pgx.Conn) error {
	return conn.Ping(ctx)
}