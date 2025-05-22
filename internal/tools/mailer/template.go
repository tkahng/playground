package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strings"
)

type SendMailParams struct {
	Subject      string
	Type         string
	TemplatePath string
	Template     string
}
type EmailParams struct {
	Token      string
	Type       string
	RedirectTo string
}

type CommonParams struct {
	SiteURL         string `json:"site_url"`
	ConfirmationURL string `json:"confirmation_url"`
	Email           string `json:"email"`
	Token           string `json:"token"`
	TokenHash       string `json:"token_hash"`
	RedirectTo      string `json:"redirect_to"`
}

type AllEmailParams struct {
	*SendMailParams
	*CommonParams
	*Message
}

func GenerateConfirmationURL(base string, path string, token, tokenType, redirectTo string) (string, error) {
	parsedURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parsedURL.Path = path
	params, err := GetPathParams(parsedURL.String(), token, tokenType, redirectTo)
	if err != nil {
		return "", err
	}
	return parsedURL.ResolveReference(params).String(), nil
}

func GetPathParams(filepath string, token, tokenType, redirectTo string) (*url.URL, error) {
	path := &url.URL{}
	if filepath != "" {
		if p, err := url.Parse(filepath); err != nil {
			return nil, err
		} else {
			path = p
		}
	}
	path.RawQuery = fmt.Sprintf("token=%s&type=%s&redirect_to=%s", url.QueryEscape(token), url.QueryEscape(tokenType), encodeRedirectURL(redirectTo))
	return path, nil
}

func encodeRedirectURL(referrerURL string) string {
	if len(referrerURL) > 0 {
		if strings.ContainsAny(referrerURL, "&=#") {
			// if the string contains &, = or # it has not been URL
			// encoded by the caller, which means it should be URL
			// encoded by us otherwise, it should be taken as-is
			referrerURL = url.QueryEscape(referrerURL)
		}
	}
	return referrerURL
}
func GetTemplate[T any](name string, mailTemplate string, params T) string {
	tmpl, err := template.New(name).Parse(mailTemplate)
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

const DefaultTeamInviteMail = `<h2>You have been invited</h2>
<p>You have been invited to joint team {{ .TeamName }} by {{ .InvitedByEmail }}. Follow this link to accept the invite:</p>
<p><a href="{{ .ConfirmationURL }}">Accept the invite</a></p>`

const DefaultConfirmationMail = `<h2>Confirm your email</h2>

<p>Follow this link to confirm your email:</p>
<p><a href="{{ .ConfirmationURL }}">Confirm your email address</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>
`

const DefaultSecurityPasswordResetMail = `<h2>Your password has been reset due to security concerns</h2>
<p>We noticed that you signed in with a social provider while you were already signed in with an unverified email/password account.</p>
<p>For your security, we have reset your password to a temporary password.</p>
<p>If you wish to sign in with your email/password account, please reset your password by clicking the link below:</p>
<p><a href="{{ .ConfirmationURL }}">Reset password</a></p>
<p>Alternatively, enter the code: {{ .Token }}</p>`

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
