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
	LastActiveTime time.Time `json:"last_active_time"`
	LastLocation   *Location `json:"location"`
}

func (u ActivityState) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *ActivityState) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}
	return json.Unmarshal(source, u)
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
	AccountNumber string         `json:"account_number" gorm:"primary_key"`
	EncPubKey     string         `json:"enc_pub_key"`
	Profile       AccountProfile `json:"profile" gorm:"foreignkey:ProfileID"`
	ProfileID     uuid.UUID      `json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type AccountProfile struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key" sql:"default:uuid_generate_v4()"`
	AccountNumber string          `json:"account_number"`
	Version       string          `json:"-"`
	State         ActivityState   `json:"state" gorm:"type:jsonb;not null;default '{}'"`
	Metadata      AccountMetadata `json:"metadata" gorm:"type:jsonb;not null;default '{}'"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
