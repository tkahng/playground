package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
)

func testErrorRender(w http.ResponseWriter, r *http.Request) {
	WriteErr(w, r, http.StatusBadRequest, "bad request")
}

func TestError_Render(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()
	testErrorRender(w, req)
	res := w.Result()
	defer res.Body.Close()
	var errModel ErrorModel
	if err := render.DecodeJSON(res.Body, &errModel); err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code to be 400 got %d", res.StatusCode)
	}
	if errModel.Title != http.StatusText(http.StatusBadRequest) {
		t.Errorf("expected title to be %s got %s", http.StatusText(http.StatusBadRequest), errModel.Title)
	}
	if errModel.Detail != "bad request" {
		t.Errorf("expected detail to be %s got %s", "bad request", errModel.Detail)
	}
}
