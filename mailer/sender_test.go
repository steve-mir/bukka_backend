package mailer

// func TestSendEmailWithSMTP(t *testing.T) {
// 	config, err := utils.LoadConfig("..")
// 	require.NoError(t, err)

// 	log.Println("Details", config.SMTPAddr, config.SMTPHost, config.SMTPUsername, config.SMTPPassword)
// 	sender := NewSMTPSender("Settle in", config.SMTPAddr, config.SMTPHost, config.SMTPUsername, config.SMTPPassword)

// 	subject := "A test email"
// 	content := `
// 	<h1> Welcome onboard </h1>
// 	<p>This is a test message from John Doe</>
// 	`
// 	to := []string{"johnnydoe16161616@gmail.com"}
// 	attachFiles := []string{"../README.md"}

// 	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
// 	require.NoError(t, err)
// }
