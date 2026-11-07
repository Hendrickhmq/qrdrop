# qrdrop

> Mande fotos e arquivos do celular para o PC escaneando um QR code. Sem cabo, sem WhatsApp, sem nuvem.

Projeto final da disciplina **CIC0093 – Linguagens de Programação** (UnB, 2026/2, Prof. Marcelo Ladeira).
Linguagem do projeto: **Go**.

## Como funciona

```
┌──────────────┐   1. roda `qrdrop`        ┌──────────────────────┐
│      PC      │ ───────────────────────▶  │ servidor HTTP local  │
│  (terminal)  │   2. mostra QR code        │ http://192.168.x.y:8080/?t=TOKEN
└──────────────┘                            └──────────▲───────────┘
                                                       │ 4. upload (multipart, em paralelo)
┌──────────────┐   3. escaneia o QR e abre a página    │
│   Celular    │ ──────────────────────────────────────┘
│ (navegador)  │   5. vê o progresso de cada arquivo em tempo real
└──────────────┘
```

O celular não instala nada: usa o navegador. O PC roda um único binário, que já traz a página web dentro dele.

## Uso (planejado)

```bash
qrdrop                       # salva em ./recebidos, porta 8080
qrdrop -dir ~/Fotos -port 9000
```

## Desenvolvimento

Requisitos: Go 1.22+.

```bash
go run ./cmd/qrdrop      # roda
go test ./...            # testes
go vet ./...             # análise estática
```

## Documentação

| Documento | Conteúdo |
|---|---|
| [docs/visao.md](docs/visao.md) | Problema, escopo, o que **não** faremos |
| [docs/arquitetura.md](docs/arquitetura.md) | Componentes, fluxo, endpoints, modelo de concorrência |
| [docs/analise-go.md](docs/analise-go.md) | Análise da linguagem Go (base do vídeo) |
| [docs/planejamento.md](docs/planejamento.md) | Papéis, marcos, critérios de avaliação |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Fluxo de branches, PRs e commits |

## Equipe

| Integrante | Papel principal |
|---|---|
| _nome_ | Backend HTTP e arquivos |
| _nome_ | Concorrência, progresso e segurança |
| _nome_ | Frontend mobile e QR code |
| _nome_ | Análise da linguagem e vídeo |
| _nome_ | Build, CI, site e coordenação |

## Referências

- [croc](https://github.com/schollz/croc), transferência de arquivos em Go (inspiração; escopo bem maior que o nosso)
- [The Go Programming Language Specification](https://go.dev/ref/spec)
