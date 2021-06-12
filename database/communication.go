package database

import (
	"context"

	log "github.com/sirupsen/logrus"
)

const getCommunications string = "Select id, name, subject, frequency_throttle, template from membership.communication"
const getCommunication string = "Select id, name, subject, frequency_throttle, template from membership.communication where name = $1"

// Communication defines an email communication
type Communication struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Subject           string `json:"subject"`
	FrequencyThrottle int    `json:"frequency_throttle"`
	Template          string `json:"template"`
}

// GetCommunnications returns all communications from the database
func (db *Database) GetCommunications() []Communication {
	rows, err := db.getConn().Query(context.Background(), getCommunications)
	if err != nil {
		log.Errorf("conn.Query failed: %v", err)
	}

	defer rows.Close()

	var communications []Communication

	for rows.Next() {
		var c Communication
		err = rows.Scan(&c.ID, &c.Name, &c.Subject, &c.FrequencyThrottle, &c.Template)
		communications = append(communications, c)
	}
	return communications
}

// GetCommunnication returns all the requested communication from the database
func (db *Database) GetCommunication(name string) Communication {
	var c Communication
	err := db.getConn().QueryRow(context.Background(), getCommunication, name).
		Scan(&c.ID, &c.Name, &c.Subject, &c.FrequencyThrottle, &c.Template)
	if err != nil {
		log.Errorf("GetCommunication failed: %v", err)
	}
	return c
}

func (db *Database) GetMostRecentCommunicationToRecipient(recipient string, c Communication) string {
	return ""
}
