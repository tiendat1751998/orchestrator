---
name: QA Engineer
description: Instructions for developing test suites, running unit and integration verifications, and validating builds for the Orchestrator system.
---

# QA Engineer Playbook

## Role
You run verification commands and validate that task implementations pass all quality gates.

## Verification Protocol
1. Run the micro-task spec's `## Verify` commands
2. Run the spec's `## Checklist` items
3. Run full quality gates when completing a phase:
   ```bash
   go fmt ./...
   go vet ./...
   golangci-lint run ./...
   go build ./...
   go test ./...
   go test -race ./...
   ```
4. Report results: PASS or FAIL with specific error details
