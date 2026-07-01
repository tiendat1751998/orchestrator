$directories = @(
    "cmd/orchestrator",
    "cmd/orchestrator-cli",
    "cmd/orchestrator-worker",
    "internal/kernel",
    "internal/runtime",
    "internal/scheduler",
    "internal/eventbus",
    "internal/registry",
    "internal/lifecycle",
    "internal/config",
    "internal/logger",
    "internal/metrics",
    "internal/bootstrap",
    "sdk/agent",
    "sdk/tool",
    "sdk/provider",
    "sdk/workflow",
    "sdk/context",
    "sdk/memory",
    "sdk/search",
    "sdk/plugin",
    "sdk/event",
    "sdk/task",
    "plugins/agents",
    "plugins/providers",
    "plugins/tools",
    "plugins/search",
    "plugins/memory",
    "plugins/workflow",
    "plugins/context",
    "plugins/reviewer",
    "plugins/planner",
    "modules/mission",
    "modules/workspace",
    "modules/project",
    "modules/artifact",
    "modules/execution",
    "modules/session",
    "api",
    "web",
    "docs",
    "specs",
    "prompts",
    "templates",
    "scripts",
    "tests",
    "examples"
)

# 1. Clean up existing files/directories except allowed root items
$allowed = @(".git", "README.md", "LICENSE", "CONTRIBUTING.md", "ROADMAP.md", "CHANGELOG.md", "rebuild_architecture.ps1")
Get-ChildItem -Path . -Force | ForEach-Object {
    if ($allowed -notcontains $_.Name) {
        Remove-Item -Path $_.FullName -Recurse -Force | Out-Null
    }
}

# 2. Create new directory structure and hidden .gitkeep files
foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    $gitkeep = "$dir/.gitkeep"
    if (-not (Test-Path $gitkeep)) {
        New-Item -ItemType File -Force -Path $gitkeep | Out-Null
    }
    # Hide .gitkeep file in Windows
    $file = Get-Item -Path $gitkeep -Force
    $file.Attributes = 'Hidden'
}

Write-Host "Rebuild architecture completed successfully!"
