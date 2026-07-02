# Changelog

---

## [v1.1.7] — Minor Release

### 9. `BindJSON` — opt-in required validation & empty slice support

**Files:** `bind.go`, `bind_test.go`

**Problems fixed**

- `BindJSON` previously treated every struct field as required — any zero value (empty string, `0`, `false`, `nil`, empty slice) returned `ErrValidation`. This made optional fields impossible without workarounds.
- Empty slice fields (e.g. `[]string{}`) always failed validation even when the field is intentionally optional.

**What changed**

- `bindData` now only validates fields tagged `pine:"required"`. All other fields are accepted regardless of their value.
- Empty slices and nil slices are no longer treated as zero values — an empty `[]string{}` is valid without any tag.
- `isZeroValue` updated: the `len == 0` early return for `reflect.Slice` / `reflect.Array` removed; only non-empty slices with a zero-value element are flagged.

**Usage**

```go
type Body struct {
    Name  string   `json:"name"  pine:"required"` // must be non-empty
    Email string   `json:"email" pine:"required"` // must be non-empty
    Tags  []string `json:"tags"`                  // optional — empty slice is fine
    Age   int      `json:"age"`                   // optional — zero is fine
}
```

**Tests added** (`bind_test.go`)
- Required field absent → `ErrValidation`.
- Required field present with optional empty slice → no error.
- Optional field with zero value → no error.

---

## [v1.1.6] — Minor Release

### 8. Route grouping — `Group`

**Files:** `group.go` (new), `group_test.go` (new), `pine.go`, `Examples/GroupExample/` (new)

**What changed**

- `Server.Group(prefix, ...middleware) *Group` added — returns a route group that prefixes every registered route with `prefix`.
- `Group.Group(prefix, ...middleware) *Group` — nests groups; sub-groups inherit the parent prefix and middleware and can add their own.
- `Group` exposes the same HTTP method shorthands as `Server`: `Get`, `Post`, `Put`, `Patch`, `Delete`, `Options`, and `AddRoute`.
- Group middleware is scoped strictly to the group and its sub-groups — routes registered directly on the server are unaffected.
- Global middleware added via `server.Use()` (before or after group route registration) correctly wraps all routes without double-applying group middleware.
- Execution order is always: **global → group → handler**.
- `joinPath` helper normalises prefix/path concatenation — handles trailing slashes, missing leading slashes, and empty paths without producing double slashes.

**Tests added** (`group_test.go`)
- Prefixed routes resolve correctly; unprefixed paths 404.
- Nested groups concatenate prefixes and extract params.
- Group middleware runs on group routes and does not run outside.
- Global middleware runs before group middleware.
- `server.Use()` called after group registration still wraps group routes.
- Trailing-slash prefix normalisation.
- Empty path registers at the group root.
- Two sibling groups do not interfere.
- `joinPath` unit tests covering all slash edge cases.

---

## [v1.1.5] — Major Release

### 1. Radix-tree router

**Files:** `router.go` (new), `router_test.go` (new), `pine.go`, `pine_test.go`

**Problems fixed**

| # | Symptom |
|---|---------|
| 1 | Method routing was broken — `POST /users` returned 405 when `GET /users` was registered first because `ServeHTTP` stopped at the first *path* match regardless of method. |
| 4 | `matchRoute` accessed `segment[0]` without a length guard; paths like `//api/v1` produced an empty segment and panicked. |

**What changed**

- Replaced the linear `stack [][]*Route` with a radix-tree `Router`. Lookup is now O(path-length) instead of O(routes).
- Static segments, `:param` captures, and `/*` wildcard are all supported with correct priority order (static > param > wildcard).
- Method isolation: `GET /x` and `POST /x` live at the same node under different method keys, so wrong-method requests get a clean 404 instead of accidentally matching.
- `matchRoute`, `splitPath`, and `methodInt` removed from `pine.go`.
- `SetCookie` fixed to use `Header().Add()` per cookie (RFC 6265 requires one `Set-Cookie` header per cookie — the old code concatenated all into one).
- `DeleteCookie` fixed: `MaxAge: -1` was silently dropped because the guard was `> 0`. Changed to `!= 0`.
- `SameSite` constants re-numbered so `0` means "unset" instead of `Lax`, preventing the `SameSite` directive from being skipped on the default value.

**Tests added** (`router_test.go`, `pine_test.go`)
- Static, param, wildcard, and deep-nested routes.
- Method isolation on the same path.
- Consecutive-slash paths — no panic.
- Multi-cookie `SetCookie`, `DeleteCookie` expiry header, `SameSite=Lax` header.
- Benchmarks: `BenchmarkRouter_StaticRoutes`, `BenchmarkRouter_ParamRoutes`, `BenchmarkRouter_Insert`.

---

### 2. Security fixes — file upload path traversal & WebSocket origin

**Files:** `file.go`, `file_test.go`, `websocket/websocket.go`, `websocket/websocket_test.go` (new)

**Problems fixed**

