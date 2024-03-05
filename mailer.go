package main

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
)

func notify(flat *Flat, recipient mail.Address) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("MAIL_USERNAME"),
		os.Getenv("MAIL_PASSWORD"),
		os.Getenv("MAIL_HOST"),
	)

	from := mail.Address{Address: os.Getenv("MAIL_USERNAME")}
	subj := "Take a look at new advert #" + strconv.Itoa(int(flat.ExtId))
	body, _ := flat.ToMailBody()

	headers := map[string]string{
		"From":         from.String(),
		"To":           recipient.String(),
		"Subject":      subj,
		"Content-Type": "text/html; charset=UTF-8",
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	conn, err := tls.Dial(
		"tcp",
		os.Getenv("MAIL_HOST")+":465",
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         os.Getenv("MAIL_HOST"),
		},
	)

	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, os.Getenv("MAIL_HOST"))
	if err != nil {
		return err
	}

	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(recipient.Address); err != nil {
		return err
	}
	writeCloser, err := client.Data()
	if err != nil {
		return err
	}

	_, err = writeCloser.Write([]byte(message))
	if err != nil {
		return err
	}

	err = writeCloser.Close()
	if err != nil {
		return err
	}
	if err = client.Quit(); err != nil {
		return err
	}

	return nil
}
