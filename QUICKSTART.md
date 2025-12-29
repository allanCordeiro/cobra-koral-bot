# Quick Start Guide

## Para Desenvolvimento Local

### 1. Setup Inicial (primeira vez)

```bash
# Clone o repositório
git clone <repository-url>
cd cobra-coral

# Configure o ambiente
make dev-setup

# Edite o .env com suas credenciais do Telegram
nano .env
# ou
vim .env
```

### 2. Configurar Credenciais do Telegram

No arquivo `.env`, defina:

```bash
BOT_TOKEN=seu_token_do_bot_aqui
USER_ID=seu_id_de_usuario_aqui
```

**Como obter estas credenciais:**

1. **BOT_TOKEN**: Fale com [@BotFather](https://t.me/botfather) no Telegram
   - Envie `/newbot`
   - Siga as instruções
   - Copie o token fornecido

2. **USER_ID**: Fale com [@userinfobot](https://t.me/userinfobot) no Telegram
   - Envie qualquer mensagem
   - O bot retornará seu User ID

### 3. Executar Localmente

```bash
# Executar diretamente
make run

# Ou compilar e executar
make build
./bin/worker
```

### 4. Executar Testes

```bash
# Testes simples
make test

# Testes com coverage
make test-coverage
```

## Comandos Úteis

```bash
make help          # Ver todos os comandos disponíveis
make build         # Compilar o worker
make run           # Executar o worker
make test          # Rodar testes
make clean         # Limpar artefatos
make docker-build  # Criar imagem Docker
make docker-run    # Executar container Docker
```

## Estrutura de Diretórios

```
cobra-coral/
├── data/              # State persistence (criado automaticamente)
│   └── state.json    # Último timestamp de execução
├── bin/              # Binários compilados
├── cmd/worker/       # Entry point
├── internal/         # Código da aplicação
└── infra/            # Implementações
```

## Solução Rápida de Problemas

### Erro: "Permission denied" ao criar /app

**Solução**: Verifique seu arquivo `.env`:

```bash
# A linha STATE_FILE_PATH deve estar comentada ou usar caminho relativo
# STATE_FILE_PATH=./data/state.json
```

### Worker não executa

1. Verifique se o diretório `data` existe:
   ```bash
   mkdir -p data
   ```

2. Verifique se as credenciais do Telegram estão corretas no `.env`

3. Teste a conexão com o bot:
   - Abra o Telegram
   - Procure seu bot (@seu_bot_name)
   - Envie `/start`

### Rebuild completo

```bash
make clean
make build
make test
```

## Próximos Passos

Depois de testar localmente, você pode:

1. **Criar imagem Docker**:
   ```bash
   make docker-build
   ```

2. **Deploy no Google Cloud Run**:
   ```bash
   make gcp-build
   make gcp-deploy-job
   make gcp-schedule
   ```

## Observações Importantes

- O worker executa **uma única vez** por invocação
- Para execução contínua, use Cloud Scheduler (a cada 30 min, por exemplo)
- O arquivo `data/state.json` rastreia a última execução
- Logs são em formato JSON estruturado no stdout
- O worker só envia alertas se houver notícias novas com keywords relevantes
