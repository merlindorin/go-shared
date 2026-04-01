package cmd

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLite provides CLI flags for configuring a SQLite database connection.
//
// Usage:
//
//	type CLI struct {
//	    cmd.Commons
//	    cmd.SQLite `embed:"" prefix:""`
//	    Run RunCmd `cmd:"" default:"withargs" help:"Start the application."`
//	}
//
//	type RunCmd struct{}
//
//	func (r *RunCmd) Run(sq *cmd.SQLite) error {
//	    db, err := sq.Open()
//	    if err != nil {
//	        return err
//	    }
//	    defer db.Close()
//	    // ...
//	    return nil
//	}
//
//	func main() {
//	    var cli CLI
//	    ctx := kong.Parse(&cli, kong.DefaultEnvars("MYAPP"))
//	    _ = ctx.Run()
//	}
type SQLite struct {
	Path            string        `name:"sqlite-path" help:"SQLite database file path" default:":memory:"`
	JournalMode     string        `name:"sqlite-journal-mode" help:"SQLite journal mode" default:"wal"`
	BusyTimeout     int           `name:"sqlite-busy-timeout" help:"Busy timeout in milliseconds" default:"5000"`
	ForeignKeys     bool          `name:"sqlite-foreign-key" help:"Enable foreign keys" default:"true"`
	Synchronous     string        `name:"sqlite-synchronous" help:"Synchronous mode" default:"normal"`
	CacheSize       int           `name:"sqlite-cache-size" help:"Cache size in pages (negative for KiB)" default:"-2000"`
	MaxOpenConns    int           `name:"sqlite-max-open-conns" help:"Max number of open connections" default:"10"`
	MaxIdleConns    int           `name:"sqlite-max-idle-conns" help:"Max number of idle connections" default:"5"`
	ConnMaxLifetime time.Duration `name:"sqlite-conn-max-lifetime" help:"Max lifetime of a connection" default:"1h"`
	ConnMaxIdleTime time.Duration `name:"sqlite-conn-max-idle-time" help:"Max idle time of a connection" default:"30m"`
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
