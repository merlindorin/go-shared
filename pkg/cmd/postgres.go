package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Host            string        `env:"POSTGRES_HOST" help:"PostgreSQL host" default:"localhost"`
	Port            int           `env:"POSTGRES_PORT" help:"PostgreSQL port" default:"5432"`
	User            string        `env:"POSTGRES_USER" help:"PostgreSQL user" default:"postgres"`
	Password        string        `env:"POSTGRES_PASSWORD" help:"PostgreSQL password"`
	Database        string        `env:"POSTGRES_DATABASE" help:"PostgreSQL database"`
	SSLMode         string        `env:"POSTGRES_SSL_MODE" help:"PostgreSQL SSL mode" default:"disable"`
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS" help:"Max number of connections" default:"10"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" help:"Min number of connections" default:"2"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" help:"Max lifetime of a connection" default:"1h"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" help:"Max idle time of a connection" default:"30m"`
}

func (p *Postgres) DSN() string {
	database := p.Database

	if database == "" {
		database = "demo"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, database, p.SSLMode)

	if p.Password != "" {
		dsn += fmt.Sprintf(" password=%s", p.Password)
	}

	return dsn
}

func (p *Postgres) Pool(ctx context.Context, options ...PoolOption) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(p.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	config.MaxConns = p.MaxConns
	config.MinConns = p.MinConns
	config.MaxConnLifetime = p.MaxConnLifetime
	config.MaxConnIdleTime = p.MaxConnIdleTime

	for _, option := range options {
		option.apply(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	return pool, nil
}

type PoolOption func(*pgxpool.Config)

func (poolOption PoolOption) apply(config *pgxpool.Config) {
	poolOption(config)
}

func WithBeforeConnect(fn func(context.Context, *pgx.ConnConfig) error) PoolOption {
	return func(config *pgxpool.Config) {
		config.BeforeConnect = fn
	}
}

func WithAfterConnect(fn func(context.Context, *pgx.Conn) error) PoolOption {
	return func(config *pgxpool.Config) {
		config.AfterConnect = fn
	}
}

func WithPrepareConn(fn func(context.Context, *pgx.Conn) (bool, error)) PoolOption {
	return func(config *pgxpool.Config) {
		config.PrepareConn = fn
	}
}

func WithAfterRelease(fn func(*pgx.Conn) bool) PoolOption {
	return func(config *pgxpool.Config) {
		config.AfterRelease = fn
	}
}

func WithBeforeClose(fn func(*pgx.Conn)) PoolOption {
	return func(config *pgxpool.Config) {
		config.BeforeClose = fn
	}
}

func WithShouldPing(fn func(context.Context, pgxpool.ShouldPingParams) bool) PoolOption {
	return func(config *pgxpool.Config) {
		config.ShouldPing = fn
	}
}

func WithMaxConnLifetime(d time.Duration) PoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnLifetime = d
	}
}

func WithMaxConnLifetimeJitter(d time.Duration) PoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnLifetimeJitter = d
	}
}

func WithMaxConnIdleTime(d time.Duration) PoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnIdleTime = d
	}
}

func WithMaxConns(n int32) PoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConns = n
	}
}

func WithMinConns(n int32) PoolOption {
	return func(config *pgxpool.Config) {
		config.MinConns = n
	}
}

func WithMinIdleConns(n int32) PoolOption {
	return func(config *pgxpool.Config) {
		config.MinIdleConns = n
	}
}

func WithHealthCheckPeriod(d time.Duration) PoolOption {
	return func(config *pgxpool.Config) {
		config.HealthCheckPeriod = d
	}
}
