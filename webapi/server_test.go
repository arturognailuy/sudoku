package webapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/recovery"
	"github.com/gnailuy/sudoku/solver"
)

const knownPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

func testHandler(t *testing.T, token string, origins []string) (http.Handler, recovery.Store, game.Options) {
	t.Helper()
	store := solver.NewStore()
	options := game.NewDefaultOptions(store)
	options.StrategySolverKeys = store.GetAllStrategySolverKeys()
	recoveryStore := recovery.NewStore(filepath.Join(t.TempDir(), "recovery"))
	registry, err := NewRegistry(recoveryStore, options)
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_, _ string) (game.Game, error) {
		board := core.NewEmptyBoard()
		board.FromString(knownPuzzle)
		return game.NewGame(board, options), nil
	}
	return NewHandler(NewServer(registry, factory), token, origins), recoveryStore, options
}

func request(t *testing.T, handler http.Handler, method, path, contentType, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func createTestSession(t *testing.T, handler http.Handler, headers map[string]string) Session {
	t.Helper()
	w := request(t, handler, http.MethodPost, "/api/v1/sessions", "application/json", `{"source":{"kind":"puzzle","puzzle":"`+knownPuzzle+`"}}`, headers)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}
	var session Session
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Id == "" || session.Revision != 0 || len(session.Snapshot.Values) != 9 {
		t.Fatalf("invalid session: %+v", session)
	}
	return session
}

func TestSessionLifecycleRecoveryAndTransfer(t *testing.T) {
	handler, recoveryStore, options := testHandler(t, "", nil)
	session := createTestSession(t, handler, nil)
	actionPath := "/api/v1/sessions/" + session.Id + "/actions"
	w := request(t, handler, http.MethodPost, actionPath, "application/json", `{"kind":"set-value","expected_revision":0,"row":1,"column":1,"value":4}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("action status=%d body=%s", w.Code, w.Body.String())
	}
	var result ActionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Revision != 1 || result.Result.Action != "set-value" {
		t.Fatalf("unexpected action result: %+v", result)
	}
	w = request(t, handler, http.MethodPost, actionPath, "application/json", `{"kind":"clear-value","expected_revision":0,"row":1,"column":1}`, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("stale revision status=%d body=%s", w.Code, w.Body.String())
	}

	exportPath := "/api/v1/sessions/" + session.Id + "/export"
	w = request(t, handler, http.MethodGet, exportPath, "", "", nil)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != SessionMediaType {
		t.Fatalf("export status=%d type=%s", w.Code, w.Header().Get("Content-Type"))
	}
	exported := append([]byte(nil), w.Body.Bytes()...)
	w = request(t, handler, http.MethodPost, "/api/v1/sessions/import", SessionMediaType, string(exported), nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("import status=%d body=%s", w.Code, w.Body.String())
	}

	recovered, err := NewRegistry(recoveryStore, options)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(recovered, func(string, string) (game.Game, error) { panic("not called") })
	w = request(t, NewHandler(server, "", nil), http.MethodGet, "/api/v1/sessions/"+session.Id, "", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("recover status=%d body=%s", w.Code, w.Body.String())
	}
	var got Session
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Revision != 1 {
		t.Fatalf("unexpected recovered session: %+v", got)
	}

	w = request(t, handler, http.MethodDelete, "/api/v1/sessions/"+session.Id, "", "", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d", w.Code)
	}
	w = request(t, handler, http.MethodGet, "/api/v1/sessions/"+session.Id, "", "", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted get status=%d", w.Code)
	}
}

func TestSecurityCORSAndRequestValidation(t *testing.T) {
	handler, _, _ := testHandler(t, "secret", []string{"https://client.example"})
	if w := request(t, handler, http.MethodGet, "/healthz", "", "", nil); w.Code != http.StatusOK {
		t.Fatalf("health status=%d", w.Code)
	}
	if w := request(t, handler, http.MethodGet, "/api/v1/sessions", "", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("auth status=%d", w.Code)
	}
	auth := map[string]string{"Authorization": "Bearer secret"}
	badOrigin := map[string]string{"Authorization": "Bearer secret", "Origin": "https://evil.example"}
	if w := request(t, handler, http.MethodGet, "/api/v1/sessions", "", "", badOrigin); w.Code != http.StatusForbidden {
		t.Fatalf("origin status=%d", w.Code)
	}
	preflight := map[string]string{"Origin": "https://client.example", "Access-Control-Request-Method": "POST", "Access-Control-Request-Headers": "authorization, content-type"}
	if w := request(t, handler, http.MethodOptions, "/api/v1/sessions", "", "", preflight); w.Code != http.StatusNoContent {
		t.Fatalf("preflight status=%d", w.Code)
	}
	if w := request(t, handler, http.MethodPost, "/api/v1/sessions", "text/plain", `{}`, auth); w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("media status=%d", w.Code)
	}
	if w := request(t, handler, http.MethodPost, "/api/v1/sessions", "application/json", `{"source":{"kind":"puzzle","puzzle":"x","extra":1}}`, auth); w.Code != http.StatusBadRequest {
		t.Fatalf("strict status=%d body=%s", w.Code, w.Body.String())
	}
	large := strings.Repeat("x", MaxRequestBytes+1)
	if w := request(t, handler, http.MethodPost, "/api/v1/sessions", "application/json", large, auth); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("large status=%d", w.Code)
	}
}

func TestConcurrentRevisionConflict(t *testing.T) {
	handler, _, _ := testHandler(t, "", nil)
	session := createTestSession(t, handler, nil)
	path := "/api/v1/sessions/" + session.Id + "/actions"
	bodies := []string{`{"kind":"set-value","expected_revision":0,"row":1,"column":1,"value":4}`, `{"kind":"set-value","expected_revision":0,"row":1,"column":2,"value":8}`}
	codes := make(chan int, 2)
	var wg sync.WaitGroup
	for _, body := range bodies {
		wg.Add(1)
		go func(body string) {
			defer wg.Done()
			codes <- request(t, handler, http.MethodPost, path, "application/json", body, nil).Code
		}(body)
	}
	wg.Wait()
	close(codes)
	counts := map[int]int{}
	for code := range codes {
		counts[code]++
	}
	if counts[http.StatusOK] != 1 || counts[http.StatusConflict] != 1 {
		t.Fatalf("statuses: %v", counts)
	}
}

func TestImportRejectsOversizedPayload(t *testing.T) {
	handler, _, _ := testHandler(t, "", nil)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/import", io.NopCloser(bytes.NewReader(make([]byte, MaxRequestBytes+1))))
	r.Header.Set("Content-Type", SessionMediaType)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d", w.Code)
	}
}
