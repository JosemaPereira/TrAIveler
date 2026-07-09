# Mock Generation Standards

## Overview
This document defines the standardized approach for mock generation across the TrAIveler project.

## Standard Mock Location Pattern

### Directory Structure
```
<package_path>/
├── interface.go              # Production interface definition
├── implementation.go         # Concrete implementation
├── *_test.go                 # Unit/integration tests
└── mocks/
    └── <interface>_mock.go   # Generated mock (build tag: test)
```

### Examples
```
internal/database/
├── client.go                 # Client interface
└── mocks/
    └── client_mock.go        # MockClient (generated)

internal/ai/
├── provider.go               # Provider interface
└── mocks/
    └── provider_mock.go      # MockProvider (generated)

internal/subscription/
├── service.go                # Service interface
└── mocks/
    └── service_mock.go       # MockService (generated)
```

## Configuration

### .mockery.yaml
Located at: `backend/.mockery.yaml`

```yaml
# Global settings
quiet: false
with-expecter: true
dir: "{{.InterfaceDir}}/mocks"
filename: "{{.InterfaceName | snakecase}}_mock.go"

# Package configurations
packages:
  github.com/JosemaPereira/capstone-project-ai-bootcamp/backend/internal/database:
    interfaces:
      Client:
```

### Key Settings
- `with-expecter: true` - Enables fluent EXPECT() API
- `dir: "{{.InterfaceDir}}/mocks"` - Places mocks in /mocks subdirectory
- `filename: "{{.InterfaceName | snakecase}}_mock.go"` - Uses snake_case naming

## Usage

### Generating Mocks
```bash
# From backend/ directory
make mocks

# Or directly
mockery --config .mockery.yaml --all
```

### Using Mocks in Tests
```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/yourproject/internal/database"
    dbmocks "github.com/yourproject/internal/database/mocks"  // Import with alias
)

func TestService_WithMock(t *testing.T) {
    // Create mock from mocks subdirectory
    mockDB := dbmocks.NewMockClient(t)
    
    // Setup expectations
    mockDB.EXPECT().Ping(mock.Anything).Return(nil).Once()
    
    // Use in service
    service := NewService(mockDB)
    err := service.DoSomething(context.Background())
    
    assert.NoError(t, err)
    // Assertions verified automatically via t.Cleanup()
}
```

## Rules

### MANDATORY
1. ✅ All mocks MUST be in `/mocks` subdirectory
2. ✅ Mock files MUST have `//go:build test` build tag
3. ✅ Import mocks with package alias (e.g., `dbmocks`)
4. ✅ Commit generated mocks to git
5. ✅ Regenerate after interface changes

### FORBIDDEN
1. ❌ DO NOT place mocks in same directory as interface
2. ❌ DO NOT manually edit generated mock files
3. ❌ DO NOT import mocks without alias
4. ❌ DO NOT use mocks in production code

## Benefits

### Organization
- Clear separation between production code and test mocks
- Easy to find all mocks (look in /mocks subdirectories)
- Prevents accidental production usage (separate directory + build tag)

### Type Safety
- Compiler catches interface changes immediately
- Mock regeneration ensures consistency
- Fluent API provides compile-time validation

### Maintenance
- Single command regenerates all mocks
- Version control tracks mock changes
- Team consistency via committed mocks

## Examples

### Complete Example
See: [client_mock_example_test.go](../backend/internal/database/client_mock_example_test.go)

### Generated Mock
See: [client_mock.go](../backend/internal/database/mocks/client_mock.go)

## References

- [Mockery Documentation](https://vektra.github.io/mockery/)
- [Testing Guidelines](testing-guidelines.md)
- [Coding Guidelines](coding-guidelines.md)
- [Patterns Discovered](../.github/memory/patterns-discovered.md)

---

**Last Updated:** 2026-07-09  
**Status:** Active Standard  
**Applies To:** All new interfaces requiring mocks
