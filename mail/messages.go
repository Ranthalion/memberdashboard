package mail

import (
	"bytes"
	"errors"
	"html/template"
	"memberserver/config"
	"memberserver/database"
	"time"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

//TODO: [ML] Redesign to throttle emails and only expose methods through a management struct?
//maybe a mailer struct

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
	db     *database.Database
	m      MailApi
	config config.Config
}

type MailApi interface {
	SendHtmlMail(address, subject, body string) (string, error)
	SendPlainTextMail(address, subject, content string) (string, error)
}

func NewMailer(db *database.Database, m MailApi, config config.Config) *mailer {
	return &mailer{db, m, config}
}

func (m *mailer) SendCommunication(communication CommunicationTemplate, recipient string, model interface{}) (bool, error) {
	c, err := m.db.GetCommunication(communication.String())
	if err != nil {
		log.Printf("%v not found", communication.String())
		return false, err
	}

	memberExists := true
	member, err := m.db.GetMemberByEmail(recipient)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			memberExists = false
		} else {
			return false, err
		}

	}

	if memberExists && m.isThrottled(c, member) {
		log.Printf("Communication %v not sent to %v due to throttling", communication.String(), recipient)
		return false, nil
	}

	content, err := generateEmailContent("./templates/"+c.Template, model)
	if err != nil {
		log.Errorf("Error generating email content. Error: %v", err)
		return false, err
	}

	_, err = m.m.SendHtmlMail(recipient, c.Subject, content)
	if err != nil {
		return false, err
	}

	if memberExists {
		m.db.LogCommunication(c.ID, member.ID)
	}

	return true, nil
}

func (m *mailer) isThrottled(c database.Communication, member database.Member) bool {

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

func SendGracePeriodMessageToLeadership(recipient string, member interface{}) {
	infoAddress := "info@hackrva.org"
	//sendCommunication(PendingRevokationLeadership, infoAddress, member)

	SendTemplatedEmail("pending_revokation_leadership.html.tmpl", infoAddress, "hackRVA Grace Period", member)
}

func SendGracePeriodMessage(recipient string, member interface{}) {
	SendTemplatedEmail("pending_revokation_member.html.tmpl", recipient, "hackRVA Grace Period", member)
}

func SendRevokedEmail(recipient string, member interface{}) {
	SendTemplatedEmail("access_revoked.html.tmpl", recipient, "hackRVA Grace Period", member)
}

func SendRevokedEmailToLeadership(recipient string, member interface{}) {
	SendTemplatedEmail("access_revoked_leadership.html.tmpl", recipient, "hackRVA Grace Period", member)
}

func SendIPHasChanged(newIpAddress string) {
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	recipient := c.AdminEmail
	ipModel := struct {
		IpAddress string
	}{
		IpAddress: newIpAddress}
	SendTemplatedEmail("ip_changed.html.tmpl", recipient, "IP Address Changed", ipModel)
}

func generateEmailContent(templatePath string, model interface{}) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Errorf("Error loading template %v", err)
		return "", err
	}
	tmpl.Option("missingkey=error")
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, model)
	if err != nil {
		log.Errorf("Error generating content %v", err)
		return "", err
	}
	return tpl.String(), nil
}

func SendTemplatedEmail(templateName string, to string, subject string, model interface{}) {
	conf, _ := config.Load()

	if !conf.EnableInfoEmails {
		log.Info("email not enabled")
		return
	}

	mp, err := Setup()
	if err != nil {
		log.Errorf("error setting up mailprovider when attempting to send email notification")
		return
	}

	if len(conf.EmailOverrideAddress) > 0 {
		to = conf.EmailOverrideAddress
	}

	content, err := generateEmailContent("./templates/"+templateName, model)
	if err != nil {
		log.Errorf("Error generating email contnent. Error: %v", err)
		return
	}

	_, err = mp.SendHtmlMail(to, subject, content)
	if err != nil {
		log.Errorf("Error sending mail %v", err)
	}
}
