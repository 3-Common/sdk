# Archon tooling

This directory holds the [Archon](https://github.com/coleam00/Archon) automation for this repo. Archon runs AI workflows in isolated git worktrees.

- `workflows/add-sdk-resource.yaml`: the add-a-resource workflow (documented below).
- `commands/*.md`: the per-step prompts the workflow invokes (shared conformance
  generation plus the per-language implementation steps).
- `config.yaml`: Archon project config.

---

## `add-sdk-resource`

Adds a new **resource** (a product domain such as `forms` or `contacts`) to all three SDKs (Node, Python, and Go) from the canonical OpenAPI spec. It implements **Node first as the reference** (Node's types are generated from the spec, so it anchors the contract), **derives** the shared cross-language conformance scenarios from that verified implementation, then implements Python and Go against them, and opens a **stack of draft PRs** (one per language plus a conformance PR on top) so each language is reviewed on its own small diff.

### Invocation

Run from the repo root, on a branch (or `main`) that contains this `.archon/` directory:

```bash
archon workflow run add-sdk-resource --branch feat/<domain>-resource "<domain>"
```

- **`"<domain>"`** (the positional argument) is the resource/domain name: the first path segment after `/v1/` in `openapi/spec.json` (e.g. `forms`, `contacts`, `payment-links`). It's matched case-insensitively and tolerates hyphens, underscores, and spaces, so `"forms"`, `"Forms"`, and even a short sentence resolve the same. It must match **exactly one** domain in the spec; if it matches none or several, the run stops and lists the available domains.
- **`--branch`** names the worktree Archon creates for the run. The PR branches this workflow generates are `<github-username>/sdk-add-resource/<domain>/{node,python,go,conformance}`, namespaced by the GitHub user who runs it so concurrent runs by different people never collide; the bottom branch is cut fresh from `origin/main` at publish time and each higher branch is stacked on the one below it.

Monitor a run with `archon workflow status` (or `archon serve` for the web UI), and list past runs with `archon workflow runs`.

### What it does

1. **preflight**: verifies your toolchain is installed (see [Prerequisites](#prerequisites)) and reports everything missing at once before doing any work.
2. **resolve-spec**: fetches `origin/main` and fails if your `openapi/spec.json` is behind, so the resource is never built against a stale contract.
3. **extract-domain**: slices the spec down to the target domain (deterministic, no AI).
4. **impl-node** (the reference): implements the resource in the Node SDK + per-endpoint examples + unit tests + the conformance dispatcher, straight from the spec slice, and self-gates to green. Node leads because its types are generated from the spec, so its wire shapes are machine-checked against the contract; the method/arg surface it settles on becomes the cross-language contract.
5. **generate-conformance**: **derives** the shared `conformance/scenarios/<domain>/` set from the verified Node implementation (not from an imagined contract), and self-validates by running Node's conformance suite against them.
6. **validate-node**: runs Node's full CI gate (install, lint, type-check, coverage, build), now also exercising the derived conformance scenarios.
7. **review-gate**: pauses for you to approve the reference API surface + derived scenarios (see [The approval gate](#the-approval-gate)).
8. **impl-{python,go}**: implements the resource + per-endpoint examples + tests in Python and Go, in parallel, against the derived scenarios (the equivalence gate) plus each language's own unit tests.
9. **validate-{python,go}**: runs each follower language's full CI gate (install, lint, type-check, coverage, build).
10. **publish-commit**: carves the implementation into a four-branch stack (`.../node` off `origin/main`, then `.../python`, `.../go`, `.../conformance` each stacked on the one below) and pushes all four. **Only runs if all three validations pass**. If any fails, nothing is committed or pushed and no PRs are opened.
11. **publish-pr**: opens one draft PR per branch, each based on the branch below it (the node PR targets `main`). If `publish-commit` found nothing to publish, this is skipped cleanly with no PRs.

Publishing is two steps (`publish-commit` then `publish-pr`) rather than one so each step's script stays well under the command-line length limit Archon hits on Windows, where a longer script is silently truncated mid-run.

The shared conformance scenarios sit at the **top** of the stack on purpose. Each language's conformance harness globs every scenario file and hard-errors on a resource with no dispatcher, so the scenarios can only land once all three dispatchers are present. With them on top, every lower branch has its dispatcher but no scenarios yet (inert and green), and the cross-language contract is actually exercised on the top conformance PR, whose CI sees all three implementations plus the scenarios. The trade-off is that the per-language PRs are not conformance-checked until the conformance PR runs - so review and merge the stack **bottom-up**, and treat the conformance PR as the gate that validates the set.

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

At `review-gate` the run pauses after the Node reference SDK has been implemented
and has passed its full CI gate. It prints the sliced endpoint list and a note that
the conformance scenarios were derived from that Node implementation under
`conformance/scenarios/<domain>/` (and Node passes every one). Review both, then:

- **Approve** to launch the Python and Go implementations (in parallel).
- **Reject** (with a reason) to stop.

Reject (don't approve-with-changes) if the contract is wrong: the conformance
scenarios are already derived *before* this gate, so the right fix is to stop,
adjust, and re-run. The gate is intentionally go/no-go.

### Output

A stack of four **draft** PRs, layered `main <- node <- python <- go <- conformance`:

- `feat(node): add <domain> resource` (base `main`),
- `feat(python): add <domain> resource` (base = the node branch),
- `feat(go): add <domain> resource` (base = the python branch),
- `test(conformance): add <domain> scenarios` (base = the go branch).

Review and merge them **bottom-up**. After you merge the node PR, GitHub
retargets the python PR onto `main` (and so on up the stack); the conformance PR
is where cross-language conformance CI runs for all three SDKs.

If the resource already exists and the implementations make no changes, the run
reports "nothing to publish" and exits cleanly without opening any PRs.
