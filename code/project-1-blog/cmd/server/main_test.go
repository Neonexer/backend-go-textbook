package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, получен %d", http.StatusOK, rec.Code)
	}

	expected := "GET запрос принят\n"
	if rec.Body.String() != expected {
		t.Errorf("ожидалось %q, получено %q", expected, rec.Body.String())
	}
}

func TestHandler_POST(t *testing.T) {
	body := strings.NewReader(`{"message": "hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, получен %d", http.StatusOK, rec.Code)
	}
}

func TestHandler_PUT_NotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("ожидался статус %d, получен %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
