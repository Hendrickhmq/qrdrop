// Pacote storage grava os arquivos recebidos em disco com nomes seguros.
package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafeName limpa o nome de arquivo enviado pelo cliente.
//
// O nome vem da rede, então não dá para confiar nele: "../../etc/passwd"
// escaparia da pasta de destino (path traversal). Ficamos só com a última
// parte do caminho e removemos caracteres problemáticos.
func SafeName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/") // caminhos do Windows
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, name)
	name = strings.TrimLeft(strings.TrimSpace(name), ".")
	if name == "" {
		return "arquivo"
	}
	return name
}

// Create cria um arquivo novo em dir sem sobrescrever nenhum existente.
// Se "foto.jpg" já existe, tenta "foto (1).jpg", "foto (2).jpg" e assim por diante.
//
// O_EXCL faz a checagem e a criação numa operação só do sistema operacional,
// então duas goroutines recebendo arquivos com o mesmo nome ao mesmo tempo
// nunca ficam com o mesmo arquivo.
func Create(dir, name string) (*os.File, error) {
	name = SafeName(name)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for i := 0; i < 1000; i++ {
		candidate := name
		if i > 0 {
			candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		f, err := os.OpenFile(filepath.Join(dir, candidate), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("criando %s: %w", candidate, err)
		}
	}
	return nil, fmt.Errorf("muitos arquivos chamados %q", name)
}
