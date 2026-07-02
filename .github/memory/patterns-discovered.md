# Patterns Discovered

Recurring code, testing, and debugging patterns that improve reliability and speed.
This is an accumulated knowledge base and should grow over time. Written in English.

## Pattern Template

### Pattern Name
- <short, descriptive name>

### Context
- <where this appears: backend/frontend/tests/build/debug>

### Problem
- <what issue repeatedly occurs>

### Solution
- <recommended approach>

### Example
```
// minimal code snippet or pseudocode
```

### Related Files
- <path 1>
- <path 2>

---

## Example Pattern

### Pattern Name
- Prefer Empty Collection Over Null

### Context
- Service/module state initialization for list-like data.

### Problem
- Initializing collection state with null requires repetitive null guards.

### Solution
- Initialize list-like state as an empty collection and treat it as the default
  no-data state; enables direct iteration without guards.

### Related Files
- <add real paths as they emerge>
