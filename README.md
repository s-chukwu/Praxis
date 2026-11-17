# Praxis

Praxis is an AI workforce transition prototype for AI-assisted customer support. Candidates complete a real support task, a person reviews the work, and the candidate gets a learning plan for the skills that still need practice.

## Services

| Service | Path | Port | Role |
|---|---|---|---|
| Web | `frontend/` | 3000 | Next.js app for candidates, reviewers, and admins |
| API | `backend/` | 8080 | Go API. Public API; signs internal assessment calls with HMAC |
| Assessment | `AI_Assessment/service/` | 8000 | FastAPI service for assessment, scoring, learning plans, and catalog data. `/internal` routes require HMAC |

## Run locally

```bash
cp .env.example .env
docker compose up --build
```

The compose stack starts all three services. It is enough to inspect the UI and check health. Sign-in needs Supabase. Server-side learner state, attempts, and reviews need Postgres.

- http://localhost:3000: web app
- http://localhost:8080/health: API health
- http://localhost:8000/health: assessment health
- http://localhost:8000/docs: assessment OpenAPI docs

The web app runs in demo mode unless it is built with `NEXT_PUBLIC_API_MODE=live`. Demo mode keeps learner state in the browser. Sign-in still uses Supabase.

To run the full live flow locally, set these values in `.env`, add the backend values to the `backend` service environment, and pass the web values as build args for the `web` service:

- `backend`: `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`, `DATABASE_URL`, `AI_SERVICE_SHARED_SECRET`
- `web` build: `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY`, `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_API_MODE=live`

## Development without Docker

Backend:

```bash
set -a
. ./.env
set +a
cd backend
go run .
```

Assessment:

```bash
cd AI_Assessment/service
python -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --reload
```

Set `AI_SERVICE_SHARED_SECRET` to call `/internal` routes. Without it, the service starts but internal routes reject every call.

Web:

```bash
cd frontend
npm install
npm run dev
```

Set `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY`, `NEXT_PUBLIC_API_URL`, and `NEXT_PUBLIC_API_MODE=live` in `frontend/.env.local` to use the API.

## Deployment

Production runs on Azure Container Apps in `praxis-rg`. A push to `main` runs the Release workflow, which builds and deploys the API, assessment, and web images.

| Service | URL |
|---|---|
| Web | https://praxis-web.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| API | https://praxis-backend.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| Assessment | https://praxis-ai.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |

The assessment service uses external ingress because the current Container Apps environment does not provide internal-only networking. Unsigned `/internal` calls return 401. The HMAC secret is the control.

Verify a deployment:

```bash
./scripts/smoke.sh
```

Deploy, rollback, and scaling steps are in `docs/runbook.md`.

## Docs

- `docs/scope.md`: scope freeze
- `docs/runbook.md`: deploy, rollback, and scaling
- `docs/architecture.md`: system and deploy diagrams
- `docs/demo-script.md`: demo runbook
- `docs/erd.md`: data model
- `AI_Assessment/docs/integration_contract.md`: Go to Python contract
- `AI_Assessment/docs/rubric.md`: assessment rubric
- `AI_Assessment/docs/scenario.md`: scenario bank
