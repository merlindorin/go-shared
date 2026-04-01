package cmd

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	Path            string        `env:"SQLITE_PATH" help:"SQLite database file path" default:":memory:"`
	JournalMode     string        `env:"SQLITE_JOURNAL_MODE" help:"SQLite journal mode" default:"wal"`
	BusyTimeout     int           `env:"SQLITE_BUSY_TIMEOUT" help:"Busy timeout in milliseconds" default:"5000"`
	ForeignKeys     bool          `env:"SQLITE_FOREIGN_KEYS" help:"Enable foreign keys" default:"true"`
	Synchronous     string        `env:"SQLITE_SYNCHRONOUS" help:"Synchronous mode" default:"normal"`
	CacheSize       int           `env:"SQLITE_CACHE_SIZE" help:"Cache size in pages (negative for KiB)" default:"-2000"`
	MaxOpenConns    int           `env:"SQLITE_MAX_OPEN_CONNS" help:"Max number of open connections" default:"10"`
	MaxIdleConns    int           `env:"SQLITE_MAX_IDLE_CONNS" help:"Max number of idle connections" default:"5"`
	ConnMaxLifetime time.Duration `env:"SQLITE_CONN_MAX_LIFETIME" help:"Max lifetime of a connection" default:"1h"`
	ConnMaxIdleTime time.Duration `env:"SQLITE_CONN_MAX_IDLE_TIME" help:"Max idle time of a connection" default:"30m"`
}

func (s *SQLite) DSN() string {
	fk := 0
	if s.ForeignKeys {
		fk = 1
	}

	path := s.Path
	extra := ""
	if path == ":memory:" {
		path = ":memory:"
		extra = "&mode=memory&cache=shared"
	}

	return fmt.Sprintf(
		"file:%s?_pragma=journal_mode(%s)&_pragma=busy_timeout(%d)&_pragma=foreign_keys(%d)&_pragma=synchronous(%s)&_pragma=cache_size(%d)%s",
		path, s.JournalMode, s.BusyTimeout, fk, s.Synchronous, s.CacheSize, extra,
	)
}

type DBOption func(*sql.DB)

func (dbOption DBOption) apply(db *sql.DB) {
	dbOption(db)
}

func WithMaxOpenConns(n int) DBOption {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}

func WithMaxIdleConns(n int) DBOption {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithConnMaxLifetime(d time.Duration) DBOption {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

func WithConnMaxIdleTime(d time.Duration) DBOption {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}

func (s *SQLite) Open(options ...DBOption) (*sql.DB, error) {
	db, err := sql.Open("sqlite", s.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(s.MaxOpenConns)
	db.SetMaxIdleConns(s.MaxIdleConns)
	db.SetConnMaxLifetime(s.ConnMaxLifetime)
	db.SetConnMaxIdleTime(s.ConnMaxIdleTime)

	for _, option := range options {
		option.apply(db)
	}

	return db, nil
}
