package mail

import (
	"bytes"
	"html/template"
	"memberserver/config"
	"memberserver/database"

	log "github.com/sirupsen/logrus"
)

//TODO: [ML] Redesign to throttle emails and only expose methods through a management struct?
//maybe a mailer struct

type Communication string

const (
	AccessRevokedMember         Communication = "AccessRevokedMember"
	AccessRevokedLeadership     Communication = "AccessRevokedLeadership"
	IpChanged                   Communication = "IpChanged"
	PendingRevokationLeadership Communication = "PendingRevokationLeadership"
	PendingRevokationMember     Communication = "PendingRevokationMember"
	Welcome                     Communication = "Welcome"
)

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

func (m *mailer) SendCommunication(communication Communication, recipient string, model interface{}) (bool, error) {

	//TODO: [ML] Add subject and template path to DB configuration?
	//Load communication settings, or get them from cache
	//Check if last communication to recipient is within threshold.  If so, abort.
	//Send communication
	//Log that communication was sent
}

func SendGracePeriodMessageToLeadership(recipient string, member interface{}) {
	infoAddress := "info@hackrva.org"
	sendCommunication(PendingRevokationLeadership, infoAddress, member)

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

	_, err = mp.SendComplexMessage(to, subject, content)
	if err != nil {
		log.Errorf("Error sending mail %v", err)
	}
}
