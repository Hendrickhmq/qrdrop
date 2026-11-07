# Planejamento

## Critérios de avaliação (planilha do professor)

| Critério | Peso | Como atacamos | Dono |
|---|---|---|---|
| Duração do vídeo ≤ 10 min | 0,5 | Roteiro cronometrado | Análise/vídeo |
| Histórico e versão da linguagem | 0,5 | `docs/analise-go.md` §1 | Análise/vídeo |
| Domínio, usuário e premissas de projeto | 1,5 | `docs/analise-go.md` §2, ligando premissas ao código | Análise/vídeo + todos |
| Construtores da linguagem | 1,0 | `docs/analise-go.md` §3 | Todos (cada um escreve sobre seu código) |
| Exemplos de código do grupo | 1,0 | Trechos reais do repo | Todos |
| Complexidade do projeto | 1,0 | Camada 2 completa | Devs |
| Corretude do projeto | 1,5 | Testes, tratamento de erro, demo ao vivo | Devs |
| Qualidade do vídeo | 0,5 | Áudio limpo, demo gravada | Análise/vídeo |
| Apresentação (**nota individual**) | 1,5 | Todos falam e todos entendem o código | Todos |
| Site | 1,0 | GitHub Pages | Build/site |
| **Bônus** melhor projeto | +1,0 | Polimento da demo | Todos |

> A apresentação é nota **individual**. Ninguém pode ficar só com "a parte de slides": cada integrante implementa pelo menos uma issue de código e explica o próprio código.

## Papéis

| Papel | Responsabilidade | Issues principais |
|---|---|---|
| **Backend** | Servidor, upload, gravação em disco, flags | `area:backend` |
| **Concorrência** | Token, hub de progresso, SSE, testes de concorrência | `area:concorrencia` |
| **Frontend/QR** | Página mobile, QR no terminal, IP local | `area:frontend` |
| **Análise/vídeo** | Documento de análise, roteiro, gravação, edição | `area:analise` |
| **Build/site/coord.** | CI, releases multiplataforma, site, quadro de issues, relatório de atividades | `area:infra` |

## Marcos

| Marco | Data | Entrega |
|---|---|---|
| **M1 – Seminário 2** | qua 11/11/2026 | MVP rodando ao vivo, repo, papéis, arquitetura |
| **M2 – Vitrine Go** | _definir (≈ 2 semanas antes do vídeo)_ | Camada 2 completa e testada |
| **M3 – Entrega final** | _penúltima semana do curso (ver Aprender3)_ | Vídeo, site, relatório de atividades |

## Seminário 2: roteiro sugerido (status + configuração do grupo)

1. Problema e proposta (1 slide)
2. Por que Go: tabela premissa → recurso (1 slide)
3. Arquitetura e modelo de concorrência (1 slide)
4. Equipe e papéis (1 slide)
5. **Demo ao vivo**: escanear o QR e enviar uma foto
6. Próximos passos: camadas 2 e 3, cronograma (1 slide)

## Registro de atividades

Para o item "descrição das atividades que cada membro executou": o histórico de issues e PRs **é** o registro. Cada um fecha suas próprias issues e abre seus próprios PRs. No fim, o script/relatório consolida por autor.
