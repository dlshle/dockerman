package smtpx

import "testing"

func TestSendMail(t *testing.T) {
	mail := &Mail{
		Body:    "另一封测试，别回复~O!",
		From:    "yes-reply@dashdoor.com",
		Rcpts:   []string{"dlshle@hotmail.com"},
		Subject: "Test Email",
	}

	err := SendMail(mail)
	if err != nil {
		t.Error(err)
	}
}
