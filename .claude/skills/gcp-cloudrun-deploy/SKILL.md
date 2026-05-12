---
name: gcp-cloudrun-deploy
description: Use when containerizing a service for Cloud Run, setting up a Cloud Build CI/CD pipeline from GitHub, wiring IAM auth for a CLI caller, or putting Cloudflare in front of a Cloud Run service.
---

# GCP Cloud Run + Cloud Build + Cloudflare Deployment

## PORT env var pattern (Cloud Run compatibility)

Cloud Run injects `PORT` (default `8080`) and requires binding to all interfaces, not `localhost`. Read `PORT` at startup and override any flag default:

```go
// Go example — adapt for other runtimes
effectiveAddr := *addr  // flag default: "localhost:8080"
if p := os.Getenv("PORT"); p != "" {
    effectiveAddr = ":" + p  // bind all interfaces on Cloud Run
}
```

## Dockerfile — multi-stage (Go → distroless)

Build context: the module directory (where `go.mod` lives), set via `dir:` in cloudbuild.yaml.

```dockerfile
# syntax=docker/dockerfile:1.7

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/<binary> ./<cmd-path>

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/<binary> /<binary>
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/<binary>"]
```

**Notes**:
- Pure-Go services: `CGO_ENABLED=0` + `distroless/static` (no libc). If CGO is needed, use `distroless/base` instead.
- No `CMD ["--addr=..."]` needed when using the PORT env var pattern.
- Alpine tags may lag for bleeding-edge Go versions; default to the Debian base image.

`.dockerignore` alongside the Dockerfile:
```
**/*_test.go
.git
.gitignore
README*
```

## cloudbuild.yaml structure

Place at the **repo root** so GitHub triggers find it without a custom path. Build context is scoped to the module subdirectory via `dir:`.

```yaml
substitutions:
  _REGION: us-central1
  _AR_REPO: <artifact-registry-repo-name>
  _SERVICE: <cloud-run-service-name>
  _RUNTIME_SA: <service>-run@${PROJECT_ID}.iam.gserviceaccount.com

options:
  logging: CLOUD_LOGGING_ONLY   # required when no GCS log bucket is specified

steps:
  - id: build
    name: gcr.io/cloud-builders/docker
    dir: <path/to/module>       # sets Docker build context to the module dir
    args:
      - build
      - -t
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}:${SHORT_SHA}
      - -t
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}:latest
      - .

  - id: push
    name: gcr.io/cloud-builders/docker
    args:
      - push
      - --all-tags
      - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}

  - id: deploy
    name: gcr.io/google.com/cloudsdktool/cloud-sdk:slim
    entrypoint: gcloud
    args:
      - run
      - deploy
      - ${_SERVICE}
      - --image=${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}:${SHORT_SHA}
      - --region=${_REGION}
      - --platform=managed
      - --no-allow-unauthenticated   # IAM gate — do not remove
      - --ingress=all
      - --port=8080
      - --use-http2                  # required for gRPC/Connect end-to-end HTTP/2
      - --service-account=${_RUNTIME_SA}
      - --cpu=1
      - --memory=256Mi
      - --min-instances=0
      - --max-instances=3

images:
  - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}:${SHORT_SHA}
  - ${_REGION}-docker.pkg.dev/${PROJECT_ID}/${_AR_REPO}/${_SERVICE}:latest
```

**Key flags**:
- `--use-http2`: Cloud Run forwards HTTP/2 to the container. Required for gRPC/Connect; omitting it silently downgrades to HTTP/1.1.
- `--no-allow-unauthenticated`: enforces IAM on every request — the `*.run.app` URL stays public-addressable but IAM blocks unauthorized callers.
- `${SHORT_SHA}`: auto-populated by GitHub triggers; gives immutable tags for clean rollbacks.
- Override substitution variables at trigger creation time, not in the file.

## IAM vs Cloudflare — roles and responsibilities

