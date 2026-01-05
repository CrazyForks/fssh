# The SDD Constitution

## Development Principles for AI-Generated Code

This constitution defines immutable principles that constrain how specifications transform into code, ensuring consistency, simplicity, and maintainability across all generated implementations.

---

## Article I: Library-First Principle

### Section 1.1: Mandatory Library Initialization

Every feature must begin its lifecycle as a standalone library. No feature may be implemented directly in application code without first being abstracted as a reusable library component.

**Rationale:** Forces modular design from inception, ensuring clear boundaries and reusability.

**Example - Correct:**
```
chat-feature/
├── lib/
│   └── chat-service/     # Standalone library
│       ├── src/
│       ├── tests/
│       └── package.json
└── app/
    └── integrates chat-service
```

**Example - Incorrect:**
```
app/
└── features/
    └── chat/            # Feature embedded in app
        └── handlers.js
```

### Section 1.2: Clear Library Boundaries

Libraries must have:
- Minimal dependencies
- Clear public interfaces
- Self-contained functionality
- Independent testability

---

## Article II: CLI Interface Mandate

### Section 2.1: Text-Based Interface Requirement

All library functionality must be accessible via command-line interface (CLI). This is non-negotiable.

### Section 2.2: Interface Standards

All CLI interfaces must:
- Accept text input (stdin, arguments, or files)
- Produce text output (stdout)
- Support JSON for structured data exchange
- Provide --help documentation

**Example:**
```bash
# Good: Observable and testable
chat-service send --message "Hello" --user alice

# Bad: Hidden in code
import { sendMessage } from 'chat-service';
```

**Rationale:** Ensures observability, testability, and composability. If functionality can't be accessed via CLI, it can't be properly tested or debugged.

---

## Article III: Test-First Imperative

### Section 3.1: Absolute Test Priority

This is non-negotiable. No implementation code may be written until:

1. Unit tests are written
2. Tests are validated and approved by user
3. Tests are confirmed to fail (red phase)

### Section 3.2: Test Approval Workflow

```
1. AI generates comprehensive tests
2. USER reviews and approves tests
3. AI confirms tests fail
4. Only then: AI writes implementation
```

**Example Test-First Flow:**
```bash
# Step 1: Write test
describe('sendMessage', () => {
  it('should deliver message to recipient', async () => {
    // Test implementation
  });
});

# Step 2: Confirm it fails ❌
# Step 3: Write implementation
# Step 4: Confirm it passes ✅
```

**Violation Example:**
```javascript
// ❌ FORBIDDEN: Writing implementation before tests
function sendMessage(text, userId) {
  // Implementation without tests
}
```

---

## Article VII: Simplicity Mandate

### Section 7.1: Maximum Three Projects

Initial implementations must use ≤3 projects. Adding projects requires documented justification.

### Section 7.2: No Future-Proofing

Do not design for hypothetical future requirements. Build for current, validated needs only.

### Section 7.3: Complexity Tracking

Any complexity beyond minimalism must be tracked and justified in a "Complexity Tracking" section.

**Example - Correct (Simple):**
```
chat-system/
├── chat-service/        # 1. Core service
├── chat-client/         # 2. Client library
└── chat-api/            # 3. HTTP interface
```

**Example - Incorrect (Over-engineered):**
```
chat-system/
├── chat-core/
├── chat-domain/
├── chat-infrastructure/
├── chat-application/
├── chat-presentation/
└── chat-shared-kernel/  # ❌ Premature abstraction
```

---

## Article VIII: Anti-Abstraction Principle

### Section 8.1: Trust Frameworks

Use framework features directly. Do not wrap or abstract framework functionality unless there's documented evidence of need.

### Section 8.2: Single Model Representation

One domain concept = one model. No separation of "entities," "DTOs," "view models," etc., unless proven necessary.

**Example - Correct (Direct):**
```javascript
// ✅ Direct framework use
const express = require('express');
app.get('/users', (req, res) => {
  // Direct Express usage
});
```

**Example - Incorrect (Over-abstracted):**
```javascript
// ❌ Unnecessary abstraction
class HttpFrameworkAbstraction {
  abstract route(path, handler);
}
class ExpressAdapter extends HttpFrameworkAbstraction {
  // Unnecessary layer
}
```

### Section 8.3: Proven Need First

Abstractions require evidence:
- "We need multiple database backends" → Proven with test cases
- "We might switch frameworks" → Not a valid justification

---

## Article IX: Integration-First Testing

### Section 9.1: Real Environment Preference

Tests must use realistic environments wherever possible:
- Real databases (not mocks)
- Actual service instances (not stubs)
- Live APIs (with test accounts)

### Section 9.2: Contract Testing Precedence

Contract tests must be written before implementation:
- Define API contracts first
- Test contracts independently
- Implement to satisfy contracts

**Example - Correct:**
```javascript
// ✅ Integration test with real database
describe('User persistence', () => {
  const db = new PostgreSQL(testConfig);
  
  beforeEach(async () => {
    await db.migrate();
  });
  
  it('should save user', async () => {
    await userRepo.save(testUser);
    const found = await db.query('SELECT * FROM users');
    expect(found).toContainUser(testUser);
  });
});
```

**Example - Incorrect:**
```javascript
// ❌ Mocked test doesn't prove real integration
describe('User persistence', () => {
  const mockDb = { save: jest.fn() };
  
  it('should save user', async () => {
    await userRepo.save(testUser);
    expect(mockDb.save).toHaveBeenCalled(); // Proves nothing
  });
});
```

---

## Implementation Gates

These gates must be passed before code generation begins:

### Gate 1: Simplicity Check (Article VII)
- [ ] ≤3 projects used?
- [ ] No future-proofing?
- [ ] Complexity justified if present?

### Gate 2: Anti-Abstraction Check (Article VIII)
- [ ] Direct framework usage?
- [ ] Single model representation?
- [ ] No unnecessary wrappers?

### Gate 3: Integration-First Check (Article IX)
- [ ] Contracts defined?
- [ ] Contract tests written?
- [ ] Real environment tests planned?

### Gate 4: Test-First Check (Article III)
- [ ] Tests written and approved?
- [ ] Tests confirmed to fail?
- [ ] No implementation written yet?

---

## Revision History

**Version 1.0** (2025-01-15)
- Initial constitution defining nine core articles
- Established immutable principles for AI-driven development
- Created enforcement gates for quality assurance

**Modification Process:**
Changes to this constitution require:
1. Clear documentation of rationale
2. Review by project maintainers
3. Backward compatibility assessment
4. Version increment

---

## Philosophy

This constitution embodies a development philosophy:

- **Observability over opacity** - CLI interfaces ensure visibility
- **Simplicity over cleverness** - Minimize complexity, maximize clarity
- **Integration over isolation** - Test in real environments
- **Modularity over monoliths** - Every feature as a library

These principles guide AI toward generating maintainable, testable, and architecturally sound code that serves human developers, not just compiles successfully.
