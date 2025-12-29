# Deploy Guide - Google Cloud Run

## Pré-requisitos

1. **Google Cloud SDK instalado**
   ```bash
   gcloud version
   ```

2. **Projeto GCP configurado**
   ```bash
   gcloud config set project YOUR_PROJECT_ID
   ```

3. **APIs habilitadas**
   ```bash
   gcloud services enable cloudbuild.googleapis.com
   gcloud services enable run.googleapis.com
   gcloud services enable containerregistry.googleapis.com
   ```

## Método 1: Build com cloudbuild.yaml (Recomendado)

### Passo 0: Executar verificação pré-build (IMPORTANTE!)
```bash
cd /caminho/para/cobra-coral
./test-cloud-build.sh
```

Este script irá verificar:
- Se você está no diretório correto
- Se todos os arquivos necessários existem
- Se a estrutura de diretórios está correta
- Se o .gcloudignore não está excluindo arquivos importantes
- Se o gcloud está configurado corretamente

**Se todas as verificações passarem (✅), continue para o próximo passo.**

### Passo 1: Navegar para o diretório do projeto
```bash
cd /caminho/para/cobra-coral
pwd  # Deve mostrar: .../cobra-coral
```

### Passo 2: Verificar estrutura de diretórios
```bash
ls -la
# Deve mostrar: cmd/, infra/, internal/, Dockerfile, etc.
```

### Passo 3: Build e push para Container Registry
```bash
gcloud builds submit --config=cloudbuild.yaml
```

Isso irá:
1. **Verificar arquivos fonte** (novo step de verificação)
2. Construir a imagem Docker
3. Fazer tag com `$COMMIT_SHA` e `latest`
4. Fazer push para `gcr.io/YOUR_PROJECT_ID/weather-worker`

**IMPORTANTE**: O build agora inclui um step de verificação que mostra todos os arquivos antes de construir. Se o build falhar, procure no log a seção "Verifying source files before build..." para ver se todos os arquivos estão presentes.

### Passo 4: Configurar Persistência de Estado com Cloud Storage

**IMPORTANTE**: Antes de fazer o deploy do Job, você precisa configurar onde o estado será salvo.

#### 4.1. Criar bucket no Cloud Storage
```bash
# Criar bucket (use um nome único incluindo seu PROJECT_ID)
gsutil mb -l us-central1 gs://weather-worker-state-YOUR_PROJECT_ID

# Verificar que foi criado
gsutil ls | grep weather-worker-state
```

#### 4.2. Dar permissão para o Cloud Run Job acessar o bucket
```bash
# O Cloud Run usa a service account padrão do Compute Engine
# Substitua YOUR_PROJECT_ID pelo ID do seu projeto

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:YOUR_PROJECT_ID-compute@developer.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

**Exemplo com projeto real**:
```bash
gcloud projects add-iam-policy-binding wheater-forecast-482703 \
  --member="serviceAccount:wheater-forecast-482703-compute@developer.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

### Passo 5: Deploy no Cloud Run Jobs

Agora você pode criar o Job com as configurações de Cloud Storage:

```bash
gcloud run jobs create weather-worker \
  --image=gcr.io/YOUR_PROJECT_ID/weather-worker:latest \
  --region=us-central1 \
  --set-env-vars="BOT_TOKEN=seu_token,USER_ID=seu_user_id,STORAGE_TYPE=gcs,GCS_BUCKET=weather-worker-state-YOUR_PROJECT_ID,GCS_STATE_FILE=state.json" \
  --max-retries=1 \
  --task-timeout=10m
```

**Variáveis de ambiente explicadas**:
- `BOT_TOKEN`: Token do seu bot do Telegram
- `USER_ID`: Seu user ID do Telegram
- `STORAGE_TYPE=gcs`: Usa Cloud Storage (em vez de arquivo local)
- `GCS_BUCKET`: Nome do bucket criado no passo 4.1
- `GCS_STATE_FILE`: Nome do arquivo de estado no bucket (padrão: state.json)