| # | Symptom |
|---|---------|
| 2 | `SaveFile` joined the raw `fh.Filename` directly into the upload path; a crafted filename like `../../etc/cron.d/evil` escaped the upload directory. |
| 3 | `defaultConfig.CheckOrigin` always returned `true`, silently accepting cross-origin WebSocket upgrades and enabling CSRF-via-WebSocket attacks. |

**What changed**

- `file.go`: `fileName = filepath.Base(filepath.Clean(fileName))` strips all directory components before the path is joined.
- `websocket/websocket.go`: `CheckOrigin` removed from `defaultConfig` (left `nil`). In `New()`, the field on the Gorilla upgrader is only set when the caller explicitly provides one — otherwise Gorilla's built-in safe default applies (Origin header must match Host header).

**Tests added** (`file_test.go`, `websocket/websocket_test.go`)
- Upload with `Filename = "../../evil.txt"` — asserts file lands at `<uploadDir>/evil.txt`.
- WebSocket upgrade with mismatched `Origin` / `Host` — asserts rejection (non-101).
- WebSocket upgrade with matching origin — asserts success.

---

### 3. Cache — atomic `GetOrSet`, expiry consistency, millisecond precision

**Files:** `cache/cache.go`, `cache/cache_test.go` (new)

**Problems fixed**

| # | Symptom |
|---|---------|
| 7 | `process()` in the rate limiter called `cache.Get()` then `cache.Set()` as two separate lock acquisitions; two concurrent first-requests both saw nil and created independent entries — the second overwrote the first. |
| 9 | `Exists` checked map membership only; `Get` checked expiry. Between sweeper runs, `Exists` returned `true` for keys that `Get` would return `nil` for. |

**What changed**

- `GetOrSet(key string, fn func() (interface{}, time.Duration)) interface{}` added. The entire check-and-create runs under one write lock — no TOCTOU window.
- `Exists` now performs the same `val.exp >= time.Now().UnixMilli()` expiry check as `Get`, so the two are always consistent.
- All timestamp comparisons switched from `Unix()` (second granularity) to `UnixMilli()` — TTLs shorter than one second now work correctly.

**Tests added** (`cache/cache_test.go`)
- `GetOrSet` called from 50 concurrent goroutines — asserts `fn` is called exactly once.
- `Exists` on an expired key — asserts `false` before the sweeper runs.
- Basic `Set`/`Get`/`Delete` round-trip.
- TTL expiry with sub-second precision.

---

### 4. Rate limiter — TOCTOU race, header data-race, off-by-one

**Files:** `limiter/rate_limit.go`, `limiter/rate_limit_test.go` (new)

**Problems fixed**

| # | Symptom |
|---|---------|
| 7 | Two concurrent first-requests from the same IP both saw no entry and created separate `*entry` values; the second write overwrote the first, corrupting the counter. (Fixed via `cache.GetOrSet`.) |
| 7b | After `process()` returned, the middleware read `e.remaining` without holding `e.mu` — a race with any concurrent request updating the same entry. |
| 8 | New entries were created with `remaining: MaxRequests`; `process()` returned immediately on a new entry without decrementing, giving `MaxRequests + 1` allowed requests. |

**What changed**

- `process()` now calls `cfg.store.GetOrSet(...)` instead of `Get` + `Set`.
- `process()` signature changed from `(*entry, error)` to `(*entry, int, error)` — the `int` is `e.remaining` captured under `e.mu` before the lock is released. Callers use this snapshot for both the response header and the block decision; `e.remaining` is never read outside the lock.
- Block condition changed from `e.remaining == 0` to `rem < 0`; the counter now starts at `MaxRequests` and decrements to `-1` on the first over-limit request, so exactly `MaxRequests` requests succeed.

**Tests added** (`limiter/rate_limit_test.go`)
- Sequential requests up to `MaxRequests` — all 200; one more — 429.
- `MaxRequests` concurrent first-requests from the same IP — counter not corrupted.
- Whitelist / blacklist behaviour.
- `X-RateLimit-*` header values.

---

### 5. Logger — file descriptor leak

**Files:** `logger/logger.go`, `logger/logger_test.go` (new)

**Problem fixed**

| # | Symptom |
|---|---------|
| 10 | `openNew()` called `os.OpenFile` to create the file but discarded the handle and never assigned `l.file`. Every subsequent `Write` therefore called `openExistingOrNew` → `openNew` again, leaking one fd per write. `openExistingOrNew` also opened the file twice (once to create, once to get a handle). |

**What changed**

- `openNew` now assigns `l.file = file` and sets `l.size = 0`.
- `openExistingOrNew` opens the file exactly once; it calls `os.Stat` first and falls through to `openNew` only when the file does not exist.
- Logging functions (`Info`, `Error`, `Warning`, `Success`, `RuntimeError`) changed from `func(message interface{})` to `func(message ...any)` — accept multiple values without forcing callers to `fmt.Sprintf` first.

**Tests added** (`logger/logger_test.go`)
- Write several lines; assert content appears in the file.
- Assert `l.file` is non-nil and stable after repeated writes (no fd churn).
- Log rotation: assert a new file is created when the size limit is reached.

---

