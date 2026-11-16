# SVNova

SVN service management application based on Gitea. Read `docs/todo.md` and relevant topic documents listed in `docs/README.md` before changing behavior. The upstream baseline and reuse boundaries are recorded in `docs/upstream/README.md`; keep any reference checkout outside this repository and do not modify it.

## Commands

Production EXE builds require a clean Git working tree, including the index and untracked files. Use `scripts/build.ps1`, which fails on dirty state before checks and again before packaging; commit changes before building a release. Release artifacts are named `dist/SVNova-<commit>.exe` with a matching `.exe.sha256` file; the filename and embedded version use the same Git commit.

Go 1.26.4, Node >=22.18, pnpm 11. Use `pnpm install --frozen-lockfile`, `pnpm typecheck`, `pnpm build`, `go test ./...` and `go vet ./...`. Run gofmt on changed Go files and go mod tidy after dependency changes.

`scripts/quality.ps1` runs type checking, tests and vet; `scripts/build.ps1` builds the frontend and single EXE. Both accept `-Go <absolute executable>`. Set `SVN_BIN` to run real SVN integration tests; otherwise they explicitly skip. `scripts/smoke.ps1` tests disposable fixtures on localhost port 3000. Never claim skipped or unperformed checks passed.

Run a single Go test with `go test -run '^TestName$' ./modulepath/`. Use the commands available in this project; upstream Makefile, Vitest and Playwright targets are not available here unless explicitly added.

## Code conventions

- Include the current year and SPDX license identifier in new Go file headers. Preserve upstream copyright and attribution when adapting existing files.
- Preserve existing comments that remain relevant. Keep new comments short and explain why, rather than narrating the code; prefer same-line comments when clear.
- Ensure edited files have no trailing whitespace.
- Prefer unit tests when logic can be tested in isolation. Keep integration fixtures small and tests fast, without weakening real SVN or recovery verification to meet a timing target.
- In TypeScript, use non-null assertions only when an invariant guarantees the value exists; use optional handling when absence is valid.
- Prefer flex/grid layout and container gaps over repeated child margins. Use existing theme variables and helpers; do not assume upstream Tailwind utilities are installed.

## Architecture

- Reuse Gitea's Go backend, models/services/routers layering, Go templates and TypeScript/Vue/Vite frontend. Preserve its visual conventions.
- Support SVN only. Do not expose Git transport, Git repositories or Git-specific workflows. Additional product features require user discussion.
- Reuse account, local authentication, LDAP, organization, session and administrator authorization code where practical; inspect Git dependencies before reuse.
- Place SVN commands behind services/svn. Never interpolate input into shell commands. Apply timeouts, cancellation and output limits. Do not host SVN protocol services.
- Enforce administrator access in the backend. Organization/team permissions protect entire repositories, browsing, downloads and task records; no path-level permissions or anonymous access.
- Never store LDAP user passwords or log credentials. Encrypt stored service credentials and preserve encryption keys across deployments.
- Serialize repository mutations and mirror operations. Persist task state and retain verified copies on failure. Backup means a remote mirror; do not add archive/restore products.
- Local migration is administrator-only and moves a stopped source into the dedicated root. Remote HTTP/HTTPS migration creates an organization mirror; owners must not receive arbitrary filesystem access.
- Windows x64 single EXE targets Windows 10/11 and Server 2016+. SQLite, repositories and keys remain outside the EXE. Single instance, local disks only.
- Preserve upstream licenses and attribution in copied code. Record the upstream baseline and local changes.
- Except for the SVNova logo, use Gitea's Octicons from the pinned @primer/octicons dependency through web_src/Icon.vue. Do not substitute Unicode symbols or mix unrelated icon sets; preserve icon licenses in bundled assets.

## Documentation and quality

Keep `docs/task-scheduling.md` synchronized in the same change whenever task triggers, scheduling cadence, locks, queue ordering, states, timestamps, retries, recovery or shutdown behavior change. Update its Mermaid diagrams and source index along with relevant tests and topic documents.

Maintain current requirements, design, constraints and operations in the topic documents listed in `docs/README.md`. Do not create timestamp-named feature documents or append development diaries, local process IDs, temporary executable names or workstation paths. Keep database, API, deployment and upstream attribution documents synchronized with changes.

Track unfinished work in `docs/todo.md` using ⏳ pending/deferred and 🚧 in progress. Update status when work starts, remove completed items and update the relevant topic documents. Record verified scope and unverified acceptance boundaries in `docs/acceptance.md`; separate user-confirmed real-environment acceptance from automated or simulated checks.

Use English identifiers, comments and log messages; Chinese documentation and Chinese-only UI. Reuse Gitea light-theme variables. Conventional single-line commit messages, no body or Co-Authored-By. Do not commit without a user request.

Use Conventional Commits for PR titles as well: `type(scope): subject`, with `!` for breaking changes and `test` for test-only changes. Never force-push, amend or squash unless explicitly requested; use new commits and normal pushes for updates. Do not add Signed-off-by on behalf of a human. Include accurate assistance attribution in issue or PR comments when asked to write them; writing comments does not authorize posting them.

Run meaningful authorization, LDAP, SVN integration and recovery checks. Real LDAP, HTTP/HTTPS SVN, Windows service registration and Server 2016 acceptance require actual environments; file-source tests are insufficient. Never use user repositories for tests; generate isolated fixtures.
