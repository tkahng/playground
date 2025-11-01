package core

import (
	"testing"

	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func ExtractTestMailer(t testing.TB, testApi App) *mailer.TestMailer {
	var testMailer *mailer.TestMailer
	if m, ok := testApi.Mailer().(*mailer.TestMailer); ok {
		testMailer = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return testMailer
}
func ExtractTestPaymentClient(t testing.TB, app App) *services.MockPaymentClient {
	var paymenClient *services.MockPaymentClient
	if m, ok := app.PaymentClient().(*services.MockPaymentClient); ok {
		paymenClient = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return paymenClient
}

func ExtractAdapterDecorator(t testing.TB, app App) *stores.StorageAdapterDecorator {
	var adapter *stores.StorageAdapterDecorator
	if m, ok := app.Adapter().(*stores.StorageAdapterDecorator); ok {
		adapter = m
	} else {
		t.Fatal("mailer is not a TestMailer")
	}
	return adapter
}
