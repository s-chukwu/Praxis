# Architecture

```mermaid
flowchart LR
  user[Browser] --> web["Web<br/>Next.js 14<br/>public"]
  web -->|"server-side fetch"| backend["Backend API<br/>Go<br/>public"]
  backend -->|"HMAC signed"| ai["AI service<br/>FastAPI<br/>internal ingress"]
  backend --> db[("Supabase<br/>Postgres + Auth")]
  ai --> kv[("Azure Key Vault<br/>assessor keys")]
  ai --> llm["Gemini primary<br/>DeepSeek fallback"]
```

## Deploy flow

```mermaid
flowchart LR
  dev[Pull request] --> ci["CI<br/>go, python, web"]
  ci --> merge[Merge to main]
  merge --> release["Release<br/>build 3 images"]
  release --> ghcr[("GHCR<br/>SHA tags")]
  release -->|"OIDC"| deploy["Deploy job"]
  deploy --> aca["Azure Container Apps<br/>praxis-env"]
  ghcr --> aca
```

## Components

| Component | Runtime | Exposure | Notes |
|---|---|---|---|
| Web | Next.js 14, Node 20 | Public | Renders backend health server-side |
| Backend | Go 1.23 | Public | Signs every AI call with HMAC |
| AI | FastAPI, Python 3.12 | Internal ingress | Rejects unsigned `/internal` calls with 401 |
| Database | Supabase | Private | Go owns reads and writes |
| Secrets | Azure Key Vault | Private | Read by `praxis-ai` managed identity |

## Constraints

- The Container Apps environment is **express**. It has no VNet and does not support Key Vault secret references in app config, so the AI service reads Key Vault directly with its managed identity.
- Because express has no VNet, the AI service is reachable from the internet even with internal ingress. HMAC signing is the compensating control.
- Images are tagged with the full commit SHA and deployed automatically on merge to `main`.

See `docs/runbook.md` for deploy, rollback, and scaling commands.
