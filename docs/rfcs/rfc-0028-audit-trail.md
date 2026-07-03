# RFC-0028: Audit Trail & Cryptographic Event Logs

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0008 (Event Model), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design of the **Audit Trail & Cryptographic Event Logs** in AEOS. To guarantee compliance and security, every event written to the append-only Event Store is cryptographically signed, creating a tamper-proof execution record.

## Motivation

In enterprise development environments, AI-generated code changes must be auditable and secure.
- We must prove that no external intruder modified execution logs.
- By chaining event records using SHA256 hashes (similar to blockchain block headers) and signing them with the Kernel's cryptographic key, we verify log integrity.

## Design

### 1. Architectural Placement

Hash chaining and signing are handled by the `EventStore` writer before events are flushed to the SQLite disk.

```
  New Event ──► Hash Chaining (Prev Event Hash) ──► Sign with Private Key ──► SQLite disk
```

---

### 2. Contracts (`contracts/event/audit.go`)

```go
package event

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// SignedRecord represents a cryptographically validated event block.
type SignedRecord struct {
	Event     fsm.TransitionRecord `json:"event"`
	PrevHash  string               `json:"prev_hash"`
	Hash      string               `json:"hash"`
	Signature string               `json:"signature"`
}

// AuditTrailVerifier validates event chain integrity.
type AuditTrailVerifier interface {
	// VerifyChain checks the hash integrity of the entire event history.
	VerifyChain(ctx context.Context, missionID string) (bool, error)
}
```

## Impact

- **Tamper-Proof Compliance**: Any manual deletion or modification of event rows in the SQLite database will break the hash chain, triggering instant validation alerts.
- **Enterprise-Ready Audits**: Meets security criteria for deployment in regulated environments (finance, health).
