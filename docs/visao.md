# Visão do projeto

## Problema

Passar uma foto do celular para o PC hoje costuma envolver cabo, WhatsApp Web, e-mail para si mesmo ou nuvem. Todas essas opções são lentas, comprimem a imagem ou mandam um arquivo local para um servidor externo sem necessidade.

## Proposta

Um programa de linha de comando que roda no PC, sobe um servidor web na rede local e mostra um QR code. O celular escaneia, abre uma página no navegador e envia os arquivos direto para o PC, pela rede de casa.

## Usuário

Quem tem celular e PC na mesma rede Wi-Fi e quer mandar arquivos de um para o outro sem instalar app no celular.

## Por que Go

O projeto foi escolhido para exercitar as premissas de projeto da linguagem:

| Necessidade do projeto | Recurso do Go |
|---|---|
| Atender vários uploads ao mesmo tempo | goroutines (uma por requisição, nativo no `net/http`) |
| Repassar o progresso dos uploads para a tela | channels |
| Servidor web sem framework | biblioteca padrão: `net/http`, `mime/multipart`, `io` |
| Distribuir sem exigir instalação | binário estático único + `embed` + compilação cruzada |
| Robustez com rede e disco | erros como valores, `defer` para fechar recursos |

## Escopo

### Camada 1: MVP (seminário 2)
- Servidor HTTP com página de upload
- Descoberta do IP local e QR code no terminal
- Upload de múltiplos arquivos, salvos numa pasta

### Camada 2: vitrine do Go (entrega final)
- Token de sessão no QR (só quem escaneou consegue enviar)
- Progresso em tempo real por arquivo (Server-Sent Events alimentados por channels)
- Frontend embutido no binário (`embed.FS`)
- Nomes de arquivo seguros (sem path traversal, sem sobrescrever)
- Testes com `net/http/httptest`
- Binários para Windows, Linux e macOS

### Camada 3: se sobrar tempo
- Envio no sentido inverso (PC → celular)
- Descoberta na rede via mDNS

## Fora do escopo (decisão explícita)

- **Transferência pela internet / relay**, o que o croc faz. Exige servidor intermediário, NAT traversal e criptografia ponta a ponta. Multiplica a complexidade sem melhorar a análise da linguagem.
- **Criptografia própria.** Na rede local usamos token aleatório; HTTPS fica como nota de limitação.
- **App nativo no celular.** O navegador basta.
