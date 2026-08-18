package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func srv() *Server {
	return New(Config{Addr: ":0", WebDir: ""})
}

func TestHealthEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	srv().Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestPriceEndpoint(t *testing.T) {
	body := `{"spot":100,"vol":0.2,"rate":0.05,"strike":100,"maturity":1,"steps":16,"paths":1000,"seed":42,"type":"euro-call"}`
	w := httptest.NewRecorder()
	srv().Handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/price", bytes.NewBufferString(body)))
	if w.Code != 200 {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp PriceResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Price <= 0 {
		t.Fatalf("price = %f", resp.Price)
	}
}

func TestBSEndpoint(t *testing.T) {
	body := `{"spot":100,"vol":0.2,"rate":0.05,"strike":100,"maturity":1}`
	w := httptest.NewRecorder()
	srv().Handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/bs", bytes.NewBufferString(body)))
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp BSResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.CallPrice <= 0 {
		t.Fatalf("call = %f", resp.CallPrice)
	}
}
