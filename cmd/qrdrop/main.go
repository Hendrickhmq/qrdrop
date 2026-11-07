// Comando qrdrop: recebe arquivos do celular via QR code na rede local.
//
// Esqueleto inicial. A implementação está dividida nas issues do marco
// "M1 – Seminário 2". Veja docs/arquitetura.md.
package main

import (
	"flag"
	"fmt"
)

func main() {
	dir := flag.String("dir", "recebidos", "pasta onde os arquivos serão salvos")
	port := flag.Int("port", 8080, "porta do servidor HTTP")
	flag.Parse()

	fmt.Printf("qrdrop: salvando em %q, porta %d (ainda não implementado)\n", *dir, *port)
}
