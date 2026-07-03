---
name: DevOps
description: Instructions for building Docker images, CI/CD pipelines, Go binary releases, and deployment automation for the Orchestrator system.
---

# DevOps Engineer Playbook

## Session Startup (MANDATORY)
1. Read `.agents/context/deployment-topology.md` — single binary deployment model.
2. Read `.agents/context/architecture.md` — understand the build structure.
3. Read `.agents/context/security-policies.md` — container and secret security.

**NEVER start work without knowing the deployment topology.**

## Workflow

### 1. Build & Package
- Go binary: `go build -o orchestrator ./cmd/orchestrator-cli/`
- Docker image: multi-stage build with distroless/alpine base.
- Cross-compilation: `GOOS=linux GOARCH=amd64 go build ...`
- Binary size optimization: `-ldflags="-s -w"` for production builds.

### 2. CI/CD Pipeline
- Lint: `golangci-lint run ./...`
- Build: `go build ./...`
- Test: `go test -race ./...`
- Docker: `docker build -t orchestrator:latest .`
- Release: GitHub Actions with goreleaser.

### 3. Docker Best Practices
```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /orchestrator ./cmd/orchestrator-cli/

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /orchestrator /orchestrator
USER nonroot:nonroot
ENTRYPOINT ["/orchestrator"]
```

### 4. Release Management
- Semantic versioning: `vMAJOR.MINOR.PATCH`
- Changelog: update `CHANGELOG.md` for each release.
- Git tags: `git tag v1.0.0 && git push origin v1.0.0`
- Binary artifacts: Linux, macOS, Windows builds.

## Key Rules
- Docker images MUST use non-root user.
- CI MUST run `go test -race` before merge.
- No secrets in Dockerfiles or CI configs.
- Binary size target: <50MB.

## 🚫 ANTI-FAKE RULES
- Every "build pass" → MUST run and paste output.
- Every "Docker image works" → MUST run `docker build` and paste output.
