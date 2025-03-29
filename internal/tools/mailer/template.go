package mailer

import (
	"bytes"
	"html/template"
	"log"
)

func GetTemplate(name string, mailTemplate string, params *CommonParams) string {
	tmpl, err := template.New("body").Parse(mailTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, params)
	if err != nil {
		log.Fatal(err)
	}
	return body.String()
}

const DefaultInviteMail = `<h2>You have been invited</h2>
<p>You have been invited to create a user on {{ .SiteURL }}. Follow this link to accept the invite:</p>
<p><a href="{{ .ConfirmationURL }}">Accept the invite</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

const DefaultConfirmationMail = `<h2>Confirm your email</h2>

<p>Follow this link to confirm your email:</p>
<p><a href="{{ .ConfirmationURL }}">Confirm your email address</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>
`

const DefaultRecoveryMail = `<h2>Reset password</h2>

<p>Follow this link to reset the password for your user:</p>
<p><a href="{{ .ConfirmationURL }}">Reset password</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

const DefaultultMagicLinkMail = `<h2>Magic Link</h2>

<p>Follow this link to login:</p>
<p><a href="{{ .ConfirmationURL }}">Log In</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

const DefaultultEmailChangeMail = `<h2>Confirm email address change</h2>

<p>Follow this link to confirm the update of your email address from {{ .Email }} to {{ .NewEmail }}:</p>
<p><a href="{{ .ConfirmationURL }}">Change email address</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

const DefaultultReauthenticateMail = `<h2>Confirm reauthentication</h2>

<p>Enter the code: {{ .Token }}</p>`
