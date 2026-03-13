# Cobra + Viper Skill

## Command Structure
- One file per command in `cmd/`
- Register with `rootCmd.AddCommand()` in `init()`
- Use `RunE` (not `Run`) for error propagation
- Use `cobra.MaximumNArgs(1)` to limit positional args

## Flag Binding

```go
func init() {
    fetchCmd.Flags().BoolP("raw", "r", false, "Output raw markdown")
    viper.BindPFlag("raw", fetchCmd.Flags().Lookup("raw"))
}
```

## Config Cascade

Order of precedence (highest to lowest):
1. Command-line flags
2. Environment variables (prefixed with app name)
3. Project `.coderank.yml` in current directory
4. Global `~/.coderank.yml`
5. Built-in defaults

## Output Modes

Every command should support three output modes:
- `--raw` — plain text (no styling)
- `--json` — JSON output (for scripting)
- Default — styled terminal output with colors and formatting

## Error Handling

Always return errors from RunE functions. Cobra handles exit codes:
```go
func runCommand(cmd *cobra.Command, args []string) error {
  if err := doWork(); err != nil {
    return fmt.Errorf("context: %w", err)
  }
  return nil
}
```

Cobra automatically exits with code 1 if an error is returned.
