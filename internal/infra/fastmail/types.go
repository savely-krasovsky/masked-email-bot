package fastmail

import (
	"encoding/json"
	"errors"
	"time"
)

type MaskedEmail struct {
	ID            string           `json:"id,omitempty"`
	Email         string           `json:"email,omitempty"`
	State         MaskedEmailState `json:"state,omitempty"`
	ForDomain     string           `json:"forDomain,omitempty"`
	Description   string           `json:"description,omitempty"`
	LastMessageAt *time.Time       `json:"lastMessageAt,omitempty"`
	CreatedAt     string           `json:"createdAt,omitempty"`
	CreatedBy     string           `json:"createdBy,omitempty"`
	URL           *string          `json:"url,omitempty"`
	EmailPrefix   string           `json:"emailPrefix,omitempty"`
}

type Request[T any] struct {
	Using       []string         `json:"using"`
	MethodCalls []*Invocation[T] `json:"methodCalls"`
}

type MaskedEmailSetRequest struct {
	AccountID string                  `json:"accountId"`
	Create    map[string]*MaskedEmail `json:"create,omitempty"`
	Update    map[string]*MaskedEmail `json:"update,omitempty"`
	Destroy   []string                `json:"destroy,omitempty"`
}

type MaskedEmailState string

const (
	MaskedEmailStatePending  MaskedEmailState = "pending"
	MaskedEmailStateEnabled  MaskedEmailState = "enabled"
	MaskedEmailStateDisabled MaskedEmailState = "disabled"
	MaskedEmailStateDeleted  MaskedEmailState = "deleted"
)

type Response[T any] struct {
	MethodResponses []*Invocation[T] `json:"methodResponses"`
	SessionState    string           `json:"sessionState"`
}

type MaskedEmailSetResponse struct {
	Created   map[string]*MaskedEmail `json:"created"`
	Updated   map[string]*MaskedEmail `json:"updated"`
	Destroyed []string                `json:"destroyed"`
}

type Invocation[T any] struct {
	Name string
	Body T
	ID   string
}

func (m *Invocation[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{m.Name, m.Body, m.ID})
}

func (m *Invocation[T]) UnmarshalJSON(b []byte) error {
	var a []json.RawMessage
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	if len(a) != 3 {
		return errors.New("json: cannot parse invocation")
	}

	if err := json.Unmarshal(a[0], &m.Name); err != nil {
		return err
	}
	if err := json.Unmarshal(a[1], &m.Body); err != nil {
		return err
	}
	if err := json.Unmarshal(a[2], &m.ID); err != nil {
		return err
	}

	return nil
}
