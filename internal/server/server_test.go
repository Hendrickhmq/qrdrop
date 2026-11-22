package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qrdrop/internal/progress"
)

const token = "segredo"

func newTestServer(t *testing.T) (*httptest.Server, string, *progress.Hub) {
	t.Helper()
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	hub := progress.NewHub()
	go hub.Run(ctx)
	ts := httptest.NewServer((&Server{Dir: dir, Token: token, Hub: hub}).Handler())
	t.Cleanup(ts.Close)
	return ts, dir, hub
}

func upload(t *testing.T, url, name string, data []byte, hash string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", name)
	part.Write(data)
	mw.Close()
	req, _ := http.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Upload-ID", "teste")
	if hash != "" {
		req.Header.Set("X-SHA256", hash)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func TestTokenObrigatorio(t *testing.T) {
	ts, _, _ := newTestServer(t)
	for _, path := range []string{"/", "/?t=errado", "/events"} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("GET %s: status %d, queria 401", path, resp.StatusCode)
		}
	}
	resp, _ := http.Get(ts.URL + "/?t=" + token)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("com token: status %d", resp.StatusCode)
	}
}

func TestUploadSalvaEConfereHash(t *testing.T) {
	ts, dir, _ := newTestServer(t)
	data := bytes.Repeat([]byte("qrdrop "), 100_000) // ~700 KB
	resp := upload(t, ts.URL+"/upload?t="+token, "../fora.txt", data, sum(data))
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, b)
	}
	var res UploadResult
	json.NewDecoder(resp.Body).Decode(&res)
	if res.Name != "fora.txt" || res.SHA256 != sum(data) || res.Size != int64(len(data)) {
		t.Errorf("resposta inesperada: %+v", res)
	}
	got, err := os.ReadFile(filepath.Join(dir, "fora.txt"))
	if err != nil || !bytes.Equal(got, data) {
		t.Errorf("arquivo salvo diferente do enviado (err=%v)", err)
	}
}

func TestUploadHashErradoApagaArquivo(t *testing.T) {
	ts, dir, _ := newTestServer(t)
	resp := upload(t, ts.URL+"/upload?t="+token, "a.txt", []byte("conteudo"), strings.Repeat("0", 64))
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, queria 422", resp.StatusCode)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("arquivo corrompido deveria ter sido apagado, sobrou %v", entries)
	}
}

func TestEventosSSE(t *testing.T) {
	ts, _, _ := newTestServer(t)
	resp, err := http.Get(ts.URL + "/events?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	lines := bufio.NewScanner(resp.Body)
	lines.Scan() // ": conectado"

	upload(t, ts.URL+"/upload?t="+token, "b.txt", []byte("oi"), "")

	deadline := time.After(3 * time.Second)
	found := make(chan progress.Event)
	go func() {
		for lines.Scan() {
			if data, ok := strings.CutPrefix(lines.Text(), "data: "); ok {
				var ev progress.Event
				json.Unmarshal([]byte(data), &ev)
				if ev.Done {
					found <- ev
					return
				}
			}
		}
	}()
	select {
	case ev := <-found:
		if ev.ID != "teste" || ev.Name != "b.txt" || ev.Bytes != 2 || ev.SHA256 != sum([]byte("oi")) {
			t.Errorf("evento inesperado: %+v", ev)
		}
	case <-deadline:
		t.Fatal("nenhum evento de conclusão chegou pelo SSE")
	}
}
