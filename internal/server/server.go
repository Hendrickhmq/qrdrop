package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"qrdrop/internal/progress"
	"qrdrop/internal/storage"
	"qrdrop/web"
)

// Server recebe arquivos e publica o progresso no Hub.
type Server struct {
	Dir   string        // pasta de destino
	Token string        // token exigido em toda requisição (menos /healthz)
	Hub   *progress.Hub // para onde vão os eventos de progresso
	Log   *log.Logger
}

// Handler monta as rotas. Server não precisa implementar http.Handler ele
// mesmo: devolvemos um ServeMux, que já implementa (interface implícita).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.requireToken(s.handleIndex))
	mux.HandleFunc("POST /upload", s.requireToken(s.handleUpload))
	mux.HandleFunc("GET /events", s.requireToken(s.handleEvents))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}

// requireToken é um middleware: recebe um handler e devolve outro que só
// chama o original se o token estiver certo. Funções como valores + closure.
func (s *Server) requireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got := r.URL.Query().Get("t")
		if got == "" {
			got = r.Header.Get("X-Token")
		}
		// Comparação em tempo constante: não vaza quantos caracteres acertaram.
		if subtle.ConstantTimeCompare([]byte(got), []byte(s.Token)) != 1 {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, web.Files, "index.html")
}

// UploadResult é a resposta JSON de POST /upload.
type UploadResult struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// handleUpload recebe UM arquivo por requisição. O cliente manda vários
// arquivos em requisições paralelas, e o net/http já atende cada requisição
// na sua própria goroutine.
//
// Cabeçalhos opcionais:
//
//	X-Upload-ID   identificador para casar os eventos de progresso
//	X-File-Size   tamanho do arquivo, para calcular a porcentagem
//	X-SHA256      hash esperado; se não bater, o arquivo é apagado (422)
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("X-Upload-ID")
	total, err := strconv.ParseInt(r.Header.Get("X-File-Size"), 10, 64)
	if err != nil {
		total = -1
	}

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "esperava multipart/form-data", http.StatusBadRequest)
		return
	}
	part, err := mr.NextPart()
	if err != nil {
		http.Error(w, "nenhum arquivo enviado", http.StatusBadRequest)
		return
	}
	defer part.Close()

	res, err := s.save(id, part.FileName(), total, part, r.Header.Get("X-SHA256"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errHashMismatch) {
			status = http.StatusUnprocessableEntity
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

var errHashMismatch = errors.New("SHA-256 não confere: arquivo corrompido no caminho")

// save grava src em disco em streaming, calculando o SHA-256 no mesmo passo.
func (s *Server) save(id, name string, total int64, src io.Reader, wantHash string) (res UploadResult, err error) {
	f, err := storage.Create(s.Dir, name)
	if err != nil {
		return res, err
	}
	path := f.Name()
	res.Name = filepath.Base(path)

	// Um único defer cuida de fechar o arquivo, apagar o que ficou pela
	// metade e avisar o hub, qualquer que seja o caminho de saída da função.
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
		ev := progress.Event{ID: id, Name: res.Name, Bytes: res.Size, Total: total, Done: true}
		if err != nil {
			os.Remove(path)
			ev.Err = err.Error()
		} else {
			ev.SHA256 = res.SHA256
		}
		s.Hub.Publish(ev)
		s.logf("%s: %s", res.Name, status(err, res))
	}()

	h := sha256.New()
	pw := &progressWriter{hub: s.Hub, ev: progress.Event{ID: id, Name: res.Name, Total: total}}
	// io.MultiWriter: cada pedaço lido vai ao mesmo tempo para o disco,
	// para o hash e para o contador de progresso. Memória usada: só o buffer
	// do io.Copy (32 KB), não importa o tamanho do arquivo.
	res.Size, err = io.Copy(io.MultiWriter(f, h, pw), src)
	if err != nil {
		return res, fmt.Errorf("recebendo %s: %w", name, err)
	}
	res.SHA256 = hex.EncodeToString(h.Sum(nil))
	if wantHash != "" && wantHash != res.SHA256 {
		return res, errHashMismatch
	}
	return res, nil
}

func status(err error, res UploadResult) string {
	if err != nil {
		return "ERRO: " + err.Error()
	}
	return fmt.Sprintf("recebido, %s, sha256 %s", humanBytes(res.Size), res.SHA256[:12])
}

func (s *Server) logf(format string, args ...any) {
	if s.Log != nil {
		s.Log.Printf(format, args...)
	}
}

// progressWriter implementa io.Writer só para contar bytes e publicar
// eventos, no máximo um a cada 100 ms para não inundar o hub.
type progressWriter struct {
	hub  *progress.Hub
	ev   progress.Event
	last time.Time
}

func (p *progressWriter) Write(b []byte) (int, error) {
	p.ev.Bytes += int64(len(b))
	if now := time.Now(); now.Sub(p.last) >= 100*time.Millisecond {
		p.last = now
		p.hub.Publish(p.ev)
	}
	return len(b), nil
}

// handleEvents mantém a conexão aberta e manda cada evento de progresso
// como Server-Sent Events (texto "data: {...}\n\n").
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming não suportado", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, ": conectado\n\n")
	flusher.Flush()

	events := s.Hub.Subscribe()
	defer s.Hub.Unsubscribe(events)

	for {
		select {
		case <-r.Context().Done(): // a página foi fechada
			return
		case ev, ok := <-events:
			if !ok { // o servidor está desligando
				return
			}
			data, _ := json.Marshal(ev) // Event só tem tipos simples: Marshal não falha
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
