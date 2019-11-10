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
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

// smtpServer data to smtp server
type smtpServer struct {
	host string
	port string
}

// serverName URI to smtp server
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

func SendMail(toName, toEmail, subject, textBody, htmlBody string) error {
	// Sender data.
	account := os.Getenv("MAIL_ACCOUNT")
	from := fmt.Sprintf("From: %s <%s>", "mkube team", account)
	password := os.Getenv("MAIL_PASSWORD")
	// Receiver email address.
	to := []string{
		toEmail,
	}
	// smtp server configuration.
	smtpServer := smtpServer{host: "smtp.gmail.com", port: "587"}
	// Message.
	//message := []byte("This is a really unimaginative message, I know.")
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject = fmt.Sprintf("Subject: %s\n", subject)

	templateData := struct {
		Name string
		URL  string
	}{
		Name: "Daniel",
		URL:  "http://min.io",
	}

	body, err := getTemplate("invite", templateData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	message := []byte(subject + mime + *body)
	// Authentication.
	auth := smtp.PlainAuth("", account, password, smtpServer.host)
	// Sending email.
	err = smtp.SendMail(smtpServer.Address(), auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Email Sent!")
	return nil
}

func getTemplate(templateFileName string, data interface{}) (*string, error) {
	// Working Directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	t, err := template.ParseFiles(wd + "/cluster/templates/invite.html")
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
