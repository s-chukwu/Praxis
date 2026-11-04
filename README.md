# Praxis

AI workforce transition prototype.

## Services

| Service | Path | Port | Purpose |
|---|---|---|---|
| Frontend | `frontend/` | 3000 | Next.js web application |
| Backend API | `backend/` | 8080 | Go API; the only public service in production |
| AI / assessment | `AI_Assessment/service/` | 8000 | Python service; internal only |

## Run locally

```bash
cp .env.example .env
docker compose up --build
```

Then:

- http://localhost:3000 - frontend (shows backend health)
- http://localhost:8080/health - backend
- http://localhost:8000/health - AI service
- http://localhost:8000/docs - AI service OpenAPI docs

## Development without Docker

```bash
# backend
cd backend && go run .

# ai service
cd AI_Assessment/service && pip install -r requirements.txt && uvicorn main:app --reload

# frontend
cd frontend && npm install && npm run dev
```

## Deployment

Production runs on Azure Container Apps in `praxis-rg`. Merging to `main` builds and deploys all three services automatically through the Release workflow.

| Service | URL |
|---|---|
| Web | https://praxis-web.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| Backend | https://praxis-backend.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| AI | internal ingress, HMAC required |

Verify a deployment:

```bash
./scripts/smoke.sh
```

Deploy, rollback, and scaling steps are in `docs/runbook.md`. The system diagram is in `docs/architecture.md`.

## Docs

- `docs/scope.md` - scope freeze
- `docs/runbook.md` - deploy, rollback, and scaling
- `docs/architecture.md` - system and deploy diagrams
- `AI_Assessment/docs/rubric.md` - assessment rubric
- `AI_Assessment/docs/scenario.md` - scenario bank
- `AI_Assessment/docs/integration_contract.md` - Go ↔ Python contract
