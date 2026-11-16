param([string]$Go = 'go')
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
function Assert-CleanWorktree {
 $changes = & git status --porcelain=v1 --untracked-files=all --ignore-submodules=none
 if ($LASTEXITCODE -ne 0) { throw 'Could not check Git working tree.' }
 if ($changes) { throw "Production build requires a clean Git working tree. Commit all changes and untracked files first.`n$($changes -join "`n")" }
}
Assert-CleanWorktree
if (!(Test-Path 'internal/bundle/payload/svn.zip')) { throw 'Missing pinned SVN payload.' }
$expected = '98BB798DD6C70CC875B9BDACD600594351CF88F0A39A831A5055DF061C871BB5'
if ((Get-FileHash 'internal/bundle/payload/svn.zip' -Algorithm SHA256).Hash -ne $expected) { throw 'SVN payload checksum mismatch.' }
& pnpm typecheck
if ($LASTEXITCODE -ne 0) { throw 'Frontend type check failed.' }
& pnpm build
if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed.' }
& $Go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Go tests failed.' }
& $Go vet ./...
if ($LASTEXITCODE -ne 0) { throw 'Go vet failed.' }
Assert-CleanWorktree
New-Item -ItemType Directory -Force dist | Out-Null
$commit = (& git rev-parse --short=7 HEAD)
if ($LASTEXITCODE -ne 0) { throw 'Could not read Git commit.' }
$releasePath = Join-Path 'dist' "SVNova-$commit.exe"
& $Go build -trimpath -ldflags "-s -w -X svnova/services/app.BuildCommit=$commit" -o $releasePath .
if ($LASTEXITCODE -ne 0) { throw 'Go build failed.' }
$releaseHash = Get-FileHash -LiteralPath $releasePath -Algorithm SHA256
Set-Content -LiteralPath "$releasePath.sha256" -Value "$($releaseHash.Hash)  $(Split-Path $releasePath -Leaf)" -Encoding ascii
$releaseHash
