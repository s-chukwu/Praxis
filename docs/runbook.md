# Operations runbook

This covers deploys, rollback, scaling, and checks for the Azure Container Apps deployment.

## Production URLs

| Service | URL |
|---|---|
| Web | https://praxis-web.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| Backend | https://praxis-backend.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |
| AI | https://praxis-ai.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io |

The AI service is configured with internal ingress. The environment is an express environment, so that is not network-isolated, and the HMAC signature requirement is what protects it.

## Deploy

Deploys are automatic. Merging to `main` runs the Release workflow, which:

1. Builds `backend`, `ai`, and `web` images and pushes them to GHCR tagged with the commit SHA and `latest`.
2. Logs in to Azure with OIDC using the `praxis-deployer` app registration.
3. Updates all three container apps to the new SHA.

No manual step is needed. Watch it under the repo's **Actions** tab.

## Manual deploy

Use this when you need to deploy a specific commit without merging.

```bash
RG=praxis-rg
SHA=<full-commit-sha>
REPO=s-chukwu/praxis

for app in ai backend web; do
  az containerapp update -g "$RG" -n "praxis-$app" \
    --image "ghcr.io/$REPO/$app:$SHA" --output none
done
```

## Rollback

Every merge produces a SHA-tagged image, so rollback is deploying the previous tag.

```bash
RG=praxis-rg
REPO=s-chukwu/praxis

git log --oneline -5 origin/main
SHA=<previous-commit-sha>

for app in ai backend web; do
  az containerapp update -g "$RG" -n "praxis-$app" \
    --image "ghcr.io/$REPO/$app:$SHA" --output none
done

./scripts/smoke.sh
```

Rehearsed 4 Nov 2026: rollback to `efbc064` took 56 s, restore to `baa7ceb` took 51 s, and the smoke test passed both ways. Rollback target is under 10 minutes.

## Scale for a demo

Keep replicas warm so the first request is not a cold start.

```bash
RG=praxis-rg
for app in backend web ai; do
  az containerapp update -g "$RG" -n "praxis-$app" --min-replicas 1 --output none
done
```

After the demo, return to scale to zero to stay inside the student grant:

```bash
for app in backend web ai; do
  az containerapp update -g "$RG" -n "praxis-$app" --min-replicas 0 --output none
done
```

## Demo day checklist

Before:

- Scale web, backend, and ai to `min-replicas 1`.
- Run `./scripts/smoke.sh`.
- Open every page you plan to show once, so nothing is cold.
- Confirm the deployed SHA equals `main`.
- Keep the Azure revisions page and Log Analytics open in separate tabs.

During:

- If a request fails, retry once, then check the revision list and the logs.
- Do not deploy anything during the demo.

After:

- Scale all three apps back to 0.
- Note anything that failed while it is fresh.

## Cold start

Measured 9 Nov 2026 from a scaled-to-zero state, with the web app on Next.js 16:

| Service | Cold | Warm |
|---|---|---|
| Backend | 1.8 s | 0.8 s |
| Web | 5.4 s | 0.4 s |
| AI | 6.3 s | 0.8 s |

The AI service is the slowest because it loads the rubric and scenario content at startup. The web app got slower on cold start after the Next.js 16 upgrade. Keep replicas warm for demos with the commands above.

## Verify

```bash
./scripts/smoke.sh
```

Override targets when testing a different environment:

```bash
BACKEND_URL=http://localhost:8080 WEB_URL=http://localhost:3000 AI_URL=http://localhost:8000 ./scripts/smoke.sh
```

## Logs and revisions

```bash
RG=praxis-rg

az containerapp revision list -g "$RG" -n praxis-backend -o table
az containerapp logs show -g "$RG" -n praxis-backend --tail 100
az containerapp logs show -g "$RG" -n praxis-ai --tail 100
```

## Secrets

List secret names, never values:

```bash
az containerapp secret list -g praxis-rg -n praxis-ai
az containerapp secret list -g praxis-rg -n praxis-backend
```

Rotating `AI_SERVICE_SHARED_SECRET` requires setting the same value on both apps, then restarting revisions.

## Assessor keys

The assessor keys live in Azure Key Vault in `praxis-rg`, read by `praxis-ai` through the `praxis-ai-identity` managed identity. To find the vault:

```bash
az keyvault list -g praxis-rg -o table
```

Production reads two secrets from Key Vault: `assessor-keys` and `capstone-held-out`. Both are JSON objects keyed by version. Local dev reads the copies under `AI_Assessment/tests/fixtures` through `PRAXIS_PROTECTED_CONTENT_DIR`, mounted read-only by `docker-compose.yml`.

### Rotating keys

Add a new version next to the existing ones. Never replace or remove a version that past attempts reference.

```bash
VAULT=praxis-kv-26c0
SECRET=assessor-keys

az keyvault secret show --vault-name "$VAULT" --name "$SECRET" --query value -o tsv > current.json

python3 - <<'PY'
import json
current = json.load(open("current.json"))
new = json.load(open("assessor-keys-1.3.0.json"))
current[new["version"]] = new
json.dump(current, open("merged.json", "w"), separators=(",", ":"))
PY

az keyvault secret set --vault-name "$VAULT" --name "$SECRET" --file merged.json --output none
```

Then deploy the matching content version, and check the AI startup log shows all versions:

```
"protected_content_loaded" ... "assessor_keys": ["1.0.0", "1.1.0", "1.2.0", "1.3.0"]
```

After the new version is live, delete the old image versions so the old keys are not baked into any public image. Keep the current SHA and `latest`. Do not commit key material to the repo.

## Environment variables

Names only. Values are set in Azure, or in `.env` for local runs.

| Service | Variable | Source |
|---|---|---|
| Backend | `SUPABASE_URL` | plain env |
| Backend | `SUPABASE_ANON_KEY` | plain env |
| Backend | `SUPABASE_SERVICE_ROLE_KEY` | Container Apps secret `supabase-service-role-key` |
| Backend | `AI_SERVICE_URL` | plain env |
| Backend | `AI_SERVICE_SHARED_SECRET` | Container Apps secret `ai-shared-secret` |
| Backend | `CORS_ALLOWED_ORIGIN` | plain env |
| AI | `AI_SERVICE_SHARED_SECRET` | Container Apps secret |
| AI | `PRAXIS_ENV` | `production` |
| AI | `AZURE_KEY_VAULT_URL` | plain env |
| AI | `ASSESSOR_KEYS_SECRET_NAME` | `assessor-keys` |
| AI | `CAPSTONE_HELD_OUT_SECRET_NAME` | `capstone-held-out` |
| AI | `AZURE_CLIENT_ID` | managed identity client id |
| Web | `API_INTERNAL_URL` | plain env |
| Web | `NEXT_PUBLIC_API_URL` | plain env |

## Known constraints

- The Container Apps environment is **express**. It has no VNet, and it does not support Key Vault secret references in Container Apps config.
- Because of express, internal ingress is reachable from the internet. All `/internal` routes require an HMAC signature; unsigned calls return 401.
- `praxis-ai` and `praxis-backend` run with `min-replicas 0`, so the first request after idle is a cold start.
