# Mind Daemon

a new cli with server agent running in the background.
this project is just a scaffold now.

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

```powershell
cd server.dmz/mind_daemon_dmz
go run ./cmd/mind-daemon-dmz
```

Run the CLI in another terminal:

```powershell
dotnet run --project cli/MindDaemonCli/MindDaemonCli.csproj
```

The CLI calls the server at `http://localhost:8080` by default. Override it with
`MIND_DAEMON_SERVER_URL` when needed.

The RPC contract lives in `proto/minddaemon/v1/hello.proto`. After editing proto
files, regenerate Go stubs from the `proto` directory:

```powershell
buf generate
```

The local server uses plaintext h2c for development only.

### cli features
- simpe file upload and download, cause why not
- sync between local machine files and server storage. (auto sync while init/ via command)
- manage todos and current plans.
- permanent tui, having features like notification 
- unified notifications that i want to check, that the status of homelab backup, the current status of the cloud intance i am having.
- a fairly simple password management


### server features
- llm for organizing todos, plans. (triggers if updated / scheduled job / via cli command)
- pop notification (notification list with timming)
- deletgate task to more powerful llm, under user permission
- notification system, that can be triggered by automation

