# cargo clippy

**What it checks:** Rust linter — catches common mistakes and suggests idiomatic Rust.

## Without prettyout

### Default output

```
    Checking clippy_test_errors v0.1.0 (/tmp/t-clippy)
warning: unneeded `return` statement
 --> src/main.rs:6:5
  |
6 |     return;
  |     ^^^^^^
  |
  = help: for further information visit https://rust-lang.github.io/rust-clippy/rust-1.95.0/index.html#needless_return
  = note: `#[warn(clippy::needless_return)]` on by default
help: remove `return`
  |
5 -     let _x = vec![1, 2, 3];
6 -     return;
5 +     let _x = vec![1, 2, 3];
  |

warning: calls to `push` immediately after creation
 --> src/main.rs:2:5
```

### JSON (NDJSON stream, what CI/CD sees)

```json
{"reason":"compiler-message","package_id":"path+file:///tmp/t-clippy#clippy_test_errors@0.1.0","manifest_path":"/tmp/t-clippy/Cargo.toml","target":{...},"message":{"rendered":"warning: unneeded `return` statement\n --> src/main.rs:6:5\n...","$message_type":"diagnostic","level":"warning","message":"unneeded `return` statement","code":{"code":"clippy::needless_return","explanation":null},...}}
{"reason":"compiler-message",...,"message":{"level":"warning","message":"calls to `push` immediately after creation","code":{"code":"clippy::vec_init_then_push","explanation":null},...}}
{"reason":"compiler-message",...,"message":{"level":"warning","message":"useless use of `vec!`","code":{"code":"clippy::useless_vec","explanation":null},...}}
{"reason":"compiler-artifact",...}
{"reason":"build-finished","success":true}
```

## With prettyout

### Group by rule (default)

```
[WARN] clippy::needless_return (1) — unneeded `return` statement
Affected files:
  - src/main.rs — line 6
────────────────────────────────────────────────
[WARN] clippy::useless_vec (1) — useless use of `vec!`
Affected files:
  - src/main.rs — line 5
────────────────────────────────────────────────
[WARN] clippy::vec_init_then_push (1) — calls to `push` immediately after creation
Affected files:
  - src/main.rs — line 2
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

### Group by file

```
src/main.rs — 3 issues
  clippy::vec_init_then_push  line 2 — calls to `push` immediately after creation
  clippy::useless_vec  line 5 — useless use of `vec!`
  clippy::needless_return  line 6 — unneeded `return` statement
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

## What prettyout improves

- **Consistent format**: cargo clippy emits a noisy NDJSON stream mixing compiler messages and metadata → prettyout filters to `compiler-message` type only and produces the same structure as all other supported tools
- **NDJSON parsed correctly**: raw output is one JSON object per line, not a JSON array → prettyout handles NDJSON natively
- **Rule grouping**: raw output lists warnings in order → group-by-rule shows all files affected by the same clippy lint
- **Occurrence counts**: no count in raw output → `clippy::needless_return (2) — message` shows how widespread each lint is
- **Summary line**: raw output ends with compiler summary lines → `N issues · M rules · K files` is clean and consistent
