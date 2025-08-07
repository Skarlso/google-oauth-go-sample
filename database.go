package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

// Database Encapsulates a connection to a database.
type Database struct {
}

// SaveUser register a user so we know that we saw that user already.
func (d *Database) SaveUser(u *User) error {
	db, err := d.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	defer db.Close()

	// We could do `LoadUser` here, but we want to avoid the marshalling.
	if err := db.View(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(u.Email)); err != nil {
			return err
		}

		return nil
	}); err == nil {
		return fmt.Errorf("user already exists")
	}

	if err := db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(u)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}

		if err := txn.Set([]byte(u.Email), data); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// LoadUser get data from a user.
func (d *Database) LoadUser(Email string) (*User, error) {
	db, err := d.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	defer db.Close()

	result := &User{}
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(Email))
		if err != nil {
			return err
		}

		if err := item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, result); err != nil {
				return fmt.Errorf("failed to unmarshal value into user: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to deal with the value: %w", err)
		}
		return nil
	})

	return result, err
}

// GetDB return a new session if there is no previous one.
func (d *Database) GetDB() (*badger.DB, error) {
	dir := filepath.Join(os.TempDir(), "google-oauth-sample", "badger")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = &badgerLogger{logger: slog.Default()}
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

type badgerLogger struct {
	logger *slog.Logger
}

func (l *badgerLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *badgerLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}
