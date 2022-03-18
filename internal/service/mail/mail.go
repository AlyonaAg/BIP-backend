package mail

import (
	"bytes"
	"crypto/tls"
	"errors"
	"html/template"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Sender struct {
	config *Config
}

type tmplData struct {
	Code string
}

func NewSender() (*Sender, error) {
	config, err := newConfig()
	if err != nil {
		return nil, err
	}

	return &Sender{
		config: config,
	}, nil
}

func (s *Sender) SendMail(toMail string, code string) error {
	server, err := s.serverSetup()
	if err != nil {
		return err
	}
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email, err := s.createEmail(toMail, code)
	if err != nil {
		return err
	}

	if err := email.Send(smtpClient); err != nil {
		return err
	}
	return nil
}

func (s *Sender) serverSetup() (*mail.SMTPServer, error) {
	configSender, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	server := mail.NewSMTPClient()
	server.Host = configSender.Host
	server.Port = int(configSender.Port)
	server.Username = configSender.Mail
	server.Password = configSender.Password
	server.Encryption = mail.EncryptionSTARTTLS
	server.KeepAlive = false
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return server, nil
}

func (s *Sender) createEmail(toMail string, code string) (*mail.Email, error) {
	configSender, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	email := mail.NewMSG()
	email.SetFrom(configSender.Mail).
		AddTo(toMail).
		SetSubject("Your code")

	var textHTML bytes.Buffer
	tpl, err := template.ParseFiles(configSender.PathTemplate)
	if err != nil {
		return nil, err
	}

	data := tmplData{
		Code: code,
	}
	if err := tpl.Execute(&textHTML, &data); err != nil {
		return nil, err
	}

	email.SetBody(mail.TextHTML, textHTML.String())
	if email.Error != nil {
		return nil, err
	}

	return email, nil
}

func (s *Sender) GetConfig() (*Config, error) {
	if s.config == nil {
		return nil, errors.New("empty config")
	}
	return s.config, nil
}
