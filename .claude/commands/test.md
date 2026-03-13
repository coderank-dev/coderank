Run the test suite. Use the language-appropriate test command:
- Go: `go test ./... -v -race`
- TypeScript: `npm test`

If $ARGUMENTS is provided, use it to target specific packages or files:
- Go: `go test ./internal/config/... -v -race`
- TypeScript: `npm test -- --filter="$ARGUMENTS"`

Report results with pass/fail counts.
