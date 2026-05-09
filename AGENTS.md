# AGENTS.md

## Project State

- This repository is an early scaffold. Treat feature descriptions in [README.md](README.md) as intent, not implemented behavior.
- The active executable surface is the .NET CLI in [cli/MindDaemonCli](cli/MindDaemonCli).
- The server area in [server.dmz/mind_daemon_dmz](server.dmz/mind_daemon_dmz) is currently only a Go module stub with `go.mod` and no Go packages.

## Working Areas

- CLI entrypoint: [cli/MindDaemonCli/Program.cs](cli/MindDaemonCli/Program.cs)
- CLI project file: [cli/MindDaemonCli/MindDaemonCli.csproj](cli/MindDaemonCli/MindDaemonCli.csproj)
- Server module file: [server.dmz/mind_daemon_dmz/go.mod](server.dmz/mind_daemon_dmz/go.mod)

## Verified Commands

Run these from the repository root.

- Build the CLI: `dotnet build cli/MindDaemonCli/MindDaemonCli.csproj`
- Run the CLI: `dotnet run --project cli/MindDaemonCli/MindDaemonCli.csproj`

Notes:

- There is no solution file yet. Target the CLI `.csproj` directly.
- `dotnet build cli/MindDaemonCli/MindDaemonCli.csproj` succeeds in the current scaffold.
- `go list ./...` inside `server.dmz/mind_daemon_dmz` currently returns no packages.

## Agent Guidance

- Prefer small, project-local changes. Do not assume README features exist until code confirms them.
- When working on the CLI, preserve the current simple console-app structure unless the task requires new architecture.
- When working on the server area, establish a real package layout before adding instructions, tests, or build steps that depend on Go source files.
- If you add behavior, also add the narrowest validation available. If no tests exist yet, document the command you used to verify the change.