package mail

import (
	"errors"
	"memberserver/config"
	"memberserver/database"
	"time"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type CommunicationTemplate string

const (
	AccessRevokedMember         CommunicationTemplate = "AccessRevokedMember"
	AccessRevokedLeadership     CommunicationTemplate = "AccessRevokedLeadership"
	IpChanged                   CommunicationTemplate = "IpChanged"
	PendingRevokationLeadership CommunicationTemplate = "PendingRevokationLeadership"
	PendingRevokationMember     CommunicationTemplate = "PendingRevokationMember"
	Welcome                     CommunicationTemplate = "Welcome"
)

// String converts CommunicationTemplate to a string
func (c CommunicationTemplate) String() string {
	return string(c)
}

type mailer struct {
	db                                CommunicationDal
	m                                 MailApi
	config                            config.Config
	generator                         templateGenerator
	enableInfoEmails                  bool
	enableNotificationEmailsToMembers bool
	emailOverrideAddress              string
}

type MailApi interface {
	SendHtmlMail(address, subject, body string) (string, error)
	SendPlainTextMail(address, subject, content string) (string, error)
}

type CommunicationDal interface {
	GetMemberByEmail(memberEmail string) (database.Member, error)
	GetCommunication(communication string) (database.Communication, error)
	LogCommunication(communicationId int, memberId string) error
	GetMostRecentCommunicationToMember(memberId string, commId int) (time.Time, error)
}

func NewMailer(db CommunicationDal, m MailApi, config config.Config) *mailer {
	mailer := mailer{
		db,
		m,
		config,
		fileTemplateGenerator{},
		config.EnableInfoEmails,
		config.EnableNotificationEmailsToMembers,
		config.EmailOverrideAddress}
	return &mailer
}

func (m *mailer) SendCommunication(communication CommunicationTemplate, recipient string, model interface{}) (bool, error) {
	memberExists := true
	member, err := m.db.GetMemberByEmail(recipient)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			memberExists = false
		} else {
			return false, err
		}
	}

	c, err := m.db.GetCommunication(communication.String())
	if err != nil {
		log.Printf("%v not found. Err: %v", communication.String(), err)
		return false, err
	}

	if memberExists && m.IsThrottled(c, member) {
		log.Printf("Communication %v not sent to %v due to throttling", communication.String(), recipient)
		return false, nil
	}

	content, err := m.generator.generateEmailContent("./templates/"+c.Template, model)
	if err != nil {
		log.Errorf("Error generating email content. Error: %v", err)
		return false, err
	}

	if len(m.emailOverrideAddress) > 0 {
		recipient = m.emailOverrideAddress
	}

	_, err = m.m.SendHtmlMail(recipient, c.Subject, content)
	if err != nil {
		log.Printf("Failed to send mail to %v.  Err: %v", recipient, err)
		return false, err
	}

	if memberExists {
		m.db.LogCommunication(c.ID, member.ID)
	}

	return true, nil
}

func (m *mailer) IsThrottled(c database.Communication, member database.Member) bool {

	if c.FrequencyThrottle > 0 {
		last, err := m.db.GetMostRecentCommunicationToMember(member.ID, c.ID)
		if err != nil {
			return false
		}
		difference := time.Since(last).Hours() / 24
		if difference < float64(c.FrequencyThrottle) {
			return true
		}
	}
	return false
}
