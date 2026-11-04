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

## Cold start

Measured 4 Nov 2026 from a scaled-to-zero state:

| Service | Cold | Warm |
|---|---|---|
| Backend | 1.8 s | 0.8 s |
| Web | 2.9 s | 1.4 s |
| AI | 6.3 s | 0.8 s |

The AI service is the slowest because it loads the rubric and scenario content at startup. Keep replicas warm for demos with the commands above.

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

Do not commit key material to the repo.

## Known constraints

- The Container Apps environment is **express**. It has no VNet, and it does not support Key Vault secret references in Container Apps config.
- Because of express, internal ingress is reachable from the internet. All `/internal` routes require an HMAC signature; unsigned calls return 401.
- `praxis-ai` and `praxis-backend` run with `min-replicas 0`, so the first request after idle is a cold start.
