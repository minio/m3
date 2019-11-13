// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"os"
)

// SendMail sends an email to `toName <toEmail>` with the provided subject and body.
// This function depends on `MAIL ACCOUNT`, `MAIL_SERVER` and `MAIL_PASSWORD` environment variables being set.
func SendMail(toName, toEmail, subject, body string) error {
	// Sender data.
	account := os.Getenv("MAIL_ACCOUNT")
	if account == "" {
		return errors.New("No mailing account configured")
	}
	// Connect to the SMTP Server
	servername := os.Getenv("MAIL_SERVER")
	if servername == "" {
		return errors.New("mail server is not set")
	}
	password := os.Getenv("MAIL_PASSWORD")
	from := mail.Address{Name: "mkube team", Address: account}
	to := mail.Address{Name: toName, Address: toEmail}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject

	// Message.
	//message := []byte("This is a really unimaginative message, I know.")
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += mime
	message += "\r\n" + body

	host, _, _ := net.SplitHostPort(servername)

	// Authentication.
	auth := smtp.PlainAuth("", account, password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}

// GetTemplate gets a template from the templates folder and applies the template date
func GetTemplate(templateFileName string, data interface{}) (*string, error) {
	// Working Directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	t, err := template.ParseFiles(wd + fmt.Sprintf("/cluster/templates/%s.html", templateFileName))
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return nil, err
	}
	body := buf.String()
	return &body, nil
}
