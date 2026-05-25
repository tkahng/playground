package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"testing"

	apphttp "github.com/tkahng/playground/internal/tools/http"
)

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("test panic")
}

func TestRecovererMiddleware(t *testing.T) {
	mw := RecovererMiddleware()

	h1 := mw(http.HandlerFunc(panicHandler))
	req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()
	h1.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	var errModel apphttp.ErrorModel
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&errModel); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code to be 500 got %d", res.StatusCode)
	}
	if errModel.Title != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("expected title to be %s got %s", http.StatusText(http.StatusInternalServerError), errModel.Title)
	}
	if errModel.Detail != "internal server error" {
		t.Errorf("expected detail to be %s got %s", "internal server error", errModel.Detail)
	}
}
