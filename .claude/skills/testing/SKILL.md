# Testing Skill

## Philosophy
- Test behavior, not implementation
- Table-driven tests in Go
- Meaningful assertion messages that explain WHY the check matters
- 2-3 good tests per function, not exhaustive coverage

## Go Testing Patterns
- Use `testify/assert` and `testify/require`
- Use `t.TempDir()` for filesystem tests
- Use `t.Helper()` in test fixtures
- Name tests `TestFunctionName_Scenario` or `TestFunctionNameScenario`

## TypeScript Testing Patterns
- Use Vitest for Workers
- Use `describe/it` blocks
- Mock Cloudflare bindings with miniflare
