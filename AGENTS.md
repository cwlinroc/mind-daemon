# AGENTS.md

## Project State

- This repository is an early scaffold. Treat feature descriptions in [README.md](README.md) as intent, not implemented behavior.
- The active executable surfaces are the .NET CLI in [cli/MindDaemonCli](cli/MindDaemonCli) and the Go DMZ server in [server.dmz/mind_daemon_dmz](server.dmz/mind_daemon_dmz).
- The CLI and server share a protobuf contract in [proto/minddaemon/v1/hello.proto](proto/minddaemon/v1/hello.proto).
- The current local integration is a simple `HelloService/SayHello` RPC over plaintext h2c on `http://localhost:8080`.

## Working Areas

- CLI entrypoint: [cli/MindDaemonCli/Program.cs](cli/MindDaemonCli/Program.cs)
- CLI project file: [cli/MindDaemonCli/MindDaemonCli.csproj](cli/MindDaemonCli/MindDaemonCli.csproj)
- RPC contract: [proto/minddaemon/v1/hello.proto](proto/minddaemon/v1/hello.proto)
- Buf config: [proto/buf.yaml](proto/buf.yaml), [proto/buf.gen.yaml](proto/buf.gen.yaml)
- Server entrypoint: [server.dmz/mind_daemon_dmz/cmd/mind-daemon-dmz/main.go](server.dmz/mind_daemon_dmz/cmd/mind-daemon-dmz/main.go)
- Server tests: [server.dmz/mind_daemon_dmz/cmd/mind-daemon-dmz/main_test.go](server.dmz/mind_daemon_dmz/cmd/mind-daemon-dmz/main_test.go)
- Generated Go RPC code: [server.dmz/mind_daemon_dmz/gen](server.dmz/mind_daemon_dmz/gen)
- Server module file: [server.dmz/mind_daemon_dmz/go.mod](server.dmz/mind_daemon_dmz/go.mod)

## Verified Commands

Run these from the repository root.

- Build the CLI: `dotnet build cli/MindDaemonCli/MindDaemonCli.csproj`
- Run the CLI: `dotnet run --project cli/MindDaemonCli/MindDaemonCli.csproj`
- Test the server: `cd server.dmz/mind_daemon_dmz; go test ./...`
- Run the server: `cd server.dmz/mind_daemon_dmz; go run ./cmd/mind-daemon-dmz`
- Regenerate Go protobuf/ConnectRPC code after proto edits: `cd proto; buf generate`

Notes:

- There is no solution file yet. Target the CLI `.csproj` directly.
- `dotnet build cli/MindDaemonCli/MindDaemonCli.csproj` succeeds in the current scaffold.
- `go test ./...` inside `server.dmz/mind_daemon_dmz` succeeds in the current scaffold.
- `buf` is required only when changing protobuf files and regenerating Go stubs. Generated Go files are committed.

## Agent Guidance

- Prefer small, project-local changes. Do not assume README features exist until code confirms them.
- When working on the CLI, preserve the current simple console-app structure unless the task requires new architecture.
- When working on the CLI-server contract, update [proto/minddaemon/v1/hello.proto](proto/minddaemon/v1/hello.proto), regenerate Go stubs with `buf generate`, and let the .NET build generate client code from the linked proto.
- When working on the server area, keep ConnectRPC handlers in the Go module and add narrow Go tests for behavior.
- Local RPC is plaintext h2c for development only. Do not treat it as a production transport or security model.
- If you add behavior, also add the narrowest validation available. If no tests exist yet, document the command you used to verify the change.
