package main

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"html/template"
	"time"
)

type StringArray []string

type Flat struct {
	ExtId        uint      `gorm:"primaryKey"`
	IsActive     bool      `gorm:"not null"`
	TotalPrice   float64   `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	PushedUpAt   sql.NullTime
	LastChecked  time.Time   `gorm:"not null"`
	Url          string      `gorm:"not null"`
	Notified     bool        `gorm:"not null;default:0"`
	Title        string      `gorm:"not null"`
	Description  string      `gorm:"not null"`
	Photos       StringArray `gorm:"type:json"`
	OriginalJson Ad          `gorm:"type:json"`
}

func (s StringArray) Value() (driver.Value, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(jsonData), nil
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported Scan type")
	}
}

func (f *Flat) ToMailBody() (string, error) {
	t, err := template.New("mail_template.tmpl").ParseFiles("mail_template.tmpl")
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, f)

	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
