# Strategy & Analysis Prompt for Gemini

> **Purpose**: Paste this into a NEW Antigravity IDE chat to get Gemini to analyze 
> and create detailed micro-tasks. NO CODE will be written — only planning documents.

---

## INSTRUCTIONS

You are a senior Go architect. Your job is to ANALYZE and PLAN — NOT write code yet.

I'm building an orchestrator in Go 1.26 that coordinates AI agents.
The project has micro-task files in `docs/tasks/` that describe what to implement.

**Your job**: Read the existing micro-tasks and create MORE DETAILED versions.
The current tasks are too surface-level for production. They need:
- More edge cases identified
- More error scenarios analyzed  
- More concurrency pitfalls documented
- More integration points clarified
- Production hardening strategies

## WHAT TO DO

### Step 1: Read the existing work
Read these files to understand what's already planned:
- `docs/tasks/phase1/index.md` — Phase 1 overview
- `docs/tasks/phase2/index.md` — Phase 2 overview
- `docs/GEMINI_PROMPT.md` — Full architecture + production gaps

### Step 2: For EACH phase, create a strategy document
Save to: `docs/strategy/phase{N}_strategy.md`

Each strategy document must analyze:

#### A. Dependency Map
- Which files depend on which other files?
- What is the exact compilation order?
- Where are circular dependency risks?

#### B. Interface Contract Analysis
- For each interface: what does the caller expect?
- What are the pre-conditions and post-conditions?
- What happens when nil is passed? Empty string? Negative number?
- What are the threading guarantees?

#### C. Error Flow Analysis
- For each function: what errors can it return?
- Which errors are retryable vs fatal?
- How do errors propagate up the call chain?
- Where should errors be wrapped vs returned as-is?

#### D. Concurrency Analysis
- Which structs are accessed from multiple goroutines?
- Where are the mutex boundaries?
- Can deadlocks occur? (lock ordering analysis)
- Where are race condition risks?

#### E. State Machine Analysis
- What states can each component be in?
- What transitions are valid?
- What happens on invalid transitions?
- Recovery from partial failures?

#### F. Resource Lifecycle
- What resources are allocated? (goroutines, channels, file handles)
- When are they released?
- What if Stop() is never called?
- What if Stop() panics?

#### G. Edge Cases Catalog
For EVERY function, list at least 3 edge cases:
- Empty input
- Nil input
- Concurrent access
- Timeout/cancellation
- Already closed/stopped
- Duplicate calls
- Very large input
- Unicode/special characters

#### H. Integration Points
- How does component A call component B?
- What data format is passed?
- What if B is not ready when A calls it?
- What if B is stopped while A is using it?

#### I. Missing Components
- What's missing for production?
- What patterns are needed that aren't in the current design?
- Security considerations?
- Performance bottlenecks?

#### J. Test Strategy
- What unit tests are needed?
- What integration tests?
- What concurrency tests?
- What failure injection tests?
- What benchmarks?

### Step 3: Create enhanced micro-tasks
After the strategy analysis, create enhanced micro-task files that include:
- ALL edge cases from the analysis
- ALL error scenarios  
- ALL concurrency concerns
- Production hardening code

Save enhanced tasks to: `docs/tasks/phase{N}_enhanced/micro_{N}.XX_name.md`

## RULES
1. DO NOT write Go source code files. Only write .md analysis/planning documents.
2. Save all output to `docs/strategy/` or `docs/tasks/phase{N}_enhanced/`
3. Be extremely detailed — the goal is that ANY AI can read these docs and produce correct, production-quality code.
4. Think about what would break in production at 1000 tasks/minute.
5. Think about what happens during a 3-hour overnight run with no human supervision.
6. Think about what happens when a provider API is down for 5 minutes.

## START

Begin with Phase 1. Read `docs/tasks/phase1/index.md` and all micro-task files,
then create `docs/strategy/phase1_strategy.md` with the full analysis.

After that, do the same for Phase 2.
Then create strategy documents for Phase 3, 4, 5, 6.
