# Phortizo

GitHub source handler for [katastroma](https://github.com/katastroma).

Receives GitHub webhooks, verifies HMAC signatures, matches push events against
registered source targets, clones the source, detects the renderer type, and
streams the source to the appropriate renderer.

Implements the [naukleros](https://github.com/katastroma/naukleros)
`RetrieverService` for replay support.

## Pipeline

### Retrieve

1. Resolve tenant namespace and source target ID from request
2. Resolve source target from namespace and source target ID
3. Start event (type `manual`) and process the source target

### Webhook

1. Receive webhook at `POST /webhook/{namespace}`
2. Acquire stored secret from tenant namespace
3. Validate payload and verify signature against stored secret
4. Parse push event and list source targets
5. Match push event paths against source targets
6. If no matches, skip
7. For each source target, process the source target

### Replay

1. Acquire source target attributes from requested replay event ID.
2. Resolve source targets from source target attributes
3. For each source target, check if lease is active, skip if so
4. For each source target, process the source target

## For Each Source Target

1. Acquire
   [source target lease](https://katastroma.github.io/docs/event-driven#source-target-leasing)
2. Resolve tenant credentials
3. Clone repository (shallow, in-memory)
4. Detect renderer type (Helm, Kustomize, or raw YAML)
5. Verify processing instance still holds the lease
6. Stream source to the renderer via
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

| Variable                     | Description                                                                  |
| ---------------------------- | ---------------------------------------------------------------------------- |
| `GITHUB_APP_CLIENT_ID`       | Platform GitHub App client ID.                                               |
| `GITHUB_APP_INSTALLATION_ID` | Platform GitHub App installation ID. Required for health checks.             |
| `GITHUB_APP_PRIVATE_KEY`     | PEM-encoded private key for the platform GitHub App.                         |
| `TEMPO_ADDRESS`              | gRPC address of Tempo. Used for source target reconstruction during replays. |

### Renderers

| Variable             | Description                                             |
| -------------------- | ------------------------------------------------------- |
| `RENDERER_HELM`      | Service address of the Helm renderer (e.g., `orpheus`). |
| `RENDERER_KUSTOMIZE` | Service address of the Kustomize renderer.              |
| `RENDERER_RAW`       | Service address of the raw YAML renderer.               |

### Optional

| Variable              | Default | Description                                                           |
| --------------------- | ------- | --------------------------------------------------------------------- |
| `SERVICE_VERSION`     | `dev`   | Service version reported to the OTel resource.                        |
| `PORT`                | `8080`  | HTTP server port. Serves `/healthz` and `/webhook/{namespace}`.       |
| `LEASE_STALE_AFTER`   | `10m`   | Duration after which a source target lease is considered stale.       |
| `MAX_REPLAY_ATTEMPTS` | `3`     | Maximum number of replay attempts per event before permanent failure. |

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

The root event span carries:

| Attribute            | Description                                                |
| -------------------- | ---------------------------------------------------------- |
| `tenant`             | Tenant namespace. Identifies the tenant.                   |
| `event.type`         | Event type: `webhook`, `manual`, or `replay`.              |
| `github.delivery_id` | GitHub webhook delivery GUID (webhook events only).        |
| `github.head_commit` | Head commit SHA from the push event (webhook events only). |

The `watch_target` span (child of the event span) carries:

| Attribute               | Description                               |
| ----------------------- | ----------------------------------------- |
| `watch_target.name`     | ConfigMap name of the source target.      |
| `watch_target.repo_url` | Repository clone URL.                     |
| `watch_target.ref`      | Git ref (e.g., `refs/heads/main`).        |
| `watch_target.path`     | Path within the repository being watched. |
