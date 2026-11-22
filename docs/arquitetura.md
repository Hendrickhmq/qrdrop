# Arquitetura

## Estrutura de pastas

```
cmd/qrdrop/        main: flags, inicialização, QR no terminal
internal/server/   handlers HTTP, upload, token, progresso (SSE)
internal/progress/ hub de progresso (channel + goroutine única)
internal/client/   `qrdrop send`: envio PC → PC com pool de goroutines
internal/netutil/  descoberta do IP local
internal/storage/  nomes seguros, gravação em disco
web/               index.html, JS e CSS da página mobile (embutidos no binário)
docs/              documentação e análise da linguagem
```

`internal/` é uma convenção do próprio compilador Go: esses pacotes não podem ser importados por outros módulos. Também é um bom exemplo para a seção de abstração/encapsulamento.

## Fluxo

1. `main` lê as flags (`-dir`, `-port`), cria a pasta de destino e gera um **token aleatório** (`crypto/rand`).
2. Descobre o IP da máquina na rede local e monta `http://IP:PORTA/?t=TOKEN`.
3. Imprime o QR code dessa URL no terminal.
4. Sobe o servidor HTTP. **Cada requisição roda na própria goroutine**: isso é do `net/http`, não escrevemos nada para isso.
5. O celular abre a página, escolhe os arquivos, e o JS envia **cada arquivo numa requisição separada, em paralelo**.
6. O handler de upload lê o corpo em streaming (`multipart.Reader` + `io.Copy`), sem carregar o arquivo inteiro na memória, e publica o progresso num channel.
7. Um **hub de progresso** (uma goroutine) recebe os eventos do channel e repassa para os clientes conectados em `/events` via SSE.

## Envio pelo terminal (`qrdrop send`)

Para mandar de um PC para outro sem navegador. Um pool de N goroutines (`-p`, padrão 4) consome um channel de tarefas; cada uma calcula o SHA-256 local, envia o arquivo em streaming (`io.Pipe` + `multipart.Writer`) e compara o hash devolvido pelo receptor. O progresso de todas vai por um channel para uma goroutine que desenha a linha no terminal.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/?t=TOKEN` | Página de upload (servida de `embed.FS`) |
| POST | `/upload?t=TOKEN` | Recebe um arquivo (multipart). Cabeçalhos opcionais: `X-Upload-ID`, `X-File-Size`, `X-SHA256` (se não bater, responde 422 e apaga o arquivo) |
| GET | `/events?t=TOKEN` | Stream SSE de progresso |
| GET | `/healthz` | Verificação simples |

Requisições sem token válido recebem `401`.

## Modelo de concorrência

```
 requisição A ─▶ goroutine A ─┐
 requisição B ─▶ goroutine B ─┼──▶ chan Progress ──▶ goroutine hub ──▶ clientes SSE
 requisição C ─▶ goroutine C ─┘
```

- Os produtores (uploads) não conhecem os consumidores (telas): o channel desacopla os dois.
- O hub é o único dono do mapa de clientes, então ele não precisa de mutex. É o lema do Go: *"Don't communicate by sharing memory; share memory by communicating."*
- Escrita em disco: cada arquivo tem nome único, então não há disputa entre goroutines.

## Decisões e alternativas

| Decisão | Alternativa descartada | Motivo |
|---|---|---|
| SSE para progresso | WebSocket | SSE é unidirecional, sai da stdlib e basta para o caso |
| Um arquivo por requisição | Todos numa requisição | Paralelismo real e progresso individual |
| Token na URL | Login/senha | O QR já é o canal de autenticação |
| Sem framework (Gin, Echo) | Gin | A stdlib basta e evidencia a linguagem |

## Dependências externas

- Geração de QR no terminal: [`github.com/mdp/qrterminal/v3`](https://github.com/mdp/qrterminal) (a confirmar na issue correspondente)
- O resto é biblioteca padrão.
