## Style guide

- Keep comments minimal; the code should be able to speak for itself.
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

## Data access

- All database access goes through the sqlc-generated `Querier` interface in
  `internal/database`. Do not open connections or run raw SQL from domain
  packages; add a named query in `sql/queries/` and regenerate sqlc instead.
- Tests use an in-memory fake `Querier`, never a live database.

## API docs

- The README is the API reference: every HTTP endpoint is documented
  there with its input, response, status codes, and a curl example. When
  adding or changing an endpoint, update its README entry in the same
  change.
- Authenticated routes read the user via `auth.UserFromContext`; never
  parse the session token yourself.

## Schema changes and the seed command

- The seed command's data lives in `cmd/seed/data/` — for now a copy of
  `internal/challengegen/samples/`, expected to become the canonical (and
  diverging) location later.
- When adding or altering tables in `sql/schema`, also update
  `cmd/seed/main.go`: extend the `TRUNCATE` list and add insert logic for new
  tables/columns, and extend `cmd/seed/data/` if the new columns need seed
  values.
- The seed command inserts through the sqlc-generated queries in
  `internal/database`. Regenerating sqlc after a schema change will break the
  seed command's compilation if it no longer matches — fix it as part of the
  same change, not later.
- Keep truncate + inserts inside one transaction so seeding is all-or-nothing.

## Commits

- Do not commit unless I tell you to.
- Use conventional commit-style messages: `type(scope): summary`
- Valid types are: `feat`, `fix`, `docs`, `chore`, `refactor`, and `test`
- Scope is optional; use the affected package or area when helpful
- Compose multiple commits if work is independent