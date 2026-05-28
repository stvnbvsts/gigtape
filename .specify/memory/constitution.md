<!--
SYNC IMPACT REPORT
==================
Version change: (unversioned template) → 1.0.0
Bump rationale: MINOR — initial population from blank template to first ratified
constitution. All content is new; no prior principles existed to remove or redefine.

Modified principles:
  All principles are new (template placeholders replaced with concrete content).

Added sections:
  - Core Principles (I–IX)
  - v1 Constraints
  - Development Workflow
  - Governance

Removed sections:
  None (template placeholders replaced, not removed).

Templates:
  - .specify/templates/plan-template.md ✅ updated — Constitution Check section now
    lists concrete gates for each of the nine principles.
  - .specify/templates/spec-template.md ✅ aligned — no mandatory section changes
    required; generic structure is compatible with all principles.
  - .specify/templates/tasks-template.md ✅ aligned — Phased Delivery and Testability
    principles are reflected in task structure (phase gates, optional test tasks,
    independent story checkpoints).

Follow-up TODOs:
  None. All placeholders resolved.
-->

# Gigtape Constitution

## Core Principles

### I. Hexagonal Architecture

The system follows Hexagonal Architecture (Ports & Adapters). The domain is the center
of the system and MUST have zero knowledge of external services, frameworks, or delivery
mechanisms. All external integrations are adapters that implement domain-defined port
interfaces. Delivery surfaces (CLI, HTTP, queues) are thin layers that translate input
into use case calls and format results for output. No business logic lives outside the
use case layer.

### II. Domain Integrity

Domain entities MUST never contain infrastructure concerns. No external service IDs,
no HTTP clients, no API-specific fields belong on domain types. Partial success is a
first-class domain concept — every operation that touches multiple targets MUST produce
a structured result carrying both what succeeded and what failed. Silent skipping of
failures is prohibited.

### III. Simplicity Mandate

Start with the simplest thing that works. No premature abstractions, no future-proofing,
no speculative features. Every layer of complexity MUST be justified by a concrete,
current requirement. The initial implementation MUST NOT exceed 3 top-level deployable
components. When in doubt, choose the simpler design.

### IV. Stateless by Design

There is no persistent storage in v1. All state is session-scoped, held in memory with
a short TTL, and discarded after use. Every operation MUST be self-contained and MUST
NOT depend on state from a prior operation. Any feature requiring durable state is out
of scope for v1.

### V. Resilience

All calls to external services MUST handle failures gracefully. Rate-limit responses
MUST trigger exponential backoff. No failure is silent — every operation MUST produce
a success, partial success, or explicit failure result with a human-readable reason.
Callers are never left guessing whether an operation succeeded.

### VI. Observability

Structured logging is mandatory from day one. Every log entry MUST include consistent
contextual fields (e.g., operation name, relevant entity identifiers). Errors MUST be
logged at the use case boundary with enough context to understand what the user was
attempting. Logs MUST carry signal, not ceremony — debug verbosity is acceptable in
development; production logs must be meaningful without being noisy.

### VII. Testability

Every use case, adapter, and delivery layer MUST be independently testable without
requiring a live external API call. Port interfaces exist precisely to enable dependency
injection and mock implementations. No test should depend on live external services;
all external dependencies MUST be replaceable with injected fakes or stubs.

### VIII. Phased Delivery

The system is built and shipped in independent, self-contained phases. Each phase MUST
be fully functional on its own without requiring future phases to be present. Features
not in the current phase MUST NOT influence its design. Phase boundaries are explicit
and reviewed before implementation begins.

### IX. Code Quality

Errors are handled explicitly — no swallowed errors, no generic handlers without
logging. Code is written to be read by a developer returning to it after three months
away: clear naming, no clever tricks, no unexplained magic. This is a single-developer
project; maintainability is the primary long-term constraint.

## v1 Constraints

The following constraints apply to the initial implementation and MUST be revisited
before any v2 design begins:

- **No persistent storage**: All state is session-scoped and in-memory only (Principle IV).
- **Maximum 3 deployable components**: Components are added only when justified by a
  distinct operational boundary (e.g., separate scaling or deployment lifecycle).
- **Single-developer scope**: No team-coordination abstractions (approval workflows,
  multi-branch strategies, changelog automation) are introduced in v1.

## Development Workflow

Features MUST follow the Spec Kit workflow: specify → plan → tasks → implement. Each
step produces a reviewed artifact before the next step begins.

- A feature MUST have an approved spec before planning begins.
- A plan MUST have an approved design (including a passing Constitution Check) before
  tasks are generated.
- Implementation MUST follow the generated task list in dependency order.
- Each phase of delivery is treated as a complete, shippable unit — not a stepping stone
  to a future phase.

When a plan or spec conflicts with the constitution, the constitution wins. Document the
conflict explicitly and resolve it before proceeding.

## Governance

This constitution supersedes all other guidance documents, coding conventions, and
informal agreements. When any document conflicts with this constitution, this constitution
takes precedence.

**Amendment procedure**: Amendments are proposed by updating this file with an appropriate
version bump. The developer reviews and self-approves the amendment. Version bumps follow
semantic versioning:

- MAJOR: Backward-incompatible removal or redefinition of an existing principle.
- MINOR: New principle or section added, or materially expanded guidance.
- PATCH: Clarifications, wording fixes, or non-semantic refinements.

**Compliance review**: The Constitution Check section in every plan MUST explicitly verify
all nine principles before implementation begins. Violations MUST be documented in the
plan's Complexity Tracking section with a concrete justification.

**Versioning policy**: The version line in this file is the authoritative version. All plan
and spec documents SHOULD note the constitution version in effect at the time of creation.

**Version**: 1.0.0 | **Ratified**: 2026-04-21 | **Last Amended**: 2026-04-21
