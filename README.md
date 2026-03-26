# Phortizo

GitHub source handler for [katastroma](https://github.com/katastroma).

Receives GitHub webhooks, verifies HMAC signatures, matches push events against
registered watch targets, clones the source, detects the renderer type, and
streams the source to the appropriate renderer.

Implements the [naukleros](https://github.com/katastroma/naukleros)
`RetrieverService` for replay support.

## Pipeline

1. Receive webhook at `POST /webhook/{namespace}`
2. Look up registration by ID
3. Verify `X-Hub-Signature-256` against stored secret
4. Parse push event and match against watch targets
5. Resolve tenant credentials
6. Clone repository (shallow, in-memory)
7. Detect renderer type (Helm, Kustomize, or raw YAML)
8. Stream source to the renderer via
   [keleustēs](https://github.com/katastroma/keleustes) gRPC

## Credentials

- **Platform App installation** — tenants install the platform GitHub App and
  provide their installation ID
- **Tenant GitHub App** — tenant provides their own App credentials (client ID,
  private key, installation ID)
- **Token** — GitHub personal or fine-grained token
- **Basic auth** — username and password
- **SSH** — SSH private key

## Environment Variables

### Required

| Variable                     | Description                                          |
| ---------------------------- | ---------------------------------------------------- |
| `GITHUB_APP_CLIENT_ID`       | Platform GitHub App client ID.                       |
| `GITHUB_APP_INSTALLATION_ID` | Platform GitHub App installation ID.                 |
| `GITHUB_APP_PRIVATE_KEY`     | PEM-encoded private key for the platform GitHub App. |

### Renderers

| Variable             | Description                                             |
| -------------------- | ------------------------------------------------------- |
| `RENDERER_HELM`      | Service address of the Helm renderer (e.g., `orpheus`). |
| `RENDERER_KUSTOMIZE` | Service address of the Kustomize renderer.              |
| `RENDERER_RAW`       | Service address of the raw YAML renderer.               |

### Optional

| Variable          | Default | Description                                                     |
| ----------------- | ------- | --------------------------------------------------------------- |
| `SERVICE_VERSION` | `dev`   | Service version reported to the OTel resource. Set by CI.       |
| `PORT`            | `8080`  | HTTP server port. Serves `/healthz` and `/webhook/{namespace}`. |
| `TEMPO_ADDRESS`   |         | gRPC address of Tempo. Enables replay via trace queries.        |

### OTel (from grpc-foundation)

| Variable                      | Default          | Description                                 |
| ----------------------------- | ---------------- | ------------------------------------------- |
| `OTEL_TRACES_EXPORTER`        | `none`           | Trace exporter: `otlp` or `none`.           |
| `OTEL_METRICS_EXPORTER`       | `none`           | Metrics exporter: `otlp` or `none`.         |
| `OTEL_LOGS_EXPORTER`          | `none`           | Log exporter: `otlp`, `console`, or `none`. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTLP collector endpoint.                    |

### gRPC Server (from grpc-foundation)

| Variable              | Default  | Description                                                   |
| --------------------- | -------- | ------------------------------------------------------------- |
| `GRPC_SERVER_ADDRESS` | `:50051` | gRPC server listen address. Health and RetrieverService here. |

### Logging (from grpc-foundation)

| Variable     | Default      | Description                                         |
| ------------ | ------------ | --------------------------------------------------- |
| `LOG_FORMAT` | `structured` | Log output format: `json`, `text`, or `structured`. |

## Span Conventions

The `pipeline.run` span carries these attributes for observability and replay:

| Attribute               | Description                                                          |
| ----------------------- | -------------------------------------------------------------------- |
| `namespace`             | Webhook registration ID. Used by replay to look up the registration. |
| `tenant`                | Tenant identity.                                                     |
| `watch_target.repo_url` | Repository clone URL.                                                |
| `watch_target.ref`      | Git ref (e.g., `refs/heads/main`).                                   |
| `watch_target.path`     | Path within the repository being watched.                            |

The `webhook.dispatch` span carries:

| Attribute            | Description                   |
| -------------------- | ----------------------------- |
| `github.delivery_id` | GitHub webhook delivery GUID. |
