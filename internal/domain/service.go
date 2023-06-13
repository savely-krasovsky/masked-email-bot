package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"io"
)

type Service interface {
	StartCommand(telegramID int64, languageCode string) (string, error)
	HandleRedirect(ctx context.Context, code, state string) error
	Link(telegramID int64, forDomain string) (string, error)
}

type service struct {
	logger   *zap.Logger
	db       Database
	email    MaskingEmail
	telegram Telegram
}

func NewService(logger *zap.Logger, db Database, email MaskingEmail, telegram Telegram) Service {
	return &service{
		logger:   logger,
		db:       db,
		email:    email,
		telegram: telegram,
	}
}

func randomBytesInHex(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("Could not generate %d random bytes: %v", count, err)
	}

	return hex.EncodeToString(buf), nil
}

func (s *service) StartCommand(telegramID int64, languageCode string) (string, error) {
	if err := s.db.CreateUser(telegramID, languageCode); err != nil {
		if !errors.Is(err, ErrSqliteUserAlreadyExists) {
			return "", err
		}

		if err := s.db.UpdateLanguageCode(telegramID, languageCode); err != nil {
			return "", err
		}
	}

	codeVerifier, err := randomBytesInHex(32)
	if err != nil {
		s.logger.Error("Error while generating random bytes!", zap.Error(err))
		return "", ErrRandom
	}

	hasher := sha256.New()
	hasher.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	state, err := randomBytesInHex(24)
	if err != nil {
		s.logger.Error("Error while generating random bytes!", zap.Error(err))
		return "", ErrRandom
	}

	authCodeURL := s.email.GetOAuth2Config().AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)

	if err := s.db.CreateOAuth2State(state, codeVerifier, telegramID); err != nil {
		return "", err
	}

	return authCodeURL, nil
}

func (s *service) HandleRedirect(ctx context.Context, code, state string) error {
	oauth2State, err := s.db.GetOAuth2State(state)
	if err != nil {
		return err
	}

	user, err := s.db.GetUser(oauth2State.TelegramID)
	if err != nil {
		return err
	}

	token, err := s.email.GetOAuth2Config().Exchange(
		ctx, code, oauth2.SetAuthURLParam("code_verifier", oauth2State.CodeVerifier),
	)
	if err != nil {
		s.logger.Error("Error while exchanging authorization code!", zap.Error(err))
		return ErrFastmailInternal
	}

	b, err := json.Marshal(token)
	if err != nil {
		s.logger.Error("Error while trying to marshall OAuth2 token!", zap.Error(err))
		return ErrJSONEncoding
	}

	if err := s.db.UpdateToken(user.TelegramID, string(b)); err != nil {
		return err
	}

	if err := s.telegram.SendMessage(user.TelegramID, user.LanguageCode, "TelegramAuthorizationComplete"); err != nil {
		return err
	}

	return nil
}

func (s *service) Link(telegramID int64, forDomain string) (string, error) {
	user, err := s.db.GetUser(telegramID)
	if err != nil {
		return "", err
	}

	if user.FastmailToken == nil {
		return "", ErrNoToken
	}

	ctx := context.Background()
	tokenSrc := s.db.NewTokenSource(
		s.email.GetOAuth2Config().TokenSource(ctx, user.FastmailToken),
		user.TelegramID,
	)

	return s.email.CreateMaskedEmail(ctx, tokenSrc, forDomain)
}
