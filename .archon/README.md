# Archon tooling

This directory holds the [Archon](https://github.com/coleam00/Archon) automation for this repo. Archon runs AI workflows in isolated git worktrees.

- `workflows/add-sdk-resource.yaml`: the add-a-resource workflow (documented below).
- `commands/*.md`: the per-step prompts the workflow invokes (shared conformance
  generation plus the per-language implementation steps).
- `config.yaml`: Archon project config.

---

## `add-sdk-resource`

Adds a new **resource** (a product domain such as `forms` or `contacts`) to all three SDKs (Node, Python, and Go) from the canonical OpenAPI spec, generates the shared cross-language conformance scenarios, and opens a single draft PR.

### Invocation

Run from the repo root, on a branch (or `main`) that contains this `.archon/` directory:

```bash
archon workflow run add-sdk-resource --branch feat/<domain>-resource "<domain>"
```

- **`"<domain>"`** (the positional argument) is the resource/domain name: the first path segment after `/v1/` in `openapi/spec.json` (e.g. `forms`, `contacts`, `payment-links`). It's matched case-insensitively and tolerates hyphens, underscores, and spaces, so `"forms"`, `"Forms"`, and even a short sentence resolve the same. It must match **exactly one** domain in the spec; if it matches none or several, the run stops and lists the available domains.
- **`--branch`** names the worktree Archon creates for the run. The PR branch this workflow generates is always `archon/feat/<domain>`, cut fresh from `origin/main` at publish time.

Monitor a run with `archon workflow status` (or `archon serve` for the web UI), and list past runs with `archon workflow runs`.

### What it does

1. **preflight**: verifies your toolchain is installed (see [Prerequisites](#prerequisites)) and reports everything missing at once before doing any work.
2. **resolve-spec**: fetches `origin/main` and fails if your `openapi/spec.json` is behind, so the resource is never built against a stale contract.
3. **extract-domain**: slices the spec down to the target domain (deterministic, no AI).
4. **generate-conformance**: writes the shared `conformance/scenarios/<domain>/` set: the behavioral contract all three SDKs are implemented against.
5. **review-gate**: pauses for you to approve the endpoint set + contract (see [The approval gate](#the-approval-gate)).
6. **impl-{node,python,go}**: implements the resource + per-endpoint examples + tests in each language, in parallel.
7. **validate-{node,python,go}**: runs each language's full CI gate (install, lint, type-check, coverage, build).
8. **publish-commit**: commits per language + the shared conformance scenarios and pushes `archon/feat/<domain>`. **Only runs if all three validations pass**. If any fails, nothing is committed or pushed and no PR is opened.
9. **publish-pr**: opens the draft PR for the pushed branch. If `publish-commit` found nothing to publish, this is skipped cleanly with no PR.

Publishing is two steps (`publish-commit` then `publish-pr`) rather than one so each step's script stays well under the command-line length limit Archon hits on Windows, where a longer script is silently truncated mid-run.

The run executes in an isolated worktree and is resume-safe: if a late step fails (e.g. `gh pr create`), fix the cause and resume the run rather than starting over.

### Prerequisites

`preflight` checks for all of these up front:

| Tool | Used by |
|---|---|
| `git`, `gh` (authenticated, run `gh auth login`) | branching + opening the PR |
| `bun` | the deterministic spec-slicer step |
| `node`, `yarn` (**Yarn 1 Classic**, not Berry) | Node SDK |
| `uv` | Python SDK: virtualenv + tooling, manages Python 3.10 |
| `go`, `make`, `gofumpt`, `go-test-coverage`, `golangci-lint` | Go SDK |

`gofumpt` and `go-test-coverage` come from `cd sdk-go && make tools`.
`golangci-lint` installs separately; see <https://golangci-lint.run/welcome/install>.

### Windows: run it in Git Bash

The workflow's shell steps are bash scripts. On Windows, run the workflow from **Git Bash** (the MSYS2 shell bundled with Git for Windows): **not** PowerShell, **not** WSL:

- PowerShell can't execute the bash step bodies.
- WSL is unsupported: when Archon invokes WSL `bash` through Windows interop, the
  multi-line step scripts get mangled and fail in confusing ways.

Git Bash inherits your **Windows `PATH`**, so any tool on your PATH is available to
the workflow. If `preflight` reports a missing prerequisite, install it however you
normally would on Windows, e.g. with [Chocolatey](https://chocolatey.org/) from
PowerShell:

```powershell
choco install make golangci-lint
```

then **open a new Git Bash window** (so it picks up the updated PATH) and run again.
This is exactly how tools like `make` and `golangci-lint` get wired up for Git Bash;
no separate install inside Git Bash or WSL is needed.

> Also run the workflow from a **standalone** Git Bash window, not one launched
> from inside a Claude Code session. Workflow runs can hang silently when the
> `CLAUDECODE` environment variable is set.

### The approval gate

At `review-gate` the run pauses and prints the sliced endpoint list and a note that
the conformance scenarios have been generated under `conformance/scenarios/<domain>/`.
Review both, then:

- **Approve** to launch the three parallel implementations.
- **Reject** (with a reason) to stop.

Reject (don't approve-with-changes) if the contract is wrong: the conformance
scenarios are fixed *before* this gate, so the right fix is to stop, adjust, and
re-run. The gate is intentionally go/no-go.

### Output

A single **draft** PR against `main`, with:

- one commit per language: `feat(node|python|go): add <domain> resource`,
- one commit for the shared scenarios: `test(conformance): add <domain> scenarios`,
- a body filled in from `.github/PULL_REQUEST_TEMPLATE.md`.

If the resource already exists and the implementations make no changes, the run
reports "nothing to publish" and exits cleanly without opening a PR.
