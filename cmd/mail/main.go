package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/url"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

func main() {
	// username := "your.address@googlemail.com"
	// password := "<app password>"
	// client, err := mail.NewClient("smtp.gmail.com", mail.WithTLSPortPolicy(mail.TLSMandatory),
	// 	mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(username), mail.WithPassword(password))
	// if err != nil {
	// 	fmt.Printf("failed to create mail client: %s\n", err)
	// 	os.Exit(1)
	// }

	// message := mail.NewMsg()
	// // Your message-specific code here
	// if err = client.DialAndSend(message); err != nil {
	// 	fmt.Printf("failed to send mail: %s\n", err)
	// 	os.Exit(1)
	// }

	ctx := context.Background()
	cfg := conf.AppConfigGetter()
	app := core.InitBaseApp(ctx, cfg)
	db := app.Db()

	opts := app.Settings().Auth
	// client := app.NewMailClient()
	user, _ := factory.New().NewUser(factory.UserMods.RandomEmail(nil)).Create(ctx, db)
	otp := security.GenerateOtp(6)
	token := security.GenerateTokenKey()
	payload := &core.OtpPayload{
		UserId:     user.ID,
		Email:      user.Email,
		Token:      token,
		Otp:        otp,
		RedirectTo: "",
	}
	_, _ = core.CreateOtpToken(payload, opts.VerificationToken)
	path, _ := mailer.GetPath("/api/auth/verify", &mailer.EmailParams{
		Token:      payload.Token,
		Type:       string(shared.VerificationTokenType),
		RedirectTo: payload.RedirectTo,
	})
	appUrl, _ := url.Parse(app.Settings().Meta.AppURL)
	param := &mailer.CommonParams{
		SiteURL:         appUrl.String(),
		ConfirmationURL: appUrl.ResolveReference(path).String(),
		Email:           user.Email,
		Token:           otp,
		TokenHash:       payload.Token,
		RedirectTo:      "",
	}

	tmpl, err := template.New("body").Parse(mailer.DefaultConfirmationMail)
	if err != nil {
		log.Fatal(err)
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, param)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body.String())

	// message := mail.NewMsg()
	// fmt.Println(tokenHash)
	// fmt.Println(param)

}
