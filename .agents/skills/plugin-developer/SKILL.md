---
name: Plugin Developer
description: Instructions for developing new agent plugins, provider plugins, and tool plugins using the Orchestrator SDK.
---

# Plugin Developer Playbook

## Role
You implement plugin code (agent, provider, tool plugins) based on micro-task specifications from `docs/tasks/phase3/` and `docs/tasks/phase4/`.

## Protocol
1. Read the micro-task spec file completely
2. Plugins live in `plugins/` directory and can only import from `contracts/` and `sdk/`
3. Every plugin must have a `plugin.yaml` manifest (see `.agents/context/data-model.md`)
4. Every plugin implementation must include compile-time interface checks
5. Test files go alongside implementation files (`*_test.go`)

## Conventions
- Constructor: `New<PluginName>(deps...) *Plugin`
- Interface assertion: `var _ contracts.Provider = (*MyProvider)(nil)`
- Named struct fields only
- Error wrapping with `%w`
