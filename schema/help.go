package schema

import (
	"time"

	"github.com/google/uuid"
)

const (
	HELP_PENDING   = "PENDING"
	HELP_RESPONDED = "RESPONDED"
)

type HelpRequest struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key" sql:"default:uuid_generate_v4()"`
	Requester string    `json:"requester"`
	Helper    string    `json:"helper"`
	Subject   string    `json:"subject"`
	Text      string    `json:"text"`
	State     string    `json:"state" sql:"default:'PENDING'"`
	CreatedAt time.Time `json:"created_at"`
}
