package smtpx

import "testing"

func TestSendMail(t *testing.T) {
	mail := &Mail{
		Body:    "hello world again!",
		From:    "dlshle@mytestlocaldomain.com",
		Rcpts:   []string{"dlshle@hotmail.com"},
		Subject: "another test bro",
	}

	err := SendMail(mail)
	if err != nil {
		t.Error(err)
	}
}
