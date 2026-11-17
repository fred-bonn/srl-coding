## Style guide

- Keep things in one function unless composable or reusable.
- Do not extract single-use helpers preemptively. Inline the logic at the call site unless the helper is reused, hides a genuinely complex boundary, or has a clear independent name that improves the caller.
- Keep everything unexported by default. Only export when needed, and keep the number of entrypoints to packages minimal.
- Avoid using the `any` type.
- Use an empty struct when we need represent existence without storing any data (e.g. maps as hash sets).
- Before adding complexity for a speculative or vanishingly unlikely race or security edge case, explain the concrete failure mode, likelihood, and complexity cost to the user and get their buy-in. Do not silently expand scope for theoretical robustness.
- Avoid deep nested control structures as much as possible; keep the code as flat as possible.
- Reduce total variable count by inlining when a value is only used once.

```go
// Good
if err := foo(); err != nil {
    return err
}
return nil

// Bad
err := foo()
if err != nil {
    return err
}
return nil
```

- Avoid else statements; prefer early returns.

```go
// Good
if cond {
    return 1
}
return 0

// Bad
if cond {
    return 1
} else {
    return 0
}

```

- Prefer short assigment statements whenever possible.

```go
// Good
x := 5

// Bad
var x = 5
```

- When a function has several validation branches or supporting details, make the main function read as the happy path and move supporting details into small helpers below it.

```go
// Good
func loadThing(input any) {
    config := requireConfig(input)
    metadata := readMetadata(input)
    return createThing(config, metadata)
}

func requireConfig(input any) {
    ...
}
```

## Commits

- Do not commit unless I tell you to.
- Use conventional commit-style messages: `type(scope): summary`
- Valid types are: `feat`, `fix`, `docs`, `chore`, `refactor`, and `test`
- Scope is optional; use the affected package or area when helpful
- Compose multiple commits if work is independent