### 6. HTML render engine — `render` package, hot reload via `websocket.Watch`

**Files:** `render/engine.go` (new), `render/reload.go` (new), `render/engine_test.go` (new), `websocket/file.go`, `pine.go`; deleted `render.go`, `websocket/live_reload.go`

**What changed**

**`websocket/file.go`**

- Added exported `Watch(dir string, done <-chan struct{}, onChange func(absPath string)) error` — a shared directory-watching primitive backed by `fsnotify`. New subdirectories are added automatically. `WatchFolder` is refactored to use `Watch` internally, eliminating a duplicated event loop. `WatchFile` and `WatchFolder` remain pure streaming tools; `Watch` is the reusable core.

**`render` package (new)**

- `Engine` struct implements `pine.ViewEngine` (for rendering) and `pine.Reloader` (for hot-reload).
- Templates are named by their path relative to `ViewPath` with forward-slash separators, e.g. `"admin/dashboard.html"`, avoiding subdirectory name collisions.
- `Engine.Render()` buffers output before writing so a template error never sends a partial response. When live reload is active, a small inline `<script>` is appended automatically — no template changes needed.
- `Engine.Rebuild()` atomically swaps the template set under `sync.RWMutex`; in-flight renders are not affected.
- `render.Setup(server)` is the single entry point:
  - Calls `render.New(server.ViewPath())` and installs the engine via `server.SetEngine(e)`.
  - If `server.ReloadTemplates()` is true, also calls `render.LiveReload(server, e)` — no separate call needed.
- `render.LiveReload` registers a WebSocket endpoint at `/__pine_reload`, starts `websocket.Watch` in a goroutine to monitor `ViewPath`, and registers a shutdown hook via `server.AddShutdownHook`. On each template file change (`.html`, `.gohtml`, `.tmpl`) a 50 ms debounce fires: templates are rebuilt, then "reload" is broadcast to all connected browser tabs.

**`pine.go`**

- `ViewEngine` interface exported (was unexported `viewEngine`).
- `Reloader` interface added — engines that support hot-reload implement `Rebuild() error`; `RebuildViews()` uses a type assertion instead of hard-coding `*htmlEngine`.
- `SetEngine(ViewEngine)`, `ReloadTemplates() bool`, `AddShutdownHook(...func())` added to `Server`.
- `Engine()` method removed from `Server` — use `render.Setup(app)` instead.
- `SetLiveReloadPath()` removed — the engine now owns the live-reload path and injects the script itself.
- `Ctx.Render()` moved into `pine.go`; it no longer checks `liveReloadPath` on the server — the engine handles script injection.

**Tests added** (`render/engine_test.go`)
- Simple template, subdirectory template, custom status code, shared partials via `{{define}}` / `{{template}}`.
- `RebuildViews` picks up file edits; cached templates do not pick up edits without a rebuild.
- `render.New` returns an error for a nonexistent view path.
- `render.Setup` propagates the error.

---

### 7. Examples — SimpleRender, FullStackApp, LiveReload

**Files:** `Examples/SimpleRender/`, `Examples/FullStackApp/`, `Examples/LiveReload/`

**SimpleRender**

Minimal example: a two-page profile site using `render.Setup(app)`. Demonstrates named subdirectory templates and passing structured data to templates.

**FullStackApp**

Full-stack todo application: HTML pages rendered server-side, a REST JSON API (`GET/POST /api/todos`, `PATCH /api/todos/:id`), and static file serving (`/static/*`). Shows Pine as a complete server without a separate API gateway.

**LiveReload**

Development-mode hot-reload example. Set `ReloadTemplates: true` and call `render.Setup(app)` — the browser reloads automatically whenever a template under `views/` is saved. Demonstrates the end-to-end flow: file change → `fsnotify` event → `websocket.Watch` callback → template rebuild → browser reload.

---

## Upgrade notes

### v1.1.6 → v1.1.7

**Breaking change:** `BindJSON` no longer validates every field as required by default. If you relied on the old behaviour, add `pine:"required"` to each field that must be non-zero.

```go
// Before — all fields implicitly required
type Body struct {
    Name string `json:"name"`
}

// After — opt in explicitly
type Body struct {
    Name string `json:"name" pine:"required"`
}
```

### v1.1.5 → v1.1.6

No breaking changes. Route grouping is purely additive.

### pre-v1.1.5 → v1.1.5

| Old API | New API |
|---------|---------|
| `app.Engine("html")` or `Config{Engine: "html"}` | `render.Setup(app)` (import `github.com/BryanMwangi/pine/render`) |
| `websocket.LiveReload(app)` | Set `Config{ReloadTemplates: true}` — `render.Setup` handles it |
| `server.SetLiveReloadPath(p)` | Removed — internal to `render.LiveReload` |
| `server.RebuildViews()` | Unchanged — delegates to `engine.Rebuild()` via `Reloader` interface |
| `cache.Get` + `cache.Set` for atomic init | `cache.GetOrSet` |
| `logger.Info("msg")` | `logger.Info("msg")` or `logger.Info("count:", n)` (variadic) |
