# Go Backend Boilerplate

Este projeto e um backend base em Go para APIs SaaS.

## O que ja vem pronto

- API HTTP com Echo
- Autenticacao com JWT (access + refresh)
- Login social opcional via Firebase
- Camadas separadas: handler, service, repository
- Middleware de seguranca (CORS, CSRF, rate limit, recovery)
- Logs estruturados e health check
- Jobs assincronos com Asynq + Redis
- Migrations com golang-migrate

## Estrutura principal

- `cmd/`: ponto de entrada da aplicacao
- `internal/config/`: configuracao por variaveis de ambiente
- `internal/router/`: registro de rotas
- `internal/handler/`: camada HTTP
- `internal/service/`: regras de aplicacao
- `internal/repository/`: acesso a dados
- `internal/model/`: modelos e DTOs
- `internal/middleware/`: middleware HTTP
- `internal/database/`: conexao e migrations
- `internal/lib/utils/job/`: fila de tarefas

## Pre-requisitos

- Go 1.26+
- MySQL 8+
- Redis 7+

## Subindo dependencias com Docker

```bash
cd backend
docker compose up -d
```

## Configuracao local

```bash
cd backend
cp .env.example .env
```

Ajuste no `.env` (minimo):

- `APP_DATABASE_HOST`
- `APP_DATABASE_PORT`
- `APP_DATABASE_USER`
- `APP_DATABASE_PASSWORD`
- `APP_DATABASE_DB_NAME`
- `APP_REDIS_ADDRESS`
- `APP_AUTH_SECRET_KEY`

## Rodando a aplicacao

```bash
cd backend
go run ./cmd
```

## Rotas base

- `GET /health`
- `GET /api/v1/csrf-token`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `GET /api/v1/users/settings`
- `PUT /api/v1/users/settings`
- `GET /api/v1/notifications`

## Taskfile

Automacoes disponiveis em [`backend/Taskfile.yml`](/c:/Users/gusta/Desktop/Laboratorio/Cashing-go/backend/Taskfile.yml):

- `task run`: sobe a API (se `APP_PRIMARY_ENV != local`, o `main` aplica migrations antes de iniciar)
- `task migrations:up`: aplica migrations
- `task migrations:down`: rollback da ultima migration
- `task migrations:status`: status das migrations
- `task tidy`: formata e organiza dependencias

### Como o `task run` se comporta com migration

O `task run` executa `go run ./cmd`. No `main`, a regra e:

- se `APP_PRIMARY_ENV=local`: **nao** aplica migration automaticamente
- se `APP_PRIMARY_ENV` for diferente de `local` (ex.: `dev`, `staging`, `prod`): aplica migration no startup

Para ambiente local, o fluxo recomendado e:

1. `task migrations:up`
2. `task run`

## Migration IaC (Atlas)

O fluxo de schema-as-code para banco esta em [`backend/atlas.hcl`](/c:/Users/gusta/Desktop/Laboratorio/Cashing-go/backend/atlas.hcl).

Comandos uteis:

- `task migrations:new name=minha_migration`
- `task migrations:validate`
- `task schema:inspect`

## Terraform vs Atlas

Essa diferenca costuma gerar duvida, entao aqui vai o resumo pratico:

- **Atlas**: gerencia **schema e migration de banco de dados** (tabelas, colunas, indices, diffs SQL).
- **Terraform**: gerencia **infraestrutura de cloud** (rede, VPC, RDS/Cloud SQL, Redis, bucket, IAM, etc.).

### Quando usar Atlas

Use Atlas quando voce precisar:

- criar/alterar tabelas e colunas
- versionar SQL de migration
- validar consistencia do historico de migration
- comparar schema desejado vs schema atual

### Quando usar Terraform

Use Terraform quando voce precisar:

- provisionar banco/redis/servidores na nuvem
- configurar seguranca e rede
- padronizar ambientes (dev/staging/prod)
- reproducao de infraestrutura via codigo

### Como eles trabalham juntos

Fluxo comum em times:

1. Terraform provisiona o banco gerenciado (ex.: RDS/MySQL)
2. Aplicacao sobe
3. Atlas/migrate aplica migrations no banco provisionado

Em resumo: **Terraform cria o banco; Atlas modela o banco**.

## Migrations

As migrations ficam em `internal/database/migrations`.

Para executar automaticamente no start, use `APP_PRIMARY_ENV` diferente de `local`.

## Como estender

1. Crie um modelo em `internal/model`.
2. Crie o repositorio em `internal/repository`.
3. Crie o servico em `internal/service`.
4. Crie o handler em `internal/handler`.
5. Registre as rotas em `internal/router/router.go`.
6. Adicione migration para as tabelas novas.