**Alternativa - Storage local com arquivo** (não recomendado para produção):
```bash
# Sem persistência entre execuções (estado será perdido)
gcloud run jobs create weather-worker \
  --image=gcr.io/YOUR_PROJECT_ID/weather-worker:latest \
  --region=us-central1 \
  --set-env-vars="BOT_TOKEN=seu_token,USER_ID=seu_user_id,STORAGE_TYPE=file" \
  --max-retries=1 \
  --task-timeout=10m
```

### Passo 6: Executar manualmente (teste)
```bash
gcloud run jobs execute weather-worker --region=us-central1

# Verificar logs da execução
gcloud run jobs executions list --job=weather-worker --region=us-central1

# Ver logs detalhados (substitua EXECUTION_NAME pelo nome da execução)
gcloud logging read "resource.type=cloud_run_job AND resource.labels.job_name=weather-worker" --limit=50
```

**O que esperar nos logs**:
- `"message":"Using storage type: gcs"` - Confirma que está usando Cloud Storage
- `"message":"Using GCS storage: gs://weather-worker-state-YOUR_PROJECT_ID/state.json"` - Mostra o bucket
- `"message":"Reading state from gs://..."` - Leitura do estado
- `"message":"Saving state to gs://..."` - Salvando estado

**Verificar arquivo no bucket**:
```bash
# Listar arquivos no bucket
gsutil ls gs://weather-worker-state-YOUR_PROJECT_ID/

# Ver conteúdo do arquivo de estado
gsutil cat gs://weather-worker-state-YOUR_PROJECT_ID/state.json
```

Você deve ver algo como:
```json
{
  "last_execution_time": "2025-12-29T00:24:00Z"
}
```

### Passo 7: Agendar execução com Cloud Scheduler
```bash
gcloud scheduler jobs create http weather-check \
  --location=us-central1 \
  --schedule="*/30 * * * *" \
  --uri="https://us-central1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/YOUR_PROJECT_ID/jobs/weather-worker:run" \
  --http-method=POST \
  --oauth-service-account-email=YOUR_PROJECT_ID@appspot.gserviceaccount.com
```

## Método 2: Build direto (Simples)

### Se não quiser usar cloudbuild.yaml:

```bash
# Certifique-se de estar no diretório raiz do projeto
cd /caminho/para/cobra-coral

# Build e push
gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/weather-worker

# Deploy
gcloud run jobs create weather-worker \
  --image=gcr.io/YOUR_PROJECT_ID/weather-worker \
  --region=us-central1 \
  --set-env-vars="BOT_TOKEN=seu_token,USER_ID=seu_user_id"
```

## Troubleshooting

### Erro: "stat /build/cmd/worker: directory not found"

**Causa**: Arquivos fonte não foram enviados para o Cloud Build ou comando foi executado do diretório errado

**Solução passo a passo**:

1. **Execute o script de verificação pré-build**:
   ```bash
   cd /caminho/para/cobra-coral
   ./test-cloud-build.sh
   ```

   Se o script encontrar erros, corrija-os antes de continuar.

2. **Verifique que você está no diretório correto**:
   ```bash
   pwd  # Deve terminar em: /cobra-coral
   ls cmd/worker/main.go  # Deve existir
   ```

3. **Verifique o .gcloudignore**:
   ```bash
   cat .gcloudignore | grep "cmd/"
   # Não deve retornar nada (cmd/ NÃO deve estar sendo ignorado)
   ```

4. **Execute o build e procure a seção de verificação nos logs**:
   ```bash
   gcloud builds submit --config=cloudbuild.yaml
   ```

   Nos logs, procure por:
   ```
   =========================================
   Verifying source files before build...
   =========================================
   ```

   Esta seção mostrará se os arquivos `cmd/worker/main.go`, `go.mod` e `Dockerfile` estão presentes.

