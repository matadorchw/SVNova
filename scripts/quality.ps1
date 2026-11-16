param([string]$Go = 'go')
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
& pnpm typecheck
if ($LASTEXITCODE -ne 0) { throw 'Frontend type check failed.' }
& pnpm test:activity-trend
if ($LASTEXITCODE -ne 0) { throw 'Activity trend tests failed.' }
& pnpm test:task-center
if ($LASTEXITCODE -ne 0) { throw 'Task center tests failed.' }
& pnpm test:search-clear
if ($LASTEXITCODE -ne 0) { throw 'Search clear tests failed.' }
& pnpm test:repository-tasks
if ($LASTEXITCODE -ne 0) { throw 'Repository task tests failed.' }
& $Go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Go tests failed.' }
& $Go vet ./...
if ($LASTEXITCODE -ne 0) { throw 'Go vet failed.' }
