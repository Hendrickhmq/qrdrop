# Análise da linguagem Go

Este documento é a base do vídeo (máx. 10 min). A estrutura segue **exatamente** os itens da planilha de avaliação. Cada seção indica o peso e o que precisa aparecer.

> Regra do grupo: todo exemplo de código aqui deve vir **do nosso repositório**, com link para o arquivo. A planilha cobra "exemplos de código na LP que foram [feitos] pelo grupo".

---

## 1. Histórico e versão (0,5)

- [ ] Origem: Google, 2007; Robert Griesemer, Rob Pike, Ken Thompson
- [ ] Motivação: builds lentos de C++, dificuldade com concorrência em servidores, complexidade das linguagens da época
- [ ] Marcos: anúncio público (2009), Go 1.0 e a promessa de compatibilidade (2012), módulos (1.11–1.16), generics (1.18)
- [ ] Versão usada no projeto: _preencher (ver `go.mod`)_
- [ ] Fontes: _preencher_

## 2. Domínio de aplicação, usuário característico e premissas de projeto (1,5)

Seção de maior peso. Precisa ligar **premissa da linguagem → decisão no nosso projeto**.

- [ ] Domínio: servidores, infraestrutura de nuvem, ferramentas de linha de comando (Docker, Kubernetes, Terraform)
- [ ] Usuário característico: equipes grandes mantendo serviços de rede
- [ ] Premissas de projeto:
  - Simplicidade (poucas palavras-chave, uma forma de fazer as coisas, `gofmt`)
  - Compilação rápida
  - Concorrência como recurso da linguagem (CSP de Hoare)
  - Composição em vez de herança
  - Erros explícitos
- [ ] Relação com os critérios do Sebesta (cap. 1): legibilidade, facilidade de escrita, confiabilidade, custo
- [ ] Onde cada premissa aparece no qrdrop: _preencher com links_

## 3. Construtores da linguagem (1,0)

Para cada item: explicação curta + trecho do nosso código.

| Construtor | Onde aparece no qrdrop |
|---|---|
| Tipos básicos e compostos (struct, slice, map) | _link_ |
| Declarações (`var`, `:=`, `const`, inferência de tipos) | _link_ |
| Comandos de controle (`for` único, `switch`, `select`) | _link_ |
| Funções: múltiplos retornos, funções como valores, closures | _link_ |
| `defer` | _link_ |
| Interfaces implícitas (`io.Reader`, `http.Handler`) | _link_ |
| Goroutines e channels | _link_ |
| Tratamento de erro (`error`, `if err != nil`, `errors.Is`) | _link_ |
| Pacotes e visibilidade (maiúscula/minúscula, `internal/`) | _link_ |

### Ganchos com o conteúdo da disciplina
- **Abstração (aula 07):** interfaces, encapsulamento por pacote
- **Controle de sequência (aula 08):** `select`, `defer`, ausência de exceções
- **Subprogramas (aula 09):** passagem de parâmetros por valor (inclusive de ponteiros), closures, goroutines como subprogramas concorrentes
- **Tradução (aula 05):** compilação para código nativo, binário estático, compilação cruzada

## 4. Exemplos de código feitos pelo grupo (1,0)

- [ ] 3 a 5 trechos curtos e comentados, cada um mostrando um recurso diferente
- [ ] Sugestões: handler de upload com `io.Copy` em streaming; hub de progresso com `select`; geração de token; `embed.FS`

## 5. Pontos fracos e críticas (recomendado)

Uma análise honesta vale mais que propaganda.
- [ ] Verbosidade do tratamento de erro
- [ ] Generics tardios e limitados
- [ ] Garbage collector (latência) vs. Rust/C
- [ ] Ausência de enums ricos e de tipos soma
