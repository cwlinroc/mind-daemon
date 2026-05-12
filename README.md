# Mind Daemon

a new cli tool with server agent running in the background.
a learning and testing purpose project, and also fit my needs for a personal assistant.

## Local development

### Requirements

- .NET SDK 10 or newer, for `cli/MindDaemonCli`.
- Go 1.26.2 or compatible, for `server.dmz/mind_daemon_dmz`.
- Buf CLI, for regenerating Go protobuf/ConnectRPC code after editing files in `proto`.
- Protobuf generation plugins for Go if using local `protoc` workflows instead of Buf:
  - `protoc`
  - `protoc-gen-go`
  - `protoc-gen-connect-go`

Notes:

- Normal CLI builds do not require `buf` or `protoc`; `Grpc.Tools` generates the .NET client during `dotnet build`.
- Generated Go stubs are committed under `server.dmz/mind_daemon_dmz/gen`, so a fresh checkout can build the server without regenerating protobuf code.
- Install `buf` before changing `.proto` files. In this session, `buf` and `protoc` were missing from the local PATH.

Start the local DMZ server:

```bash
cd server.dmz/mind_daemon_dmz
go run ./cmd/mind-daemon-dmz
```

Run the CLI in another terminal:

```bash
dotnet run --project cli/MindDaemonCli/MindDaemonCli.csproj
```

The CLI calls the server at `http://localhost:8080` by default. Override it with
`MIND_DAEMON_SERVER_URL` when needed.

The RPC contract lives in `proto/minddaemon/v1/hello.proto`. After editing proto
files, regenerate Go stubs from the `proto` directory:

```bash
buf generate
```

The local server uses plaintext h2c for development only.

### cli planed features
- simple file upload and download, cause why not
- sync between local machine files and server storage. (auto sync while init/ via command)
- manage todos and current plans.
- permanent tui, having features like notification 
- unified notifications that i want to check, that the status of home lab backup, the current status of the cloud instance i am having.
- a fairly simple password management


### server planed features
- llm for organizing todos, plans. (triggers if updated / scheduled job / via cli command)
- pop notification (notification list with timing)
- delegate task to more powerful llm, under user permission
- notification system, that can be triggered by automation

## Cloud deployment

The DMZ server is hosted on **GCP Cloud Run**, built and deployed by **Cloud Build** on every push via `cloudbuild.yaml` at the repo root.

### Architecture

- **Registry**: Artifact Registry (`_REGION-docker.pkg.dev/$PROJECT_ID/mind-daemon/mind-daemon-dmz`)
- **Runtime**: Cloud Run (managed, `--use-http2` for end-to-end HTTP/2 / Connect RPC)
- **Auth**: IAM — service deployed `--no-allow-unauthenticated`; callers need `roles/run.invoker` + a Google ID token
- **Edge**: Cloudflare DNS/TLS/WAF on a custom domain (DNS-only or proxied); IAM remains the authoritative gate
- **CLI auth**: Application Default Credentials (`gcloud auth application-default login`); the CLI mints an ID token at runtime

### One-time GCP setup

```bash
export PROJECT_ID=<your-gcp-project>
export REGION=us-central1

gcloud config set project $PROJECT_ID
gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com iamcredentials.googleapis.com

gcloud artifacts repositories create mind-daemon --repository-format=docker --location=$REGION

gcloud iam service-accounts create mind-daemon-dmz-run --display-name="Cloud Run runtime SA for mind-daemon-dmz"

PROJECT_NUM=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
CB_SA=${PROJECT_NUM}@cloudbuild.gserviceaccount.com
for role in roles/run.admin roles/artifactregistry.writer roles/iam.serviceAccountUser; do
  gcloud projects add-iam-policy-binding $PROJECT_ID --member=serviceAccount:$CB_SA --role=$role
done
```

Connect the GitHub repo via **Cloud Build → Triggers → Connect repository** in the GCP console, then create a trigger pointing at `/cloudbuild.yaml`.

Grant yourself (or any CLI user) invoker access on the deployed service:

```bash
gcloud run services add-iam-policy-binding mind-daemon-dmz \
  --region=$REGION --member=user:<your-email> --role=roles/run.invoker
```

### Cloudflare DNS

After the first deploy, create a Cloud Run domain mapping in the GCP console, then add the resulting CNAME in Cloudflare (proxy mode: **Full (strict)** SSL). IAM auth is required on both the `*.run.app` URL and the custom domain.

### Substitution variables (`cloudbuild.yaml`)

Override these at trigger-creation time if your project uses different names:

| Variable | Default | Meaning |
|---|---|---|
| `_REGION` | `us-central1` | Cloud Run / AR region |
| `_AR_REPO` | `mind-daemon` | Artifact Registry repo name |
| `_SERVICE` | `mind-daemon-dmz` | Cloud Run service name |
| `_RUNTIME_SA` | `mind-daemon-dmz-run@$PROJECT_ID.iam.gserviceaccount.com` | Runtime service account |

## Current decisions

- gRPC/Connect between CLI and server (learning + type-safe contract via protobuf)
- DMZ server hosted on GCP Cloud Run; Cloudflare in front for DNS/TLS/WAF
- CLI authenticates to Cloud Run via GCP Application Default Credentials (ID token)
- Maybe a home-lab worker server later for local LLM inference jobs


