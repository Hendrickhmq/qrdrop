# Como contribuir

## Fluxo

1. Pegue uma issue e atribua a si mesmo.
2. Crie uma branch a partir da `main`: `git switch -c 12-upload-multipart` (número da issue + descrição curta).
3. Commits pequenos, no padrão [Conventional Commits](https://www.conventionalcommits.org/pt-br/): `tipo: descrição`, com a descrição em português, no imperativo e em minúsculas. Exemplo: `feat: adiciona handler de upload`.

   | Tipo | Quando usar |
   |---|---|
   | `feat` | Nova funcionalidade |
   | `fix` | Correção de bug |
   | `docs` | Só documentação |
   | `test` | Adiciona ou corrige testes |
   | `refactor` | Mudança de código que não altera comportamento |
   | `ci` | Workflows e configuração do CI |
   | `chore` | Manutenção (dependências, scripts, configuração) |
4. Abra um PR com `Closes #12` na descrição.
5. **Pelo menos 1 revisão de outro integrante** antes do merge. Revisar também é aprender o código que você vai ter que explicar na apresentação.
6. O CI (build + vet + test) precisa passar.

## Padrões de código

- `gofmt` sempre (o editor faz sozinho com a extensão de Go).
- `go vet ./...` sem avisos.
- Erros: trate ou retorne com contexto (`fmt.Errorf("salvando %s: %w", nome, err)`). Nunca descarte com `_` sem comentário.
- Funções exportadas (maiúscula) com comentário de documentação.

## Regra de ouro

Se você não consegue explicar uma linha que escreveu (ou que uma IA escreveu para você), ela não entra. A nota da apresentação é individual.
