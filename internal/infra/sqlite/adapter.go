package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/L11R/masked-email-bot/internal/domain"
	"github.com/golang-migrate/migrate/v4"
	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/mattn/go-sqlite3"
	// file driver for the golang-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"log"
)

type adapter struct {
	logger *zap.Logger
	config *Config
	db     *sql.DB
}

func NewAdapter(logger *zap.Logger, config *Config) (domain.Database, error) {
	db, err := sql.Open("sqlite3", config.DBFile)
	if err != nil {
		log.Fatal(err)
	}

	// Migrations block
	driver, err := sqlite3migrate.WithInstance(db, &sqlite3migrate.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(config.MigrationsSourceURL, config.Name, driver)
	if err != nil {
		return nil, err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return &adapter{
		logger: logger,
		config: config,
		db:     db,
	}, nil
}

func (a *adapter) CreateUser(telegramID int64) error {
	_, err := a.db.Exec(
		`INSERT INTO users (telegram_id) VALUES (?)`,
		telegramID,
	)

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == 1555 {
		a.logger.Info("Duplicate key value violation!", zap.Error(err))
	} else if err != nil {
		a.logger.Error("Error while creating a user!", zap.Error(err))
		return domain.ErrSqliteInternal
	}

	return nil
}

func (a *adapter) UpdateToken(telegramID int64, fastmailToken string) error {
	_, err := a.db.Exec(
		`UPDATE users SET fastmail_token = ? WHERE telegram_id = ?`,
		fastmailToken,
		telegramID,
	)
	if err != nil {
		a.logger.Error("Error while updating a token!", zap.Error(err))
		return domain.ErrSqliteInternal
	}

	return nil
}

func (a *adapter) GetUser(telegramID int64) (*domain.User, error) {
	row := a.db.QueryRow(
		`SELECT telegram_id, fastmail_token FROM users WHERE telegram_id = ?`,
		telegramID,
	)

	var user domain.User
	var tokenStr string
	if err := row.Scan(
		&user.TelegramID,
		&tokenStr,
	); err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return nil, domain.ErrNoUser
		}

		a.logger.Error("Error while getting a user!", zap.Error(err))
		return nil, domain.ErrSqliteInternal
	}

	if err := json.Unmarshal([]byte(tokenStr), &user.FastmailToken); err != nil {
		a.logger.Error("Error while decoding a Fastmail token!", zap.Error(err))
		return nil, domain.ErrSqliteInternal
	}

	return &user, nil
}

func (a *adapter) CreateOAuth2State(state, codeVerifier string, telegramID int64) error {
	_, err := a.db.Exec(
		`INSERT INTO oauth2_states (state, code_verifier, telegram_id) VALUES (?, ?, ?)`,
		state,
		codeVerifier,
		telegramID,
	)
	if err != nil {
		a.logger.Error("Error while creating an OAuth2 state!", zap.Error(err))
		return domain.ErrSqliteInternal
	}

	return nil
}

func (a *adapter) GetOAuth2State(state string) (*domain.OAuth2State, error) {
	row := a.db.QueryRow(
		`SELECT code_verifier, telegram_id FROM oauth2_states WHERE state = ?`,
		state,
	)

	var oauth2State domain.OAuth2State
	oauth2State.State = state
	if err := row.Scan(
		&oauth2State.CodeVerifier,
		&oauth2State.TelegramID,
	); err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return nil, domain.ErrNoState
		}

		a.logger.Error("Error while getting an OAuth2 state!", zap.Error(err))
		return nil, domain.ErrSqliteInternal
	}

	return &oauth2State, nil
}

func (a *adapter) Close() error {
	return a.db.Close()
}
