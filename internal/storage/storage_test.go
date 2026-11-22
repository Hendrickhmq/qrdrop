package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSafeName(t *testing.T) {
	// Teste de tabela: o jeito idiomático de testar em Go.
	cases := []struct{ in, want string }{
		{"foto.jpg", "foto.jpg"},
		{"../../etc/passwd", "passwd"},
		{`C:\Users\ana\doc.pdf`, "doc.pdf"},
		{".bashrc", "bashrc"},
		{"", "arquivo"},
		{"..", "arquivo"},
		{"a:b?.txt", "a_b_.txt"},
	}
	for _, c := range cases {
		if got := SafeName(c.in); got != c.want {
			t.Errorf("SafeName(%q) = %q, queria %q", c.in, got, c.want)
		}
	}
}

func TestCreateNaoSobrescreve(t *testing.T) {
	dir := t.TempDir()
	const n = 20
	var wg sync.WaitGroup
	names := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) { // 20 goroutines criando "foto.jpg" ao mesmo tempo
			defer wg.Done()
			f, err := Create(dir, "foto.jpg")
			if err != nil {
				t.Error(err)
				return
			}
			names[i] = filepath.Base(f.Name())
			f.Close()
		}(i)
	}
	wg.Wait()

	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			t.Fatalf("nome repetido: %s", name)
		}
		seen[name] = true
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != n {
		t.Fatalf("esperava %d arquivos, achei %d", n, len(entries))
	}
	if !seen["foto.jpg"] || !seen["foto (1).jpg"] {
		t.Errorf("nomes inesperados: %v", names)
	}
}
