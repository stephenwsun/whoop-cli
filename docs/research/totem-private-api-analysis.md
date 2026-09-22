# Totem private WHOOP API analysis and implications for `whoop-cli`

**Research snapshot:** 2026-09-22. The Totem observations below are pinned to the `main` tree at commit [`f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702`](https://github.com/thebriangao/totem/tree/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702), whose latest GitHub push was 2026-09-16. GitHub repository metadata and release data were inspected on the snapshot date. No credentials, private account data, packet captures, or private API calls were used in this research.

This is a risk and architecture analysis, not an endorsement or implementation plan for private-endpoint support. The recommendation for `whoop-cli` is to keep its public, OAuth, read-only boundary and use Totem as a source of ideas for public-contract reliability and future public API coverage—not as a private-API dependency.

## Executive summary

Totem is a TypeScript/Node Model Context Protocol (MCP) server that intentionally impersonates the WHOOP iOS application's network client. Its README describes 55 MCP tools, including 38 reads, 15 writes, and two escape hatches; its private endpoint catalog records 311 templated operations across 47 microservices. The public repository describes itself as using WHOOP's “private reverse-engineered iOS API,” not the documented developer API ([repository overview](https://github.com/thebriangao/totem), [tool registry](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/tools/register.ts), [endpoint catalog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/data/endpoints.ts), [research notes](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/WHOOP.md)).

Its architecture is comparatively disciplined: a transport-neutral WHOOP client, Cognito token manager, endpoint-specific projections, Zod output schemas, MCP tool handlers, bounded read caching, write previews, and a catalog of captured endpoint shapes. That engineering quality does not make the underlying access supported or stable. Totem's own license explicitly warns that the private surface is contrary to the author's understanding of WHOOP's membership terms ([Totem `LICENSE`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/LICENSE)).

The most important contrast is authentication and authority. `whoop-cli` uses WHOOP's documented authorization-code OAuth flow, developer client credentials, least-privilege read scopes, and a localhost callback; Totem logs a member in with email/password (and MFA when challenged) through `/auth-service/v3/whoop/`, obtains AWS Cognito-style bearer and refresh tokens, and sends those tokens to private app endpoints. The official API documents the OAuth flow, client ID/secret, registered redirect URI, `offline` refresh scope, and token rotation ([WHOOP OAuth documentation](https://developer.whoop.com/docs/developing/oauth/)); Totem's implementation instead hard-codes the private Cognito proxy and iOS SDK identity headers ([Totem Cognito client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts), [Totem device headers](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/device.ts)).

WHOOP's official API Terms require access only by documented means, prohibit identity masking, reverse engineering, and scraping/data copies, permit WHOOP to monitor and suspend API access, and disclaim compatibility/support. Those provisions are directly relevant risk signals even though the precise legal applicability of a consumer's private-app traffic is a legal question, not something this report resolves ([WHOOP API Terms, §§1.4–1.6, 3.1, 4, 7–8](https://developer.whoop.com/api-terms-of-use/)).

**Recommendation:** do not add private Cognito login, iOS-header mimicry, undocumented endpoint calls, writes, or reverse-engineering tooling to `whoop-cli`. Preserve the current public read-only contract. Consider only public-surface improvements: documented v2 detail endpoints, one-time v1-ID migration, webhook-assisted refresh, rate-limit observability, and explicit contract/version documentation ([current WHOOP API landscape in this repository](whoop-api-landscape.md)).

## 1. Scope and evidence

### Sources inspected

* Totem source, README, `WHOOP.md`, `TOOLS.md`, generated endpoint catalog, package manifest, security policy, license, contributor guide, CI workflow, changelog, GitHub repository metadata, commit history, and releases. All source claims are linked to the pinned tree or pinned raw file above.
* WHOOP's official Developer Platform introduction, OAuth documentation, OpenAPI/API reference, rate-limit documentation, support FAQ, and API Terms of Use ([Developer Platform](https://developer.whoop.com/docs/introduction/), [OAuth](https://developer.whoop.com/docs/developing/oauth/), [API reference](https://developer.whoop.com/api/), [OpenAPI JSON](https://api.prod.whoop.com/developer/doc/openapi.json), [rate limits](https://developer.whoop.com/docs/developing/rate-limiting/), [support FAQ](https://developer.whoop.com/docs/developing/support/), [API Terms](https://developer.whoop.com/api-terms-of-use/)).
* `whoop-cli` source and docs in this repository, especially [`README.md`](../../README.md), [`whoop/client.go`](../../whoop/client.go), [`internal/auth/oauth.go`](../../internal/auth/oauth.go), [`internal/auth/store.go`](../../internal/auth/store.go), [`docs/auth.md`](../auth.md), [`docs/commands/README.md`](../commands/README.md), and [`internal/cli/schema.go`](../../internal/cli/schema.go).

### Evidence boundaries

The Totem repository contains endpoint names, request-shape notes, projections, and historical capture methodology. It does **not** publish the raw mitmproxy captures: its research document says those captures contained personal account data and are kept separately ([Totem methodology](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/WHOOP.md)). Therefore, this report treats the source and documentation as evidence of what Totem claims and implements, not as independent confirmation that every listed operation still works today. No request was sent to a private endpoint.

## 2. Repository architecture

Totem is organized as a layered MCP application:

```text
MCP client
  ├─ stdio (default) or Streamable HTTP
  └─ MCP tool registry
       ├─ Zod output schemas (`src/schemas`)
       ├─ pure projections (`src/projections`)
       ├─ tool handlers (`src/tools/v2`)
       └─ WHOOP client (`src/whoop`)
            ├─ Cognito token manager
            ├─ device/iOS identity headers
            ├─ request timeout + error classes
            └─ bounded in-memory GET cache
```

The contributor guide explicitly documents this schema → projection → handler pattern, with projections converting deeply nested BFF responses into flat domain objects and handlers doing little more than input parsing, client calls, projection, and schema validation ([Totem `CONTRIBUTING.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CONTRIBUTING.md)). The README describes Zod failures as `WhoopProjectionError` so response-shape drift is surfaced rather than silently returned as arbitrary data ([Totem architecture section](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).

The entry point selects stdio by default and imports the HTTP server only when `MCP_TRANSPORT=http`; it loads `.env`, resolves a stable installation ID, warns if the captured app version is stale, creates a `TokenManager`, starts timezone autodetection, and registers tools ([`src/server.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/server.ts)). The package requires Node `>=24`, TypeScript 6, Zod 4, Express 5, and the MCP SDK 1.30 ([`package.json`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/package.json)).

### MCP surface

The registry wires the tools by family. The README and registry together show:

* snapshot/profile/calendar, deep dives, trends/compare, stress/sleep need, and live state;
* activities and sports catalog;
* Strength Trainer history, exercise, PRs, progression, library, catalog, and writes;
* Journal and behavior impact;
* women's health;
* Coach/performance;
* smart alarm;
* communities/leaderboards;
* settings, hidden metrics, and HR-zone configuration; and
* `whoop_raw` and `whoop_endpoints` escape hatches.

The exact registration is visible in [`src/tools/register.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/tools/register.ts); per-tool inputs, source endpoints, and projected outputs are documented in [`TOOLS.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md). The README currently presents 55 tools as 38 reads, 15 writes, and two escape hatches ([Totem tool summary](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).

### Public versus private layering

Totem does not use only private endpoints. Its `whoop_workouts` tool intentionally calls the documented `/developer/v2/activity/workout` endpoint, while detailed activity data uses private `/core-details-bff/v1/cardio-details`; this mixed strategy is documented in `TOOLS.md` ([Totem workouts reference](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md)). That distinction matters for `whoop-cli`: reusing a documented path is not equivalent to adopting Totem's private authentication/session model.

## 3. Endpoints and data models

### Captured service inventory

The generated catalog is an array of method/status/path records. Its header says identifiers from the captures were templated to placeholders and that real account/community IDs, email, device serial/signature, HealthKit tokens, and capture timestamps must never ship ([`src/data/endpoints.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/data/endpoints.ts)). Totem's research says the catalog came from three mitmproxy sessions and was reduced from 419 operations to 311 templated paths across 47 microservices ([`WHOOP.md`, methodology](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/WHOOP.md)).

The catalog includes, among others:

| Domain | Representative private paths and model family | What this implies |
|---|---|---|
| Home/deep dives | `/home-service/v1/home`, `/home-service/v1/deep-dive/{recovery,sleep,strain}`, `/home-service/v1/calendar/*` | App-shaped snapshots and tile payloads rather than stable public resource models. |
| Health/live | `/health-tab-bff/v1/health-tab`, `/health-service/v2/stress-bff/{date}`, `/activities-service/v1/user-state` | Current heart-rate, activity-state, stress gauge/graph, and freshness-sensitive state. |
| Progression | `/progression-service/v3/trends/{metric}`, `/progression-service/v3/exercise/{exercise_id}` | Time-series windows for HRV, RHR, recovery, strain, sleep, stress, VO2, weight, body composition, and strength metrics. |
| Strength | `/weightlifting-service/v3/prs`, `/v3/exercise/{id}/exercise_history`, `/v3/workout-library`, `/v2/weightlifting-workout/activity` | Exercise catalogs, set history, PRs, templates, and write bodies. |
| Journal/behavior | `/journal-service/v3/journals/drafts/mobile/{date}`, `/behavior-impact-service/v1/impact*`, `/activities-service/v1/journals/behaviors/user` | Daily behavior entries, catalog/impact correlations, and journal writes. |
| Coach/AI | `/ai-conversation-bff/v1/conversation`, `/conversation/{id}/turn`, `/coaching-service/v1/health/report` | Conversation/turn state and generated health/coaching surfaces. |
| Settings/device | `/smart-alarm-*`, `/hr-zones-service/*`, `/device-config/v1/value`, `/users-service/v1/stealth-mode`, `/users-service/v1/hidden-metrics/*` | Alarm, zones, hidden-metric, stealth, and device/application preferences. |
| Social/wellness | `/community-service/*`, `/followers-service/*`, `/womens-health-service/*`, `/streaks-service/*` | Communities, leaderboards, menstrual-cycle/symptom data, and streaks. |
| Account/membership | `/users-service/v2/bootstrap*`, `/auth-service/v2/user`, `/membership-service/*`, `/privacy-service/*`, `/profile-service/*` | Identity, membership/billing, privacy, profile, and account operations. |
| Integrations/export | `/integrations-bff/*`, `/candidate-service/v1/applehealthkit/events`, `/member-data-export-service/v1/member-data-export-details` | Integration status, Apple HealthKit candidate events, and export metadata. |

These are representative groupings of the catalog, not a claim that each operation is wrapped or appropriate for `whoop-cli`; the exact list and observed statuses are in the pinned source file ([endpoint catalog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/data/endpoints.ts)).

### Data models Totem projects

The public API's documented member model is intentionally narrower: profile, body measurements, cycles, recovery, sleep, workouts, and a v1 activity-ID mapping utility. Its OpenAPI reference exposes OAuth scopes for six read areas and the public docs describe cycle, recovery, sleep, workout, profile, and body data ([WHOOP API reference](https://developer.whoop.com/api/), [OpenAPI JSON](https://api.prod.whoop.com/developer/doc/openapi.json), [WHOOP user-data docs](https://developer.whoop.com/docs/developing/user-data/)).

Totem projects additional app/BFF shapes into stable-ish tool output. Examples documented in `TOOLS.md` include:

* recovery contributors with HRV/RHR baselines, respiratory rate, SpO2, skin temperature, sleep performance, and calibration state;
* sleep stage durations and percentages, a reconstructed hypnogram, in-sleep HR, disturbances, latency/debt fields where present;
* day strain targets, zone buckets, steps, strength activity time, and workout counts;
* trends over week/month/six-month/year windows for 25 metrics, including stress, VO2 max, body composition, weight, and strength activity;
* live HR, live activity state, and live stress;
* Strength Trainer exercise metadata, set history, PRs, templates, per-exercise aggregates, custom exercises, and logged workouts;
* Journal behaviors and behavior-impact analysis;
* smart alarm schedules, HR zones, hidden metrics, stealth mode, profile updates, community leaderboards, women's health, and Coach conversations.

The source-backed per-tool descriptions and endpoint mappings are in [`TOOLS.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md). The report does not treat all nullability notes there as a stable server contract: the same document records that fields disappear or move when WHOOP changes BFF tile shapes.

### Writes and authority

Totem exposes writes despite `whoop-cli`'s read-only policy. The endpoint catalog includes POST, PUT, PATCH, and DELETE operations for creating/deleting activities, logging strength workouts, saving templates, creating custom exercises, recording journal and symptoms, editing sleep, configuring smart alarms and HR zones, editing profiles/privacy, changing communities, and toggling hidden metrics. The catalog's write entries are visible in its POST/PUT/DELETE sections ([`src/data/endpoints.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/data/endpoints.ts)).

Totem puts a `confirm:false` preview gate around writes and performs local catalog/input validation. That is a useful defense against accidental AI mutation, but it is not a permission boundary: the same process still possesses a bearer token accepted by the private app surface, and an explicit confirmation fires the write. The write-safety design is documented in the README ([write-safety harness](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)) and in the tool reference ([`TOOLS.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md)).

## 4. Authentication and session model

### Official `whoop-cli` model

`whoop-cli` follows the official authorization-code flow: a developer registers an app and redirect URI, the user authorizes it, the CLI verifies a random state at a localhost callback, and the resulting offline refresh token is stored. The project requests only `read:profile`, `read:body_measurement`, `read:cycles`, `read:recovery`, `read:sleep`, `read:workout`, and `offline` ([`docs/auth.md`](../../docs/auth.md), [`internal/auth/oauth.go`](../../internal/auth/oauth.go), [WHOOP OAuth docs](https://developer.whoop.com/docs/developing/oauth/)).

WHOOP's docs say access tokens are short-lived, refresh invalidates the old access and refresh tokens, and clients must persist the replacement refresh token ([WHOOP OAuth token-refresh documentation](https://developer.whoop.com/docs/developing/oauth/#refreshing-an-access-token)). The CLI implements this with Keychain on macOS and a `0600` JSON fallback elsewhere ([`internal/auth/store.go`](../../internal/auth/store.go), [`docs/auth.md`](../../docs/auth.md)).

### Totem's private Cognito model

Totem's bootstrap sends `InitiateAuth` with `USER_PASSWORD_AUTH` to `https://api.prod.whoop.com/auth-service/v3/whoop/`, with `ClientId: ""`; if the response requests `SMS_MFA`, `SOFTWARE_TOKEN_MFA`, or `EMAIL_OTP`, it sends `RespondToAuthChallenge` with the code. It extracts an access token, optional refresh token, ID token, and JWT expiry ([Totem `cognito.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts)).

The implementation claims WHOOP's proxy fills in the Cognito client ID and secret hash server-side, and therefore the tool does not need an extracted iOS client secret ([Totem `cognito.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts)). This is materially different from the official OAuth model: the user supplies their WHOOP password to a local script, and the script obtains app-session credentials rather than a developer-scoped grant.

Refresh uses `REFRESH_TOKEN_AUTH` at the same proxy and assumes MFA is not required for refresh. `TokenManager.getToken()` performs a single-flight refresh when the JWT is within a default 60-second skew of expiry and persists rotated tokens through an injected store ([Totem `cognito.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts), [Totem `token_manager.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/token_manager.ts)). The default `.env` store writes tokens in place and attempts `0600`; a memory-only mode intentionally loses rotations across process restart ([Totem `token_store.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/token_store.ts)).

### Session/deployment implications

Local stdio mode is one account per process. HTTP mode is also one account per deployment: the server loads one WHOOP token set and shares it across MCP clients. Totem's security policy explicitly says there is no multi-tenancy and that whoever controls an HTTP host can read the environment and process memory ([Totem `SECURITY.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/SECURITY.md)).

Totem's HTTP layer adds a separate MCP bearer token and optional OAuth 2.1/PKCE connector flow. It derives an HMAC signing key from `MCP_AUTH_TOKEN`, maintains one MCP server/transport pair per session, expires idle sessions after 30 minutes, and exposes unauthenticated `/health`; `/mcp` accepts either static bearer or its connector-issued token ([Totem `server-http.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/server-http.ts)). This is an MCP access layer, not WHOOP authorization, and it increases the blast radius of a remote deployment because a valid connector can invoke private reads and writes.

## 5. Transport behavior and app impersonation

### Request construction

The private client fixes `BASE_URL` to `https://api.prod.whoop.com` and appends `apiVersion=7` to every request. It calls `getToken()` before every request, sends JSON bodies for writes, uses a 30-second `AbortController` deadline, classifies 401 as auth-expired and 5xx as server errors, and parses other non-2xx bodies into `WhoopApiError` ([Totem constants](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/constants.ts), [Totem client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts)).

Unlike the public `whoop-cli` client, this code does not use the documented `/developer` base path, OAuth client credentials, documented scopes, or public pagination model. The current CLI constructs GETs under `https://api.prod.whoop.com/developer`, sends `Authorization: Bearer <OAuth access token>`, applies context timeouts, and retries 429/5xx according to `Retry-After` or exponential backoff ([`whoop/client.go`](../../whoop/client.go), [`whoop/errors.go`](../../whoop/errors.go)).

### Device identity headers

Totem sends a fixed set of headers described as captured from iOS app version 5.52.0/build 595097: `user-agent: iOS`, `x-whoop-device-platform`, iOS version/build/bundle, a per-install uppercase UUID, timezone, clock format, currency, locale, language, accept, and priority ([Totem `device.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/device.ts)). The comments expressly say the purpose is to make requests “indistinguishable” from iOS traffic and to avoid a one-line detection rule. The README/changelog similarly describe the headers as camouflage and say per-request Sentry/marketing fields were deliberately omitted ([Totem device implementation](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/device.ts), [Totem 1.2.1 changelog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CHANGELOG.md)).

This is a particularly important risk boundary for `whoop-cli`: copying these headers would not merely add compatibility; it would explicitly adopt a private-client impersonation technique. It should not be done.

### Cache, concurrency, and invalidation

Totem's client has a process-local LRU capped at 100 entries/10 MiB. It coalesces identical in-flight GETs, stores only completed non-pending responses, assigns endpoint-specific TTLs, never caches live HR/activity-state/stress, and invalidates by tags after writes ([Totem `client.ts`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts)). The 1.5.0 release notes document current-day TTLs of 30 seconds for home, 60 seconds for activity/sleep-need, five minutes for recovery/sleep/trends/lifts/journal, 15 minutes for calendar, and ten minutes for settings, with longer historical TTLs ([Totem 1.5.0 release](https://github.com/thebriangao/totem/releases/tag/v1.5.0)).

The current CLI does not cache data. That is appropriate for its small public read-only contract; if caching is considered later, the official API Terms' restriction on permanent copies and cache duration must be treated as a design requirement ([WHOOP API Terms §4.2](https://developer.whoop.com/api-terms-of-use/)).

## 6. Reverse-engineering evidence and maintenance behavior

### Capture methodology

Totem's primary research document says it used mitmproxy with an iPhone Wi-Fi proxy, a trusted mitmproxy CA, three sessions across two accounts, and a deduplication pipeline keyed by method, templated path, body shape, and status. It states that the iOS app did not implement certificate pinning in the tested version, allowing HTTPS contents to be observed after the proxy CA was trusted ([Totem methodology](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/WHOOP.md)).

The same document says the captures included a read-heavy primary account session and a separate test account for onboarding/write testing, and that raw captures were withheld because they contained personal data ([Totem capture-session description](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/WHOOP.md)). This is responsible data handling for the published repository, but it also means outside maintainers cannot reproduce all observations from committed fixtures alone.

### Shape drift is expected, not exceptional

The README and changelog document real fixes caused by BFF drift: recovery/strain tiles moved from `GRAPHING_CARD` to `SCORE_GAUGE`/`CONTRIBUTORS_TILE`; trend bar values moved into per-bar scrubber details; journal entries became nested; stress data moved from a misleading string field to gauge/graph data; workout HR curves required reconstructing timestamps from `position_x`; and a pending activity's detail endpoint returned 400 until scoring completed ([Totem README architecture](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md), [Totem changelog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CHANGELOG.md), [Totem tools reference](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md)).

These fixes are evidence of a high maintenance burden: undocumented payloads can return HTTP 200 while changing semantic location, and output schema validation alone cannot prove the values are correct. Totem's own contributor guide says maintainers should use `whoop_raw`, save a new fixture, diff the shape, update the projection, and document the migration ([Totem projection-drift workflow](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CONTRIBUTING.md)).

### Catalog and fixture hygiene

The 1.2.0 changelog records that concrete identifiers accidentally present in the generated endpoint catalog—including personal email, user/community IDs, device serial/signature, HealthKit tokens, and timestamps—were later templated and deduplicated ([Totem 1.2.0 changelog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CHANGELOG.md)). This is a concrete reminder that reverse-engineering artifacts can leak credentials or health identifiers even when application code itself appears safe.

## 7. Operational reliability

### Strengths

1. **Single-flight refresh.** Concurrent tool calls do not independently race refresh-token use; `TokenManager` shares one refresh promise ([Totem token manager](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/token_manager.ts)).
2. **Explicit timeout.** Cognito and data requests have a 30-second timeout ([Totem Cognito/client/constants](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts), [constants](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/constants.ts)).
3. **Bounded cache and session sweep.** The read cache has hard size limits, and the HTTP server reclaims idle MCP sessions after the code's documented 30-minute window ([Totem client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts), [HTTP server](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/server-http.ts)).
4. **Write previews and local validation.** Writes require explicit confirmation and validate IDs/values against bundled catalogs before a request ([Totem write-safety/README](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).
5. **Release discipline.** v1.5.0 and v1.5.1 release notes describe focused reliability/security passes, regression tests, and fixes for token persistence, timezone handling, prompt handling, caches, and shape drift ([Totem releases](https://api.github.com/repos/thebriangao/totem/releases?per_page=20), [CI workflow](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/.github/workflows/ci.yml)).

### Weaknesses and failure modes

1. **Private API drift is an availability risk.** WHOOP can rename internal services, change API version 7, require different headers, change BFF tile shapes, or remove an endpoint without notice. Totem's changelog already records several such changes ([Totem changelog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CHANGELOG.md)).
2. **No official private contract or support.** WHOOP's official API Terms disclaim uninterrupted/error-free operation, accuracy, completeness, compatibility, and technical support even for the documented APIs ([WHOOP API Terms §§1.11, 8.1](https://developer.whoop.com/api-terms-of-use/)); a private app surface has less contractual basis still.
3. **Bearer compromise is total.** Totem's security policy says the bearer or refresh token gives full account access, and a remote host operator can read secrets and process memory ([Totem `SECURITY.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/SECURITY.md)).
4. **Memory token store trades durability for deployment compatibility.** The documented memory mode loses rotated refresh tokens on restart, requiring re-bootstrap around token expiry ([Totem token store](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/token_store.ts)).
5. **Live data has semantic freshness caveats.** Totem intentionally bypasses completed caching for live HR/state/stress, but its own tools document that current BPM may be null/stale and pending activities cannot be detailed until scored ([Totem `TOOLS.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md)).
6. **Official quota guidance does not necessarily cover private traffic.** WHOOP documents 100 requests/minute and 10,000/day for the public API and returns 429 with rate-limit headers ([WHOOP rate limits](https://developer.whoop.com/docs/developing/rate-limiting/)); Totem warns to keep use single-digit RPS but has no published private quota guarantee ([Totem disclaimers](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).

## 8. Data beyond the WHOOP public API

The official OpenAPI reference exposes the standard member surface under six read scopes: profile, body measurement, cycles, recovery, sleep, and workout; it also documents access revocation, v1 activity-ID mapping, and a separate Trusted Partner API ([WHOOP API reference](https://developer.whoop.com/api/), [OpenAPI JSON](https://api.prod.whoop.com/developer/doc/openapi.json)). WHOOP's FAQ explicitly says continuous heart-rate data is not available via the public API ([WHOOP support FAQ](https://developer.whoop.com/docs/developing/support/#does-the-api-offer-access-to-continuous-heart-rate-data)).

Against that baseline, Totem's documented private surface reaches at least these categories not present in the standard public read-only contract:

* per-stage sleep timeline/hypnogram and in-sleep heart-rate curves;
* stress-monitor intraday timeline and live stress;
* 25-metric trend windows, including VO2 max, body composition, weight, stress, and strength activity;
* Strength Trainer set history, exercise catalog, PRs, templates, custom exercises, and workout logging;
* Journal behavior catalog, entries, and behavior-impact correlations;
* Whoop Coach conversations and performance assessments;
* smart-alarm reads/writes, HR-zone reads/writes, profile/privacy settings, hidden metrics, stealth mode;
* communities, leaderboards, followers, streaks, integrations, Apple HealthKit candidate events, and member export metadata; and
* women's-health cycle/symptom surfaces and associated writes.

Those differences are supported by Totem's per-tool source/endpoints table, not by an inference that all of the private catalog is a stable product promise ([Totem `TOOLS.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md), [endpoint catalog](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/data/endpoints.ts)).

## 9. Legal, terms, privacy, and security risks

### Official API Terms signals

The official WHOOP API Terms contain several clauses that make private API adoption a poor fit for a supported CLI:

* **Documented access only / no masking:** §1.4 says a company may access an API only by means described in that API's documentation, must use assigned developer credentials, and must not misrepresent or mask its identity or application identity ([WHOOP API Terms §1.4](https://developer.whoop.com/api-terms-of-use/)). Totem's stated design goal is to blend with iOS traffic and it uses app identity headers rather than developer credentials ([Totem device headers](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/device.ts)).
* **Monitoring and suspension:** §1.6 allows WHOOP to monitor API use and suspend access without notice if it reasonably believes the terms are violated ([WHOOP API Terms §1.6](https://developer.whoop.com/api-terms-of-use/)).
* **Reverse engineering:** §3.1(10) prohibits reverse engineering or extracting WHOOP algorithms/source code from an API or related software except where applicable law prohibits the restriction ([WHOOP API Terms §3.1(10)](https://developer.whoop.com/api-terms-of-use/)). Totem's own license describes its network reverse engineering and warns users about this risk ([Totem `LICENSE`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/LICENSE)).
* **Data extraction and copies:** §4.2 prohibits scraping, building databases, permanent copies, or retaining cached copies longer than the cache header permits, absent express permission or legal authorization ([WHOOP API Terms §4.2](https://developer.whoop.com/api-terms-of-use/)). Totem's bounded in-memory cache is limited and non-persistent, but a downstream deployment or MCP client may still copy health data; this needs explicit policy review rather than being assumed safe ([Totem client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts)).
* **Privacy policy and security duties:** §§2.1–2.4 require end-user authorization, privacy disclosures, security controls, encryption at rest/in transit, and prompt incident reporting for a company application ([WHOOP API Terms §§2.1–2.4](https://developer.whoop.com/api-terms-of-use/)).
* **Termination and deletion:** §7 requires stopping use and deleting permitted cached/stored content on termination or API discontinuation ([WHOOP API Terms §7](https://developer.whoop.com/api-terms-of-use/)).
* **No compatibility/support warranty:** §§1.11 and 8.1 disclaim compatibility, support, uninterrupted/error-free operation, and accuracy/completeness ([WHOOP API Terms §§1.11, 8.1](https://developer.whoop.com/api-terms-of-use/)).

The official Terms are written for “Company” use of the documented WHOOP APIs; they do not settle every question about a member using an unofficial client against internal app traffic. The safe engineering conclusion is narrower and stronger: a project that intentionally reverse-engineers, masks identity, and invokes undocumented write endpoints cannot honestly be presented as an official/supported WHOOP integration. Totem itself makes that disclosure in its license, README FAQ, and disclaimers ([Totem license](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/LICENSE), [Totem FAQ/disclaimers](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).

### Credential and health-data risks

Totem's local auth flow stores the account email and rotated Cognito tokens in `.env`; the password is persisted only during bootstrap and then removed. Its security policy correctly warns that anyone with local file/process access has full account access ([Totem bootstrap](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/scripts/cognito_bootstrap.ts), [Totem security policy](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/SECURITY.md)). Remote HTTP deployment additionally gives the host operator access to the WHOOP tokens and health data; the project is deliberately single-account, not tenant-isolated ([Totem `SECURITY.md`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/SECURITY.md)).

The data categories are sensitive health, sleep, biometric, location/integration, journal, reproductive-health, and social data. Totem's public code avoids shipping raw captures, templates identifiers in its generated catalog, uses HTTPS to WHOOP, limits logs, and documents token hygiene; nonetheless, MCP clients receive the projected data and may retain conversation transcripts. The repository's claim that Totem itself has no telemetry or third-party outbound data is source-backed for the implementation, not a guarantee about the user's MCP host or LLM provider ([Totem privacy/security README](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md), [Totem security policy](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/SECURITY.md)).

## 10. Licensing and maintenance signals

### License

The package manifest labels the project MIT and the committed license grants ordinary use/modification/distribution rights, but adds a prominent non-affiliation/private-API warning and legal-risk disclaimer ([`package.json`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/package.json), [`LICENSE`](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/LICENSE)). The MIT copyright line is 2026; the repository was created 2026-05-26 and the latest metadata reports 119 stars, 29 forks, 3 open issues, and an unarchived public repository ([GitHub repository metadata](https://api.github.com/repos/thebriangao/totem)). These are signals of a young, visible project—not evidence of vendor support or long-term compatibility.

### Release and CI signals

The release API shows v1.5.1 as the latest published release, with v1.5.0 preceding it. The v1.5.1 notes describe a password-quoting/prompt bug and 30 added tests; v1.5.0 describes cache, session, timezone, write-validation, and auth-resilience work ([GitHub releases](https://api.github.com/repos/thebriangao/totem/releases?per_page=20), [v1.5.1 release](https://github.com/thebriangao/totem/releases/tag/v1.5.1), [v1.5.0 release](https://github.com/thebriangao/totem/releases/tag/v1.5.0)).

CI runs Node 24, `npm ci`, TypeScript typecheck, Vitest, and a build on pushes to `main` and pull requests ([Totem CI](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/.github/workflows/ci.yml)). The source is active and test-oriented, but the workflow uses floating `actions/checkout@v4` and `actions/setup-node@v4`, and the repository has no vendor SLA, formal private-API compatibility guarantee, or official WHOOP relationship. The README calls maintenance “best-effort” ([Totem contributor guide](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/CONTRIBUTING.md), [Totem FAQ](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/README.md)).

A practical dependency decision should therefore account for both ordinary open-source supply-chain risk and private-API breakage/account risk. Pinning a Totem version would make software behavior reproducible, but it would not make the upstream WHOOP surface stable.

## 11. Comparison with the current `whoop-cli` public read-only contract

| Dimension | `whoop-cli` today | Totem | Implication |
|---|---|---|---|
| Product contract | Go CLI/reusable client, explicitly public WHOOP account-data API, read-only | Node MCP with private iOS API, reads + writes | Do not merge Totem's private semantics into the public client. |
| Authentication | OAuth authorization code; developer client ID/secret; registered localhost callback; `offline` refresh | WHOOP email/password + MFA through private Cognito proxy; app-session bearer/refresh tokens | Keep OAuth; never ask `whoop-cli` users for WHOOP passwords. |
| Scopes | Six documented read scopes plus `offline` | No public OAuth scopes; app-session authority | Public scopes are least privilege and auditable. |
| Endpoints | `/developer/v2/user/profile/basic`, body, cycle, recovery, sleep, workout, with documented pagination | 311 captured private operations across 47 microservices plus a few public paths | Public list is smaller but supported/documented. |
| Writes | None by policy and schema (`read_only: true`) | 15 writes plus raw escape hatch | Preserve no-write invariant. |
| Models | Typed Go structs for profile/body/cycle/recovery/sleep/workout, nullable score pointers and string timestamps | Projection-specific flattened Zod schemas from unstable nested BFF payloads | Totem's projection pattern is a design lesson, not a contract to copy. |
| Transport | Go `net/http`; 30s context timeout; documented bearer; retries 429/5xx; no cache | Fetch/undici; 30s abort; private iOS headers; cache, session transport, MCP stdio/HTTP | Public client should keep simple documented transport and improve observability only where needed. |
| Pagination | Follows `next_token`, rejects repeated tokens | Endpoint-specific calls; many private paths use offsets/dates; cache/coalescing | Totem's loop protection is conceptually reusable; endpoint semantics are not. |
| Error behavior | Redacted API errors, request IDs, `Retry-After` parsing, retry on 429/5xx | 401 auth-expired, 5xx server class, projection errors on schema mismatch | Public client has stronger documented retry semantics; retain them. |
| Credential storage | macOS Keychain; `0600` file fallback; diagnostic command redacts values | `.env`/memory token stores; remote host secrets | CLI's Keychain-first model is safer for a public developer app. |
| Automation | Stable JSON/plain output, schema command, exit codes, read-only/safety profile | MCP structured tool outputs; remote connector authorization | Keep executable machine contract separate from MCP. |

Evidence for the CLI side is in [`whoop/client.go`](../../whoop/client.go), [`whoop/errors.go`](../../whoop/errors.go), [`internal/auth/oauth.go`](../../internal/auth/oauth.go), [`internal/auth/store.go`](../../internal/auth/store.go), [`docs/auth.md`](../../docs/auth.md), [`internal/cli/schema.go`](../../internal/cli/schema.go), and [`docs/commands/README.md`](../commands/README.md). Evidence for Totem is in its pinned client, auth, registry, and tool reference ([client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts), [Cognito](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/cognito.ts), [registry](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/tools/register.ts), [tools](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/TOOLS.md)).

## 12. Concrete recommendations for `whoop-cli`

### Must not do

1. **Do not implement Totem's Cognito login.** No private `/auth-service/v3/whoop/`, no email/password collection, no MFA relay, and no app-session token storage. Keep OAuth authorization-code flow and public scopes ([WHOOP OAuth docs](https://developer.whoop.com/docs/developing/oauth/), [`docs/auth.md`](../../docs/auth.md)).
2. **Do not copy iOS identity headers or attempt traffic camouflage.** The Totem code explicitly frames those headers as indistinguishable-app camouflage ([Totem device headers](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/device.ts)); this conflicts with the official Terms' documented-access/no-masking language ([WHOOP API Terms §1.4](https://developer.whoop.com/api-terms-of-use/)).
3. **Do not add undocumented endpoints, private writes, raw escape hatches, or hidden metric/profile mutations.** They would expand privilege, legal exposure, and maintenance scope without a public contract ([WHOOP API Terms §§1.4, 3.1, 7](https://developer.whoop.com/api-terms-of-use/)).
4. **Do not market parity with the WHOOP app.** Keep the README statement that this is a read-only public API client and distinguish documented account data from device control and private app features ([`README.md`](../../README.md)).

### Safe, public-surface improvements

1. **Add documented v2 detail lookups where they improve CLI utility.** The current client has collection reads; the official reference also documents cycle, sleep, and workout detail and cycle relationships. Add only after checking the current OpenAPI schema and preserving read-only scopes ([WHOOP API reference](https://developer.whoop.com/api/), [OpenAPI JSON](https://api.prod.whoop.com/developer/doc/openapi.json)).
2. **Consider a one-time v1 activity-ID migration command.** WHOOP documents `/v1/activity-mapping/{activityV1Id}` as a migration utility; keep it explicit and avoid making it part of ordinary sync ([WHOOP migration guide](https://developer.whoop.com/docs/developing/v1-v2-migration/), [API reference](https://developer.whoop.com/api/)).
3. **Consider opt-in public webhooks.** WHOOP documents v2 update/delete events for sleep, recovery, and workouts, signed with HMAC, with retries; a future CLI companion could validate signatures and fetch changed resources, but it should not weaken the CLI's read-only command contract ([WHOOP webhooks](https://developer.whoop.com/docs/developing/webhooks/)).
4. **Improve public rate-limit observability.** Preserve existing 429/5xx retry behavior, expose parsed `X-RateLimit-*` headers in debug-safe diagnostics, and honor documented limits of 100/minute and 10,000/day ([WHOOP rate limits](https://developer.whoop.com/docs/developing/rate-limiting/), [`whoop/errors.go`](../../whoop/errors.go)). Do not use Totem's undocumented “stealth” or burst-shaping techniques.
5. **Document model boundaries.** Explicitly state that public sleep gives stage totals rather than Totem's reconstructed hypnogram, public API does not provide continuous HR, and scores may be pending/null ([WHOOP support FAQ](https://developer.whoop.com/docs/developing/support/#does-the-api-offer-access-to-continuous-heart-rate-data), [`README.md`](../../README.md)).
6. **Borrow reliability patterns without private behavior.** Single-flight token refresh, bounded/careful caching, strict output validation, and explicit stale-data semantics are worthwhile abstractions only if applied to documented public endpoints and terms-compliant retention ([Totem token manager](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/token_manager.ts), [Totem client](https://raw.githubusercontent.com/thebriangao/totem/f9ac7f18198ccc1ea3a21e4d6bdc83733b4ef702/src/whoop/client.ts), [WHOOP API Terms §4.2](https://developer.whoop.com/api-terms-of-use/)).
7. **Maintain a public-only compatibility policy.** Pin the OpenAPI document/version used to review models, link every endpoint to official docs, treat undocumented responses as unsupported, and state that WHOOP may change APIs without compatibility/support obligations ([WHOOP API Terms §1.11](https://developer.whoop.com/api-terms-of-use/)).
8. **Keep credential handling least-privilege.** Continue Keychain/0600 storage, never accept tokens on command lines, never print secrets, and make `config diagnose` presence-only ([`internal/auth/store.go`](../../internal/auth/store.go), [`docs/auth.md`](../../docs/auth.md)).

### Product decision

If users request a `--private-api` or “full WHOOP app” mode, decline that scope for this repository rather than hiding it behind an experimental flag. A flag would not solve the underlying password/token, terms, privacy, account-suspension, and maintenance risks; it would also contradict the current machine-readable `read_only: true` schema contract ([`internal/cli/schema.go`](../../internal/cli/schema.go), [WHOOP API Terms](https://developer.whoop.com/api-terms-of-use/)). If a future project wants private access, it should be a separately named, explicitly unofficial project with its own security/legal review and no shared credential store.

## Bottom line

Totem is valuable comparative evidence: it shows how much richer the iOS app's private BFF surface is and demonstrates strong projection, validation, caching, and MCP safety patterns. It also demonstrates why those capabilities are not a safe extension of `whoop-cli`: private Cognito login, iOS-header impersonation, unstable BFF models, write authority, health-data exposure, and documented WHOOP terms risk are all materially different from the public OAuth read-only contract. The durable path for `whoop-cli` is a small, supported, auditable public client with better coverage of official endpoints—not private API parity.
