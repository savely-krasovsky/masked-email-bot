package domain

type Service interface {
	StartCommand(telegramID int) error
	TokenCommand(telegramID int, token string) error
	Link(telegramID int, forDomain string) (string, error)
}

type service struct {
	db    Database
	email MaskingEmail
}

func NewService(db Database, email MaskingEmail) Service {
	return &service{
		db:    db,
		email: email,
	}
}

func (s *service) StartCommand(telegramID int) error {
	return s.db.CreateUser(telegramID)
}

func (s *service) TokenCommand(telegramID int, token string) error {
	return s.db.UpdateToken(telegramID, token)
}

func (s *service) Link(telegramID int, forDomain string) (string, error) {
	user, err := s.db.GetUser(telegramID)
	if err != nil {
		return "", err
	}

	if user.FastmailToken == nil {
		return "", ErrNoToken
	}

	return s.email.CreateMaskedEmail(*user.FastmailToken, forDomain)
}
