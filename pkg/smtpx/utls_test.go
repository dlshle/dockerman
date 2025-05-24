package smtpx

import "testing"

func TestSendMail(t *testing.T) {
	mail := &Mail{
		Body:    "不要回复，不要回复，不要回复！！",
		From:    "noreply@doordash.com",
		Rcpts:   []string{"xuritest@notifications.citystoragesystems-staging.com"},
		Subject: "Xuri Test!!!",
	}

	err := SendMail(mail)
	if err != nil {
		t.Fatal(err)
	}
}
