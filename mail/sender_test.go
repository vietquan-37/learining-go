package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vietquan-37/simplebank/util"
)

func TestMail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)
	sender := NewEmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	`
	to := []string{"qnguyenviet67@gmail.com"}
	attachFiles := []string{"../app.env"}
	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)

}
