# Phortizo

Katastroma's GitHub API for [katastroma](https://github.com/katastroma).

- Receives GitHub webhooks, verifies their signatures, matches source events
  against registered watch targets, and fetches source content.
- Manages watch targets and credentials on behalf of tenants.

## Operations

- **Register** — full lifecycle management of watch targets, credentials, and
  webhook secrets.
  - Tenants create, update, query, and remove these resources through the API.
- **Verify** — validates incoming webhook HMAC signatures against stored
  secrets. Verification identifies the tenant.
- **Match** — given a source event, find the matching watch target. Returns the
  watch target if matched, empty if not.
- **Fetch** — given a watch target, authenticate and fetch the source content.
- **Publish** — emits an event notifying source is ready for rendering.

## Watch Target

A watch target represents a tracked location within a GitHub repository — a
repository URL, a ref, and a path. The API matches incoming push events against
registered watch targets to determine if a pipeline run is needed.

## Credentials

Credentials are how the API authenticates to fetch from a repository.

- **Platform App installation** (preferred) — tenants install the phortizo
  GitHub App in their org and provide only their installation ID. No
  tenant-side secrets to store.
- **Tenant GitHub App** — the tenant provides their own GitHub App credentials
  (client ID, private key, and installation ID).
- **Token** — any valid GitHub token.
- **Basic auth** — username and password.
- **SSH** — SSH key for git transport.

## Webhook Secret

A webhook secret is a shared secret between the API and GitHub. The API
generates it, the tenant configures it on their GitHub repository webhook, and
GitHub signs each webhook payload with it. The API verifies the HMAC signature
on incoming payloads to authenticate the event and identify which tenant it
belongs to.

## Multiple Tenants

Multiple webhooks can be added to a single repository, allowing multiple tenants
to send events for a single repo.
