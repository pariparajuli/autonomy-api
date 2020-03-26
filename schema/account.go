package schema

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ActivityState struct {
	LastActiveTime time.Time
	LastLocation   *Location
}

func (u ActivityState) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *ActivityState) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, &u)
}

type AccountMetadata map[string]interface{}

func (u AccountMetadata) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *AccountMetadata) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(source, &u)
}

type Account struct {
	AccountNumber string `gorm:"primary_key"`
	EncPubKey     string
	Profile       AccountProfile `gorm:"foreignkey:ProfileID"`
	ProfileID     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AccountProfile struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" sql:"default:uuid_generate_v4()"`
	AccountNumber string
	Version       string
	State         ActivityState   `gorm:"type:jsonb;not null;default '{}'"`
	Metadata      AccountMetadata `gorm:"type:jsonb;not null;default '{}'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
