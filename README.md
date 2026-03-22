# Phortizo

Katastroma's git retriever. Standalone
[naukleros](https://github.com/katastroma/naukleros) implementation.

Phortizo receives git webhooks, manages gitops identities and repository
credentials, matches source events against registered identities, and fetches
source content before forwarding it to the remaining platform pipeline pieces
(orpheus, orderer, then katartismos implementations).

## Operations

- **Match** — given a source event, find the matching gitops identity. Returns
  the identity if matched, empty if not.
- **Fetch** — given a source identity, authenticate and fetch the source
  content.

## Resource Labeling

All resources naukleros implementations create are labeled with the labels it
receives.

## GitHub

Multiple Webhooks can be added to a single repository, allowing multiple tenants
to send events for a single repo.