5. **Se os arquivos estiverem faltando no Cloud Build**:
   - Verifique se há um `.gcloudignore` excluindo arquivos importantes
   - Tente remover o `.gcloudignore` temporariamente
   - Use o método alternativo (build direto sem cloudbuild.yaml)

### Erro: "permission denied"

**Solução**:
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

### Ver logs do build

```bash
gcloud builds list --limit=5
gcloud builds log BUILD_ID
```

### Ver logs da execução

```bash
gcloud run jobs executions list --job=weather-worker --region=us-central1
gcloud run jobs executions describe EXECUTION_NAME --region=us-central1
```

## Persistência de Estado

O worker suporta duas opções de persistência de estado:

### Opção 1: Cloud Storage (Recomendado ✅ - Já Implementado)

**Vantagens**:
- Totalmente serverless
- Simples de configurar
- Custo muito baixo (~$0.026/GB/mês)
- Adequado para Cloud Run Jobs

**Como usar**: Veja os passos 4 e 5 da seção de deploy acima.

**Configuração**:
```bash
STORAGE_TYPE=gcs
GCS_BUCKET=weather-worker-state-YOUR_PROJECT_ID
GCS_STATE_FILE=state.json
```

### Opção 2: File Storage (Local/Desenvolvimento)

**Vantagens**:
- Simples para desenvolvimento local
- Sem dependências externas

**Desvantagens**:
- No Cloud Run Jobs, o estado é perdido entre execuções (sem persistência)
- Requer volume montado para persistência real

**Como usar**: Para desenvolvimento local, basta não definir `STORAGE_TYPE` ou usar `STORAGE_TYPE=file`.

**Configuração**:
```bash
STORAGE_TYPE=file
STATE_FILE_PATH=./data/state.json  # local dev
# ou
STATE_FILE_PATH=/app/data/state.json  # Docker
```

### Opção 3: Firestore (Não Implementado)

Se você precisar de queries complexas ou dados estruturados adicionais:

1. **Habilitar Firestore**
   ```bash
   gcloud firestore databases create --region=us-central1
   ```

2. **Implementar StateRepository usando Firestore**
   - Criar novo arquivo `infra/storage/firestore_state_repo.go`
   - Implementar interface `domain.StateRepository`
   - Adicionar lógica de escolha em `cmd/worker/main.go`

## Atualizar Deployment

```bash
# Rebuild
gcloud builds submit --config=cloudbuild.yaml

# Atualizar job
gcloud run jobs update weather-worker \
  --image=gcr.io/YOUR_PROJECT_ID/weather-worker:latest \
  --region=us-central1
```

## Monitoramento

```bash
# Ver execuções
gcloud run jobs executions list --job=weather-worker --region=us-central1

# Ver logs
gcloud logging read "resource.type=cloud_run_job AND resource.labels.job_name=weather-worker" --limit=50

# Metrics
gcloud monitoring time-series list \
  --filter='resource.type="cloud_run_job" AND resource.labels.job_name="weather-worker"'
```

## Custos Estimados

- **Cloud Run Jobs**: $0.00002400/vCPU-second, $0.00000250/GiB-second
- **Cloud Storage**: $0.020/GB/month
- **Cloud Scheduler**: $0.10/job/month

**Estimativa para execução a cada 30 min**:
- 48 execuções/dia
- ~5s por execução
- ~$1-2 USD/mês

## Segurança

### Armazenar secrets no Secret Manager

```bash
# Criar secrets
echo -n "seu_bot_token" | gcloud secrets create bot-token --data-file=-
echo -n "seu_user_id" | gcloud secrets create user-id --data-file=-

# Dar permissão ao Cloud Run
gcloud secrets add-iam-policy-binding bot-token \
  --member="serviceAccount:YOUR_PROJECT_ID-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Usar no deployment
gcloud run jobs update weather-worker \
  --set-secrets="BOT_TOKEN=bot-token:latest,USER_ID=user-id:latest" \
  --region=us-central1
```
