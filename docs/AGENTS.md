# Agent 协作规范

This repository is a Go CLI project in stabilization. Treat documentation and execution rules as part of the product.

## Working Rules

1. Start by checking the current tree and the exact files that are in scope.
2. Do not revert user changes unless explicitly asked.
3. Prefer `apply_patch` for edits and avoid destructive commands.
4. Keep changes scoped; if a task crosses file boundaries, state the boundary before editing.
5. Preserve CLI and documentation compatibility unless the task explicitly calls for a breaking change.

## Verification Order

1. Run `gofmt -l .` when Go files change.
2. Run `go vet ./...` after structural changes.
3. Run `GOCACHE=/tmp/md2wechat-go-build go test ./...` for regression coverage.
4. Run `GOCACHE=/tmp/md2wechat-go-build go test -cover ./...` when coverage is part of the task.
5. If `make release-check` exists in this branch, run it before declaring release-related work done.
6. When release assets or installer scripts change, run artifact smoke and installer smoke against the same bundle before calling the work done.
7. If the task touches release or installer paths, keep the documented primary path versioned and non-`latest`.

## Release And Version Discipline

1. Keep the version source singular and explicit when the repository has one.
2. Keep install scripts, release notes, changelog, and release assets aligned.
3. If the repository does not yet provide a workflow or release gate, document the gap instead of inventing it.

## Documentation Discipline

1. Document only supported paths as the primary path.
2. Label aspirational or not-yet-shipped paths clearly.
3. If a feature is not backed by current code or release assets, do not describe it as shipped.
4. Prefer short, operational instructions over marketing copy.

## Escalation

1. Ask for approval before commands that require network access or writing outside the workspace.
2. If a command fails because of sandboxing, rerun it with the required escalation instead of working around it.
