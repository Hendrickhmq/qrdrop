#!/usr/bin/env bash
# Cria labels, marcos e as issues iniciais do qrdrop no repositório GitHub atual.
# Uso (de dentro do repo, com `gh auth login` feito):
#   ./scripts/criar-issues.sh
# Rode UMA vez: rodar de novo duplica as issues.
set -euo pipefail

REPO="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
echo "Repositório: $REPO"

# ---------- labels ----------
label() { gh label create "$1" --color "$2" --description "$3" --force >/dev/null; }
label "area:backend"      "1d76db" "Servidor, upload, disco, flags"
label "area:concorrencia" "5319e7" "Goroutines, channels, SSE, token"
label "area:frontend"     "0e8a16" "Página mobile, QR, IP local"
label "area:analise"      "fbca04" "Análise da linguagem e vídeo"
label "area:infra"        "c5def5" "CI, build, site, coordenação"
label "camada:mvp"        "d93f0b" "Necessário para o seminário 2"
label "camada:vitrine"    "f9d0c4" "Recursos que evidenciam Go"
label "camada:extra"      "ededed" "Só se sobrar tempo"

# ---------- marcos ----------
milestone() {
  local title="$1" due="${2:-}"
  if [[ -n "$due" ]]; then
    gh api "repos/$REPO/milestones" -f title="$title" -f due_on="$due" >/dev/null
  else
    gh api "repos/$REPO/milestones" -f title="$title" >/dev/null
  fi
}
M1="M1 – Seminário 2"
M2="M2 – Vitrine Go"
M3="M3 – Entrega final"
milestone "$M1" "2026-11-11T23:59:59Z"
milestone "$M2"
milestone "$M3"

# ---------- issues ----------
issue() { # título, labels, marco, corpo
  gh issue create --title "$1" --label "$2" --milestone "$3" --body "$4" >/dev/null
  echo "  + $1"
}

echo "Criando issues..."

# ===== M1 – Seminário 2 =====
issue "Configurar repositório, go.mod e CI" "area:infra,camada:mvp" "$M1" \
"## Objetivo
Repo pronto para todos trabalharem.

