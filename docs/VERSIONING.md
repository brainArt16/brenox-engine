# Brenox Engine — Versioning

How engine releases are versioned, published, and reflected in docs.

## Two version concepts

| Concept | Example | Purpose |
|---------|---------|---------|
| **Engine release** | `1.0.0` | Binary / container semver — ops, `GET /version`, GitHub releases |
| **Developer API contract** | `v1` | Public `/v1/*` routes (API key auth) — bump when breaking HTTP contract |

JWT routes under `/auth/*` and `/api/*` are not path-versioned today. Breaking changes there ship with a new engine release and SDK updates.

## Source of truth

| What | File |
|------|------|
| Engine semver | `VERSION` (repo root) |
| Default in code | `internal/version/version.go` (`Engine` var — keep in sync) |
| API contract | `internal/version.API` (`v1`) |
| Docs catalog | `brenox-web/lib/docs/engine-versions.ts` |
| OpenAPI document version | `docs/openapi.yaml` `info.version` |

**Production API:** `https://api.breno-x.com`  
**Developer console / docs:** `https://www.breno-x.com`

## Runtime

```bash
curl https://api.breno-x.com/version
# {"engine":"1.0.0","api_version":"v1","commit":"abc123"}
```

`GET /health` remains a liveness probe (DB + Redis). Version info lives on `/version` only.

## Release checklist

When shipping engine **1.0.1** (example):

1. **Engine repo**
   - [ ] Bump `VERSION` and `internal/version/version.go` default
   - [ ] Update `docs/openapi.yaml` if Developer API changed
   - [ ] Changelog row in `docs/BACKEND_TASKS.md`
   - [ ] Tag: `git tag v1.0.1 && git push --tags` (GHCR image builds from `v*` tags)
   - [ ] GitHub release notes

2. **Docs catalog** (`brenox-web/lib/docs/engine-versions.ts`)
   - [ ] Prepend new row; mark previous `current` → `supported`
   - [ ] Confirm `baseUrl` is `https://api.breno-x.com`

3. **Cross-repo**
   - [ ] SDK snippets / README if default API URL or behavior changed
   - [ ] Postman prod environment (`docs/postman/brenox.postman_environment.prod.json`)

4. **Verify**
   - [ ] `curl https://api.breno-x.com/version` matches catalog
   - [ ] Docs page shows engine version + API URL

## Docker / CI

Image build injects version at link time:

```dockerfile
ARG VERSION=1.0.0
ARG COMMIT=unknown
# -X github.com/brainart16/brenox/internal/version.Engine=${VERSION}
```

GitHub Actions `images.yml` tags images on `v*` semver tags.

## Version statuses (docs)

| Status | Meaning |
|--------|---------|
| `current` | Running in production; default in docs |
| `supported` | Still valid; security fixes only |
| `deprecated` | Upgrade recommended |