| Layer | What it does | What it does NOT do |
|---|---|---|
| **GCP IAM** (`--no-allow-unauthenticated`) | Authoritative auth gate on Cloud Run. Blocks any caller without `roles/run.invoker` + valid ID token — including direct `*.run.app` hits that bypass Cloudflare. | Replace Cloudflare |
| **Cloudflare** (DNS/TLS/WAF) | Edge: custom domain, TLS termination, DDoS/WAF, rate limiting. | Replace IAM. Anyone who finds the `*.run.app` URL can still hit it — IAM is still required. |

**Bottom line**: always keep `--no-allow-unauthenticated`. Cloudflare is an additive edge layer, not an auth replacement.

For ingress restriction (optional, advanced):
- `--ingress=all` (default): `*.run.app` reachable; IAM gates it. Simplest.
- `--ingress=internal-and-cloud-load-balancing`: requires a GCLB + serverless NEG; `*.run.app` becomes unreachable. Use when you need Cloudflare as the exclusive entry point.

## CLI auth — Application Default Credentials

The CLI should not embed service account keys. Instead:
1. User runs `gcloud auth application-default login` once on their machine.
2. CLI uses ADC at runtime to mint a Google ID token for the Cloud Run service URL.
3. Send the token as `Authorization: Bearer <id-token>`.

Grant the CLI user invoker access:
```bash
gcloud run services add-iam-policy-binding <service> \
  --region=<region> \
  --member=user:<email> \
  --role=roles/run.invoker
```

## One-time GCP setup

```bash
export PROJECT_ID=<your-project>
export REGION=us-central1
export SERVICE=<service-name>
export AR_REPO=<ar-repo-name>

gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  iamcredentials.googleapis.com

# Artifact Registry repo
gcloud artifacts repositories create $AR_REPO \
  --repository-format=docker --location=$REGION

# Dedicated runtime SA for the Cloud Run service
gcloud iam service-accounts create ${SERVICE}-run \
  --display-name="Cloud Run runtime SA for $SERVICE"

# Grant Cloud Build SA the roles it needs
PROJECT_NUM=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
CB_SA=${PROJECT_NUM}@cloudbuild.gserviceaccount.com
for role in roles/run.admin roles/artifactregistry.writer roles/iam.serviceAccountUser; do
  gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:$CB_SA --role=$role
done
```

Connect GitHub via **Cloud Build → Triggers → Connect repository** in the GCP console (2nd gen). Create a trigger pointing at `/cloudbuild.yaml`, push-to-branch event.

## Cloudflare DNS wiring

1. Deploy the service once to get its `*.run.app` URL.
2. In GCP console: Cloud Run → Manage Custom Domains → add domain mapping for `api.yourdomain.com`.
3. In Cloudflare: add the CNAME GCP provides (`ghs.googlehosted.com`).
   - Start with **DNS only (gray cloud)** for GCP domain ownership validation.
   - After validation passes, switch to **proxied (orange cloud)**, SSL mode = **Full (strict)**.
4. IAM is enforced on both the `*.run.app` URL and the custom domain — callers always need an ID token.

## Verification checklist

```bash
# 1. Build succeeded
gcloud builds list --limit=1

# 2. Image tagged correctly
gcloud artifacts docker images list $REGION-docker.pkg.dev/$PROJECT_ID/$AR_REPO/$SERVICE

# 3. Service is live
gcloud run services describe $SERVICE --region=$REGION --format='value(status.url)'

# 4. IAM blocks unauthenticated (expect 403)
curl -i -X POST https://<run-url>/minddaemon.v1.HelloService/SayHello \
  -H 'content-type: application/json' -d '{"name":"ci"}'

# 5. IAM allows authenticated (expect 200)
TOKEN=$(gcloud auth print-identity-token --audiences=https://<run-url>)
curl -i -X POST https://<run-url>/minddaemon.v1.HelloService/SayHello \
  -H "authorization: Bearer $TOKEN" \
  -H 'content-type: application/json' -d '{"name":"ci"}'
```