## Critérios de aceite
- [ ] Todos os 5 integrantes com acesso de escrita
- [ ] Proteção da \`main\`: exige PR + 1 aprovação + CI verde
- [ ] CI (\`.github/workflows/ci.yml\`) passando
- [ ] Tabela de equipe do README preenchida

## Conceito de Go
Módulos (\`go.mod\`), toolchain unificada (\`go vet\`, \`gofmt\`, \`go test\`)."

issue "Servidor HTTP básico servindo a página de upload" "area:backend,camada:mvp" "$M1" \
"## Objetivo
\`go run ./cmd/qrdrop\` sobe um servidor e \`GET /\` devolve \`web/index.html\`.

## Critérios de aceite
- [ ] \`http.NewServeMux\` com rotas \`/\` e \`/healthz\`
- [ ] Porta vinda da flag \`-port\`
- [ ] Erro ao abrir a porta é tratado com mensagem clara

## Conceito de Go
\`net/http\`, interface \`http.Handler\`, funções como valores (\`http.HandlerFunc\`)."

issue "Receber upload multipart e salvar na pasta de destino" "area:backend,camada:mvp" "$M1" \
"## Objetivo
\`POST /upload\` recebe um arquivo e grava em \`-dir\`.

## Critérios de aceite
- [ ] Leitura em **streaming** com \`r.MultipartReader()\` + \`io.Copy\` (não carregar tudo em memória)
- [ ] Pasta criada se não existir (\`os.MkdirAll\`)
- [ ] Arquivo fechado com \`defer\`
- [ ] Testado com arquivo > 100 MB

## Conceito de Go
\`io.Reader\`/\`io.Writer\` (interfaces implícitas), \`defer\`, erros como valores."

issue "Descobrir o IP local da máquina" "area:frontend,camada:mvp" "$M1" \
"## Objetivo
Função que retorna o IPv4 da interface de rede local (ex.: 192.168.0.12).

## Critérios de aceite
- [ ] Ignora loopback e interfaces desligadas
- [ ] Retorna \`(net.IP, error)\`
- [ ] Teste unitário da função de filtragem

## Conceito de Go
Múltiplos valores de retorno, pacote \`net\`, slices."

issue "Exibir QR code da URL no terminal" "area:frontend,camada:mvp" "$M1" \
"## Objetivo
Ao iniciar, imprime o QR de \`http://IP:PORTA/?t=TOKEN\` e a URL em texto.

## Critérios de aceite
- [ ] Avaliar \`github.com/mdp/qrterminal/v3\` e justificar a escolha no PR
- [ ] Funciona no terminal do Windows e do Linux
- [ ] Celular escaneia de primeira

## Conceito de Go
Gerenciamento de dependências com módulos (\`go get\`, \`go.sum\`)."

issue "Página mobile de envio de múltiplos arquivos" "area:frontend,camada:mvp" "$M1" \
"## Objetivo
Página simples, usável no celular, com seleção de vários arquivos.

## Critérios de aceite
- [ ] \`<input type=file multiple>\` (abre a galeria/câmera no celular)
- [ ] Cada arquivo enviado numa requisição \`fetch\` separada, **em paralelo**
- [ ] Mostra sucesso/erro por arquivo
- [ ] Sem framework JS

## Conceito de Go
(Indireto) o paralelismo do cliente é o que exercita as goroutines do servidor."

issue "Slides e demo do seminário 2" "area:analise,camada:mvp" "$M1" \
"## Objetivo
Apresentação de status + configuração do grupo (roteiro em \`docs/planejamento.md\`).

## Critérios de aceite
- [ ] 5–6 slides
- [ ] Demo ensaiada **no Wi-Fi da sala** (testar se a rede isola clientes; plano B: roteador do celular)
- [ ] Todos sabem explicar o fluxo de ponta a ponta"

# ===== M2 – Vitrine Go =====
issue "Token de sessão no QR e validação" "area:concorrencia,camada:vitrine" "$M2" \
"## Objetivo
Só quem escaneou o QR consegue enviar arquivos.

## Critérios de aceite
- [ ] Token de 128 bits com \`crypto/rand\`, em hex/base64url
- [ ] Middleware que valida \`?t=\` e responde 401
- [ ] Comparação em tempo constante (\`crypto/subtle\`)
- [ ] Teste com \`httptest\` para token válido e inválido

## Conceito de Go
Middleware como função que recebe e retorna \`http.Handler\` (closures, composição)."

issue "Hub de progresso com goroutine e channels" "area:concorrencia,camada:vitrine" "$M2" \
"## Objetivo
Uploads publicam progresso; um hub distribui para os inscritos.

## Critérios de aceite
- [ ] \`type Progress struct { ID, Name string; Bytes, Total int64 }\`
- [ ] Um único goroutine dono do mapa de clientes, com \`select\` entre registrar, remover e publicar
- [ ] Sem mutex no hub
- [ ] \`go test -race\` passa com 50 uploads simultâneos

## Conceito de Go
CSP: goroutines, channels, \`select\`. *Share memory by communicating.*"

issue "Contador de bytes para medir o progresso do upload" "area:concorrencia,camada:vitrine" "$M2" \
"## Objetivo
Envolver o \`io.Reader\` do upload num tipo que conta bytes e publica no hub.

## Critérios de aceite
- [ ] \`type progressReader struct\` que implementa \`Read\`
- [ ] Publica no máximo ~10 eventos/s por arquivo

## Conceito de Go
Interfaces implícitas e composição. Bom exemplo para o vídeo."

issue "Stream de progresso via Server-Sent Events" "area:concorrencia,camada:vitrine" "$M2" \
"## Objetivo
\`GET /events\` envia o progresso em tempo real; a página mostra barras.

## Critérios de aceite
- [ ] \`Content-Type: text/event-stream\` + \`http.Flusher\`
- [ ] Cliente desconectado é removido do hub (\`r.Context().Done()\`)
- [ ] Barra de progresso por arquivo na página
- [ ] O terminal do PC também mostra o que está chegando

## Conceito de Go
\`context\`, type assertion (\`w.(http.Flusher)\`), \`select\`."

issue "Nomes de arquivo seguros e sem sobrescrita" "area:backend,camada:vitrine" "$M2" \
"## Objetivo
Corretude e segurança na gravação.

## Critérios de aceite
- [ ] \`filepath.Base\` + limpeza de caracteres: \`../../etc/passwd\` não escapa da pasta
- [ ] Se o nome existir, salva \`foto (1).jpg\`, \`foto (2).jpg\`…
- [ ] Criação atômica (\`os.OpenFile\` com \`O_CREATE|O_EXCL\`) para não haver corrida entre goroutines
- [ ] Limite de tamanho configurável (\`http.MaxBytesReader\`)
- [ ] Testes de tabela (table-driven tests)

## Conceito de Go
Table-driven tests, \`errors.Is(err, fs.ErrExist)\`."

issue "Embutir o frontend no binário com embed.FS" "area:backend,camada:vitrine" "$M2" \
"## Objetivo
O binário funciona sozinho, sem a pasta \`web/\` ao lado.

## Critérios de aceite
- [ ] \`//go:embed\` + \`http.FileServerFS\`
- [ ] Binário copiado para outra pasta continua servindo a página

## Conceito de Go
Diretivas do compilador, binário estático único."

issue "Testes de integração dos handlers com httptest" "area:backend,camada:vitrine" "$M2" \
"## Objetivo
Cobertura do fluxo principal sem precisar de celular.

## Critérios de aceite
- [ ] Upload de ponta a ponta com \`httptest.NewServer\`
- [ ] Uploads concorrentes no teste (\`sync.WaitGroup\`)
- [ ] CI roda com \`-race\`

## Conceito de Go
Testes na própria toolchain, \`sync.WaitGroup\`, detector de race."

issue "Builds para Windows, Linux e macOS em releases" "area:infra,camada:vitrine" "$M2" \
"## Objetivo
Qualquer pessoa baixa e roda, sem instalar Go.

## Critérios de aceite
- [ ] Workflow que, numa tag \`v*\`, compila com \`GOOS/GOARCH\` (windows/amd64, linux/amd64, darwin/arm64)
- [ ] Binários anexados à Release do GitHub
- [ ] Testado em pelo menos Windows e Linux

## Conceito de Go
Compilação cruzada nativa, sem dependência de libc (\`CGO_ENABLED=0\`)."

issue "Encerramento gracioso com Ctrl+C" "area:backend,camada:vitrine" "$M2" \
"## Objetivo
Ctrl+C espera os uploads em andamento terminarem (com timeout).

## Critérios de aceite
- [ ] \`signal.NotifyContext\` + \`srv.Shutdown(ctx)\`
- [ ] Mensagem com o resumo: N arquivos recebidos, X MB

## Conceito de Go
\`context\` com cancelamento, sinais do SO."

# ===== M3 – Entrega final =====
issue "Análise: histórico e versão do Go" "area:analise" "$M3" \
"Preencher \`docs/analise-go.md\` §1 com fontes. Peso 0,5."

issue "Análise: domínio, usuário e premissas de projeto" "area:analise" "$M3" \
"Preencher \`docs/analise-go.md\` §2. **Maior peso da análise (1,5).** Toda premissa precisa apontar para um trecho do nosso código. Relacionar com os critérios do Sebesta (cap. 1)."

issue "Análise: construtores e exemplos do nosso código" "area:analise" "$M3" \
"Preencher \`docs/analise-go.md\` §3 e §4 (peso 2,0 somados). **Cada integrante escreve o trecho sobre o código que implementou.** Ligar com as aulas 05, 07, 08 e 09."

issue "Site do projeto (GitHub Pages)" "area:infra" "$M3" \
"## Critérios de aceite
- [ ] Página com: o que é, GIF da demo, como baixar (link para Releases), equipe, link do vídeo
- [ ] Publicado via GitHub Pages
- [ ] URL no README e entregue no Aprender3

Peso 1,0."

issue "Roteiro, gravação e edição do vídeo (≤ 10 min)" "area:analise" "$M3" \
"## Critérios de aceite
- [ ] Roteiro cronometrado: análise da linguagem (~6 min) + projeto e demo (~4 min)
- [ ] Todos os integrantes aparecem/falam
- [ ] Duração final **≤ 10:00** (critério eliminatório de 0,5)
- [ ] Áudio limpo; demo gravada da tela do celular e do PC
- [ ] Enviado no Aprender3 na penúltima semana"

issue "Relatório de atividades por integrante" "area:infra" "$M3" \
"Consolidar, a partir das issues e PRs fechados, o que cada membro fez. Gerar com \`gh pr list --state merged --json author,title\` e revisar à mão."

# ===== Extras =====
issue "Envio no sentido inverso: PC para celular" "area:frontend,camada:extra" "$M3" \
"Pasta compartilhada listada na página; o celular baixa. Só depois do M2 completo."

issue "Descoberta na rede via mDNS (qrdrop.local)" "area:concorrencia,camada:extra" "$M3" \
"Anunciar o serviço por mDNS como alternativa ao IP. Só depois do M2 completo."

echo "Pronto."
