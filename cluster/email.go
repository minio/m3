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
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	"github.com/minio/m3/cluster/db"

	"github.com/minio/minio/pkg/env"
)

// SendMail sends an email to `toName <toEmail>` with the provided subject and body.
// This function depends on `MAIL_ACCOUNT`, `MAIL_SERVER` and `MAIL_PASSWORD` environment variables being set.
func SendMail(toName, toEmail, subject, body string) error {
	// Sender data.

	account := env.Get(mailAccount, "")
	if account == "" {
		return errors.New("No mailing account configured")
	}
	// Connect to the SMTP Server
	servername := env.Get(mailServer, "")
	if servername == "" {
		return errors.New("mail server is not set")
	}
	password := env.Get(mailPassword, "")
	fromName := env.Get(mailFromName, "mkube team")
	from := mail.Address{Name: fromName, Address: account}
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
func GetTemplate(templateName string, data interface{}) (*string, error) {
	// validate that the template is only alpha numerical stuff
	var re = regexp.MustCompile(`^[a-z0-9-]{2,}$`)
	if !re.MatchString(templateName) {
		return nil, errors.New("invalid template name")
	}
	// try to load the template from the db
	dbTemplate, err := getTemplateFromDB(nil, templateName)
	if err != nil {
		// ignore no results error
		if !strings.Contains(err.Error(), "no rows in result set") {
			return nil, err
		}
	}
	var t *template.Template
	if dbTemplate != nil && *dbTemplate != "" {
		// parse the template from the DB
		t, err = template.New(templateName).Parse(*dbTemplate)
		if err != nil {
			return nil, err
		}
	} else {
		// if we found no template, load from disk
		// Working Directory
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		t, err = template.ParseFiles(wd + fmt.Sprintf("/cluster/templates/%s.html", templateName))
		if err != nil {
			return nil, err
		}
	}
	// replace the {{.Tokens}} from the template with the provided struct data
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return nil, err
	}
	body := buf.String()
	return &body, nil
}

func getTemplateFromDB(ctx *Context, templateName string) (*string, error) {
	query :=
		`SELECT 
				et.template
			FROM 
				email_templates et
			WHERE et.name=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return nil, err
		}
		row = tx.QueryRow(query, templateName)
	} else {
		// no context? straight to db
		row = db.GetInstance().Db.QueryRow(query, templateName)
	}

	// Save the resulted query on the User struct
	var template string
	err := row.Scan(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// SetEmailTemplate upserts a template into the database. If the id is not present the record will be inserted, if it's
// present it will be updated
func SetEmailTemplate(ctx *Context, templateName, templateBody string) error {
	// validate that the template is only alpha numerical stuff
	var re = regexp.MustCompile(`^[a-z0-9-]{2,}$`)
	if !re.MatchString(templateName) {
		return errors.New("invalid template name")
	}
	// Insert or Update template
	query := `INSERT INTO 
					email_templates (name, template) 
				VALUES ($1, $2) 
				ON CONFLICT (name) DO 
			    UPDATE SET template=$2`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, templateName, templateBody)
	if err != nil {
		return err
	}
	return nil
}
