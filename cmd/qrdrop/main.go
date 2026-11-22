// Comando qrdrop: transfere arquivos na rede local.
//
//	qrdrop [-dir recebidos] [-port 8080]       recebe (mostra o QR code)
//	qrdrop send [-p 4] URL arquivo...          envia para outro PC
//
// Veja docs/arquitetura.md.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mdp/qrterminal/v3"

	"qrdrop/internal/client"
	"qrdrop/internal/netutil"
	"qrdrop/internal/progress"
	"qrdrop/internal/server"
)

func main() {
	var err error
	if len(os.Args) > 1 && os.Args[1] == "send" {
		err = runSend(os.Args[2:])
	} else {
		err = runReceive(os.Args[1:])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "qrdrop:", err)
		os.Exit(1)
	}
}

// ---------- receber ----------

func runReceive(args []string) error {
	fs := flag.NewFlagSet("qrdrop", flag.ExitOnError)
	dir := fs.String("dir", "recebidos", "pasta onde os arquivos serão salvos")
	port := fs.Int("port", 8080, "porta do servidor HTTP")
	fs.Parse(args)

	if err := os.MkdirAll(*dir, 0o755); err != nil {
		return fmt.Errorf("criando a pasta %s: %w", *dir, err)
	}
	token, err := newToken()
	if err != nil {
		return err
	}
	ip, err := netutil.LocalIP()
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		return fmt.Errorf("abrindo a porta %d (já tem outro programa nela?): %w", *port, err)
	}

	// Ctrl+C cancela o contexto, e tudo que depende dele para.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	hub := progress.NewHub()
	go hub.Run(ctx)

	srv := &server.Server{Dir: *dir, Token: token, Hub: hub, Log: log.New(os.Stdout, "", log.Ltime)}
	httpSrv := &http.Server{Handler: srv.Handler()}

	url := fmt.Sprintf("http://%s:%d/?t=%s", ip, *port, token)
	fmt.Println("Escaneie com o celular ou abra no navegador de outro computador:")
	qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)
	fmt.Println(url)
	fmt.Printf("\nDe outro PC: qrdrop send %q arquivo...\n", url)
	fmt.Printf("Salvando em %s. Ctrl+C para sair.\n\n", *dir)

	// Desligamento gracioso: no Ctrl+C o servidor para de aceitar conexões
	// e espera até 5 s os envios em andamento terminarem.
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintln(os.Stderr, "qrdrop: envios interrompidos ao sair:", err)
		}
	}()

	// Serve retorna assim que Shutdown é chamado; esperamos o Shutdown acabar.
	if err := httpSrv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-shutdownDone
	fmt.Println("\ntchau!")
	return nil
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("gerando token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ---------- enviar ----------

func runSend(args []string) error {
	fs := flag.NewFlagSet("qrdrop send", flag.ExitOnError)
	parallel := fs.Int("p", 4, "quantos arquivos enviar ao mesmo tempo")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: qrdrop send [-p 4] URL arquivo...")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	if fs.NArg() < 2 {
		fs.Usage()
		os.Exit(2)
	}
	base, err := client.ParseURL(fs.Arg(0))
	if err != nil {
		return err
	}
	files := fs.Args()[1:]

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prog := make(chan client.Progress, 64)
	printerDone := make(chan struct{})
	go printProgress(prog, printerDone)

	s := &client.Sender{BaseURL: base, Parallel: *parallel, Progress: prog}
	start := time.Now()
	results := s.SendAll(ctx, files)
	close(prog)
	<-printerDone

	failed := 0
	for _, r := range results {
		if r.Err != nil {
			failed++
			fmt.Printf("ERRO  %s: %v\n", r.File, r.Err)
		} else {
			fmt.Printf("OK    %s -> %s (sha256 %s, integridade conferida)\n", r.File, r.SavedAs, r.SHA256[:12])
		}
	}
	fmt.Printf("%d de %d arquivos enviados em %s\n", len(results)-failed, len(results), time.Since(start).Round(time.Millisecond))
	if failed > 0 {
		return fmt.Errorf("%d arquivo(s) falharam", failed)
	}
	return nil
}

// printProgress redesenha uma linha com a porcentagem de cada arquivo.
// Recebe do channel até ele ser fechado; o ticker limita a 5 redesenhos/s.
func printProgress(in <-chan client.Progress, done chan<- struct{}) {
	defer close(done)
	state := map[string]client.Progress{}
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	width := 0 // tamanho da última linha desenhada

	draw := func() {
		names := make([]string, 0, len(state))
		for name := range state {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			p := state[name]
			pct := 100
			if p.Total > 0 {
				pct = int(p.Sent * 100 / p.Total)
			}
			parts = append(parts, fmt.Sprintf("%s %3d%%", short(name), pct))
		}
		line := strings.Join(parts, " | ")
		// "\r" volta ao início da linha; completar com espaços apaga o que
		// sobrou da linha anterior (funciona até no prompt antigo do Windows).
		n := utf8.RuneCountInString(line)
		fmt.Printf("\r%s%s", line, strings.Repeat(" ", max(0, width-n)))
		width = n
	}

	for {
		select {
		case p, ok := <-in:
			if !ok {
				draw()
				fmt.Println()
				return
			}
			state[p.File] = p
		case <-tick.C:
			draw()
		}
	}
}

func short(path string) string {
	name := path[strings.LastIndexAny(path, `/\`)+1:]
	if r := []rune(name); len(r) > 18 {
		return string(r[:15]) + "..."
	}
	return name
}
