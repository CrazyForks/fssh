# Implementation Plan: [Feature Name]

**Feature ID:** [XXX]  
**Based on:** `spec.md` version [date]  
**Status:** Draft | Ready for Implementation  
**Created:** [YYYY-MM-DD]

---

## Executive Summary

**Feature:** [One sentence description]  
**Approach:** [High-level technical strategy]  
**Complexity:** Low | Medium | High  
**Estimated Effort:** [X] developer-days

---

## Phase -1: Pre-Implementation Gates

### Simplicity Gate (Article VII)

- [ ] Number of projects: [X] (must be ≤3)
- [ ] No "future-proofing" or speculative features
- [ ] Complexity justified below if >3 projects

### Anti-Abstraction Gate (Article VIII)

- [ ] Using frameworks directly without wrappers
- [ ] Single model representation (no DTO/Entity splits)
- [ ] No abstractions without proven need

### Integration-First Gate (Article IX)

- [ ] API contracts defined
- [ ] Contract tests written
- [ ] Real service integration planned (not mocks)

### Test-First Gate (Article III)

- [ ] Test plan approved before implementation
- [ ] Tests will be written and verified before code
- [ ] User approval process defined

**Complexity Tracking:**
[Document any deviations from simplicity with justification]

---

## Technical Architecture

### High-Level Design

```
[ASCII diagram or description of system components and their interactions]
```

### Technology Choices

| Decision | Technology | Rationale |
|----------|-----------|-----------|
| [Category] | [Choice] | [Why this choice maps to requirements] |
| Backend Framework | [e.g., Express] | [Traced to specific requirements in spec] |
| Database | [e.g., PostgreSQL] | [Performance/consistency needs] |
| Caching | [e.g., Redis] | [Latency requirements] |

**Traceability:** Each choice must reference specific requirements from spec.md.

---

## Library Structure (Article I)

### Primary Library

**Name:** `[library-name]`  
**Purpose:** [Core functionality]  
**Dependencies:** [Minimal, specific list]  
**Public Interface:** [Key exports/API]

### Supporting Libraries (if any)

**Name:** `[supporting-library]`  
**Purpose:** [Why separate from main library]  
**Justification:** [Why this complexity is necessary]

---

## CLI Interface Design (Article II)

### Commands

```bash
# Command 1
[library-name] [command] [options]
# Example: chat-service send --message "Hello" --user alice

# Command 2
[library-name] [command] [options]
```

### Input/Output Contracts

**Input Formats:**
- Text (stdin/arguments)
- JSON (structured data)
- Files (configuration)

**Output Formats:**
- stdout (success messages)
- JSON (structured results)
- stderr (error messages)

---

## Data Model

**Refer to:** `data-model.md` for detailed schemas

### Core Entities

**Entity 1: [Name]**
```
{
  "id": "string",
  "field1": "type",
  "field2": "type"
}
```

**Relationships:** [How entities relate]

---

## API Contracts

**Refer to:** `contracts/` directory for full specifications

### HTTP Endpoints

**POST /api/[resource]**
- Purpose: [What this endpoint does]
- Contract: See `contracts/post-resource.json`
- Authentication: [Required/Optional]

### WebSocket Events (if applicable)

**Event: `[event-name]`**
- Direction: Client→Server | Server→Client
- Contract: See `contracts/event-name.json`

---

## Testing Strategy (Article III & IX)

### Test Creation Order

1. **Contract Tests** - Define and test API contracts first
2. **Integration Tests** - Test with real services (not mocks)
3. **End-to-End Tests** - User scenario validation
4. **Unit Tests** - Component-level verification

### Test Environment

- **Database:** [Real PostgreSQL test instance, not mocks]
- **External Services:** [Test accounts/sandboxes]
- **Configuration:** [Environment-specific settings]

### Key Test Scenarios

From `spec.md` acceptance criteria:

1. **Scenario:** [User Story → Test Case]
   - **Given:** [Initial conditions]
   - **When:** [Action taken]
   - **Then:** [Expected outcome]

2. **Scenario:** [Another mapping]

**Complete scenarios:** See `quickstart.md`

---

## Implementation Phases

### Phase 1: Foundation

**Deliverables:**
- [ ] Library scaffolding with CLI interface
- [ ] Database schema migrations
- [ ] Contract definitions
- [ ] Contract tests (MUST PASS before Phase 2)

**Prerequisites:** None  
**Duration:** [X] days

### Phase 2: Core Functionality

**Deliverables:**
- [ ] [Feature component 1]
- [ ] [Feature component 2]
- [ ] Integration tests passing

**Prerequisites:** Phase 1 complete, tests approved  
**Duration:** [X] days

### Phase 3: Integration & Polish

**Deliverables:**
- [ ] End-to-end tests
- [ ] Error handling
- [ ] Documentation
- [ ] Performance validation

**Prerequisites:** Phase 2 complete  
**Duration:** [X] days

---

## File Creation Order

**Critical: Create files in this order to support TDD:**

1. `contracts/` - API specifications
2. `tests/contract-tests/` - Contract validation
3. `tests/integration/` - Real service tests
4. `tests/e2e/` - User scenario tests
5. `tests/unit/` - Component tests
6. `src/` - Implementation (only after tests approved)

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|-----------|
| [Technical risk] | [High/Med/Low] | [How to handle] |
| [Dependency risk] | [High/Med/Low] | [Contingency plan] |

---

## Research & Technical Decisions

**Refer to:** `research.md` for detailed analysis

### Key Research Areas

- [Comparison of technology options]
- [Performance benchmarks]
- [Security considerations]

---

## Quality Gates

### Before Starting Implementation

- [ ] All [NEEDS CLARIFICATION] in spec.md resolved
- [ ] Pre-implementation gates passed
- [ ] Test strategy approved
- [ ] Contracts defined and reviewed

### Before Completion

- [ ] All acceptance criteria met
- [ ] All tests passing (contract, integration, e2e, unit)
- [ ] No abstraction violations (Article VIII)
- [ ] CLI interface complete and documented
- [ ] Performance requirements met

---

## Traceability Matrix

| Spec Requirement | Implementation Component | Test Coverage |
|-----------------|-------------------------|---------------|
| [REQ-1] | [Component/Module] | [Test file] |
| [REQ-2] | [Component/Module] | [Test file] |

**Purpose:** Ensure every requirement maps to implementation and tests.

---

## Supporting Documents

- **Specification:** `spec.md`
- **Data Model:** `data-model.md`
- **Contracts:** `contracts/`
- **Research:** `research.md`
- **Quickstart:** `quickstart.md` (validation scenarios)

---

## Notes

**Important:** This plan should remain high-level and readable. Detailed algorithms, code examples, or verbose technical specifications must go in `implementation-details/` directory if needed.

---

## Plan Completeness Checklist

- [ ] Constitutional gates addressed and passed
- [ ] Technology choices traced to requirements
- [ ] Library structure follows Article I
- [ ] CLI interface designed per Article II
- [ ] Test-first strategy defined (Article III)
- [ ] Integration-first approach planned (Article IX)
- [ ] Simplicity maintained (Article VII)
- [ ] No premature abstraction (Article VIII)
- [ ] All phases have clear deliverables
- [ ] Risks identified with mitigations
- [ ] Traceability to spec.md requirements
