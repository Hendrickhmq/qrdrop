package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qrdrop/internal/progress"
	"qrdrop/internal/server"
)

// Teste de ponta a ponta: um receptor de verdade (httptest) e o Sender
// mandando 10 arquivos com 4 goroutines.
func TestSendAllEmParalelo(t *testing.T) {
	recvDir, sendDir := t.TempDir(), t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hub := progress.NewHub()
	go hub.Run(ctx)
	ts := httptest.NewServer(server.New(server.Config{Dir: recvDir, Token: "tk", Hub: hub}).Handler())
	defer ts.Close()

	var files []string
	for i := 0; i < 10; i++ {
		p := filepath.Join(sendDir, fmt.Sprintf("arq%d.bin", i))
		os.WriteFile(p, bytes.Repeat([]byte{byte(i)}, 50_000*(i+1)), 0o644)
		files = append(files, p)
	}
	files = append(files, filepath.Join(sendDir, "nao-existe.txt"))

	base, err := ParseURL(ts.URL + "/?t=tk")
	if err != nil {
		t.Fatal(err)
	}
	prog := make(chan Progress, 1024)
	go func() {
		for range prog { // só esvazia o channel
		}
	}()
	s := &Sender{BaseURL: base, Parallel: 4, Progress: prog}
	results := s.SendAll(ctx, files)
	close(prog)

	for i, r := range results[:10] {
		if r.Err != nil {
			t.Errorf("%s: %v", r.File, r.Err)
			continue
		}
		want, _ := os.ReadFile(files[i])
		got, _ := os.ReadFile(filepath.Join(recvDir, r.SavedAs))
		if !bytes.Equal(want, got) {
			t.Errorf("%s chegou diferente", r.File)
		}
	}
	if results[10].Err == nil {
		t.Error("arquivo inexistente deveria dar erro")
	}
}

// Com o token errado o receptor responde 401 sem ler o corpo. SendAll tem que
// devolver o erro e só retornar depois que nenhuma goroutine vai mais mandar
// progresso: fechar o channel logo em seguida não pode causar panic.
func TestSendAllTokenErrado(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hub := progress.NewHub()
	go hub.Run(ctx)
	ts := httptest.NewServer(server.New(server.Config{Dir: t.TempDir(), Token: "certo", Hub: hub}).Handler())
	defer ts.Close()

	p := filepath.Join(t.TempDir(), "grande.bin")
	os.WriteFile(p, make([]byte, 8<<20), 0o644) // 8 MB

	base, _ := ParseURL(ts.URL + "/?t=errado")
	prog := make(chan Progress)
	go func() {
		for range prog {
		}
	}()
	s := &Sender{BaseURL: base, Parallel: 2, Progress: prog}
	results := s.SendAll(ctx, []string{p, p})
	close(prog) // se alguma goroutine ainda mandasse progresso, daria panic aqui

	for _, r := range results {
		if r.Err == nil || !strings.Contains(r.Err.Error(), "401") {
			t.Errorf("esperava erro 401, veio %v", r.Err)
		}
	}
}

func TestParseURLExigeToken(t *testing.T) {
	if _, err := ParseURL("192.168.0.10:8080"); err == nil {
		t.Error("URL sem token deveria ser recusada")
	}
	if _, err := ParseURL("192.168.0.10:8080/?t=abc"); err != nil {
		t.Error(err)
	}
}
