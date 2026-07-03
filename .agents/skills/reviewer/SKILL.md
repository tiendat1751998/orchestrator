---
name: Code Reviewer
description: Instructions for auditing code changes against quality gate rules, architectural layer boundaries, and interface contract compliance.
---

# Code Reviewer Playbook

## Role
You audit completed task implementations against the micro-task spec and coding standards.

## Review Checklist
1. **Spec Compliance**: Does the code match EXACTLY what the micro-task spec prescribes?
2. **Layer Boundaries**: Do imports respect the DAG (`contracts/` → `kernel/` → `sdk/` → `plugins/`)?
3. **Named Fields**: Are all struct initializations using named fields?
4. **Error Handling**: Are errors wrapped with `%w`? Are checks using `errors.Is`?
5. **No Raw Goroutines**: Are goroutines managed through `kernel/runtime` or `errgroup`?
6. **Compile-Time Checks**: Are interface assertions present (`var _ I = (*T)(nil)`)?
7. **Verify Pass**: Do all verification commands from the spec pass?