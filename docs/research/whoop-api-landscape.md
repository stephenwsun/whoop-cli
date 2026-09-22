# WHOOP API landscape

**Research snapshot:** 2026-09-22. This report uses WHOOP-owned sources only: the Developer Portal, its published OpenAPI document, WHOOP Engineering announcements, and WHOOP support/terms pages.

## Executive summary

The public WHOOP Developer Platform is a REST API for building apps and integrations around member data. WHOOP documents OAuth 2.0, webhooks, an API reference, and an app dashboard; the standard user-data flow requires a member to authorize the app. ([Developer Platform](https://developer.whoop.com/docs/introduction/), [Overview](https://developer.whoop.com/docs/developing/overview/), [OAuth 2.0](https://developer.whoop.com/docs/developing/oauth/))

The current published OpenAPI document exposes a v2 user-data API, a v1-to-v2 activity-ID migration utility, and a separate Trusted Partner API. The standard member API is read-oriented, apart from deleting/revoking the current user's access grant. ([OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json))

## Current public surface

### Standard member OAuth API

The following is the current standard, user-authorized surface in the official OpenAPI reference. Collection endpoints are paginated and accept time bounds; the reference sets a default `limit` of 10 and a maximum of 25 for the collections. ([OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json), [Pagination](https://developer.whoop.com/docs/developing/pagination))

| Capability | Current v2 endpoints and data | OAuth scope / source |
| --- | --- | --- |
| Basic profile | `GET /v2/user/profile/basic`; user ID, name, and email | `read:profile` — [API reference](https://developer.whoop.com/api/#tag/User/operation/getProfileBasic), [User data](https://developer.whoop.com/docs/developing/user-data/user) |
| Body measurements | `GET /v2/user/measurement/body`; height, weight, and WHOOP-calculated maximum heart rate | `read:body_measurement` — [API reference](https://developer.whoop.com/api/#tag/User/operation/getBodyMeasurement), [User data](https://developer.whoop.com/docs/developing/user-data/user) |
| Physiological cycles | `GET /v2/cycle` and `GET /v2/cycle/{cycleId}`; cycle timing, timezone, score state, strain, kilojoules, and average/max heart rate | `read:cycles` — [API reference](https://developer.whoop.com/api/#tag/Cycle), [Cycle model](https://developer.whoop.com/docs/developing/user-data/cycle) |
| Cycle relationships | `GET /v2/cycle/{cycleId}/sleep` and `GET /v2/cycle/{cycleId}/recovery` | Relationship endpoints are documented in the [API reference](https://developer.whoop.com/api/#tag/Cycle) and [migration guide](https://developer.whoop.com/docs/developing/v1-v2-migration) |
| Sleep | `GET /v2/activity/sleep` and `GET /v2/activity/sleep/{sleepId}`; naps, sleep stages, sleep need, respiratory rate, performance, consistency, and efficiency | `read:sleep` — [API reference](https://developer.whoop.com/api/#tag/Sleep), [Sleep model](https://developer.whoop.com/docs/developing/user-data/sleep) |
| Recovery | `GET /v2/recovery` and `GET /v2/cycle/{cycleId}/recovery`; score state, Recovery score, HRV, resting heart rate, and (where available) SpO2 and skin temperature | `read:recovery` — [API reference](https://developer.whoop.com/api/#tag/Recovery), [Recovery model](https://developer.whoop.com/docs/developing/user-data/recovery) |
| Workouts | `GET /v2/activity/workout` and `GET /v2/activity/workout/{workoutId}`; sport, strain, heart rate, energy, recording percentage, distance/altitude where available, and heart-rate-zone durations | `read:workout` — [API reference](https://developer.whoop.com/api/#tag/Workout), [Workout model](https://developer.whoop.com/docs/developing/user-data/workout) |
| Legacy activity-ID migration | `GET /v1/activity-mapping/{activityV1Id}` returns a v2 UUID for a historical v1 activity ID; WHOOP says this is for one-time migration, not recurring production lookups | No OAuth scope is listed for this utility in the OpenAPI document — [Migration guide](https://developer.whoop.com/docs/developing/v1-v2-migration), [OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json) |
| User de-authorization | `DELETE /v2/user/access` revokes the user's access grant and stops configured webhooks for that user | OAuth-authenticated, no data scope listed — [API reference](https://developer.whoop.com/api/#tag/User/operation/revokeUserOAuthAccess), [OAuth 2.0](https://developer.whoop.com/docs/developing/oauth#revoking-an-access-token) |

WHOOP's v2 migration guide describes more consistent paths, stronger typing, UUID identifiers for sleep/workout activities, improved pagination/error handling, and the cycle-to-sleep relationship. It says v1 is no longer supported and future features are to be added to v2 first. ([v1 to v2 Migration Guide](https://developer.whoop.com/docs/developing/v1-v2-migration))

The API changelog records the v2 launch on **2025-07-01**, Strength Trainer availability through `/workout` on **2024-05-01**, the de-authorization endpoint on **2023-02-01**, and the Developer Platform launch on **2022-09-01**. The launch entry links WHOOP's original member-facing announcement about integrations and data export. ([API Changelog](https://developer.whoop.com/docs/api-changelog/), [WHOOP launch announcement](https://www.whoop.com/thelocker/access-your-whoop-data-with-new-integrations-data-export-options/))

### OAuth and token lifecycle

The documented user flow is OAuth 2.0 authorization code: authorization URL `https://api.prod.whoop.com/oauth/oauth2/auth`, token/refresh URL `https://api.prod.whoop.com/oauth/oauth2/token`, and a registered redirect URI. WHOOP recommends a state value for CSRF protection and says a self-generated state must be eight characters long. ([OAuth 2.0](https://developer.whoop.com/docs/developing/oauth))

The data scopes in the API reference are `read:profile`, `read:body_measurement`, `read:cycles`, `read:recovery`, `read:sleep`, and `read:workout`. The `offline` scope is required to receive a refresh token. Access tokens are short-lived and expose `expires_in`; refreshing invalidates both the old access token and the old refresh token, so clients must persist the replacement refresh token. ([API reference authentication](https://developer.whoop.com/api/#section/Authentication), [OAuth 2.0](https://developer.whoop.com/docs/developing/oauth))

### Webhooks

WHOOP's current webhook documentation lists six event types: `workout.updated`, `workout.deleted`, `sleep.updated`, `sleep.deleted`, `recovery.updated`, and `recovery.deleted`. An update event also represents creation for workouts, sleeps, and recoveries. Webhooks are notifications rather than payloads containing the changed data; the recipient must fetch the current resource from the API. ([Webhooks](https://developer.whoop.com/docs/developing/webhooks/#webhook-event-types), [Webhooks FAQ](https://developer.whoop.com/docs/developing/webhooks/#webhooks-faqs))

v2 webhooks use UUID activity IDs and are the default model. v2 recovery webhook IDs are the UUID of the associated sleep. WHOOP's changelog dated **2025-11-01** says v1 webhooks are no longer being published and documents the v1 activity-ID mapping endpoint; the portal banner also directs developers to the v1-to-v2 migration guide. ([Webhooks](https://developer.whoop.com/docs/developing/webhooks/#webhook-model-versions), [API Changelog](https://developer.whoop.com/docs/api-changelog/), [Migration Guide](https://developer.whoop.com/docs/developing/v1-v2-migration))

Webhook requests are signed with `X-WHOOP-Signature` and `X-WHOOP-Signature-Timestamp`; WHOOP documents HMAC-SHA256 over the timestamp concatenated with the raw body. WHOOP retries failed delivery five times over about one hour, so its docs recommend a fast 2xx response, signature validation, and periodic reconciliation through the API. ([Webhooks security and delivery](https://developer.whoop.com/docs/developing/webhooks/#webhooks-security), [WHOOP Engineering webhook announcement, 2022-10-29](https://engineering.prod.whoop.com/dev-platform-3/))

The FAQ explicitly says there are no webhooks at present for Day Strain, cycles, or body measurements; those must be retrieved through their APIs. ([Webhooks FAQ](https://developer.whoop.com/docs/developing/webhooks/#webhooks-faqs))

### Trusted Partner API (separate from member API)

The portal now documents a Healthcare Partner API that is **only for WHOOP-approved healthcare partners**, not general developers. It uses server-to-server Trusted Partner OAuth 2.0 client credentials and the `whoop-partner/token` scope, with no end-user login or consent. ([Partner Overview](https://developer.whoop.com/docs/partner/overview), [Partner Authentication](https://developer.whoop.com/docs/partner/authentication))

Its documented resources are WHOOP-created lab requisitions and service requests: partners can retrieve requisitions/service requests, update service-request statuses, and submit diagnostic-report results. A non-production test-data endpoint is also documented. This is a healthcare workflow, not a member-account or wearable-control API. ([Partner Overview](https://developer.whoop.com/docs/partner/overview), [Lab Requisitions](https://developer.whoop.com/docs/partner/lab-requisitions), [Service Requests](https://developer.whoop.com/docs/partner/service-requests), [OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json))

## Explicit limits and unsupported capabilities

- **No direct device control is documented.** The current public OpenAPI standard user surface contains reads for account data plus access revocation; it contains no endpoint for changing device settings, starting/stopping a recording, commanding a wearable, or writing member metrics. This is an inventory finding from the published reference, not a claim about private/internal APIs. ([OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json))
- **Continuous heart-rate export is explicitly unavailable through the API.** WHOOP says devices can broadcast heart rate over BLE to another device, but continuous heart-rate data is not available through the WHOOP API. ([Developer Support FAQ](https://developer.whoop.com/docs/developing/support/#does-the-api-offer-access-to-continuous-heart-rate-data))
- **No sandbox for developers without WHOOP is offered.** WHOOP requires developers on the platform to have a WHOOP device; the API is currently free, but a WHOOP membership/device is required. ([Developer Support FAQ](https://developer.whoop.com/docs/developing/support/#do-you-offer-a-sandbox-environment-for-developers-who-do-not-have-a-whoop), [FAQ](https://developer.whoop.com/docs/developing/support/#are-there-any-costs-associated-with-using-the-whoop-api))
- **Default rate limits are 100 requests/minute and 10,000 requests/day.** WHOOP documents `429` responses and a dashboard request process for increases. ([Rate Limiting](https://developer.whoop.com/docs/developing/rate-limiting))

## Roadmap / coming-soon status

I found **no public, dated, named “coming soon” API capability or public roadmap** in the reviewed Developer Portal pages, OpenAPI document, API changelog, or WHOOP Engineering developer-platform announcements. WHOOP's support answer is intentionally non-specific: it says the API will be updated as functionality is released, points developers to the changelog, and accepts feature requests through the dashboard. ([Developer Support FAQ — future improvements](https://developer.whoop.com/docs/developing/support/#what-are-the-future-improvement-plans-for-the-whoop-api), [API Changelog](https://developer.whoop.com/docs/api-changelog/), [Developer Portal sitemap](https://developer.whoop.com/sitemap.xml), [WHOOP Engineering developer-platform archive](https://engineering.prod.whoop.com/archive/))

The only forward-looking API statement found is the v1-to-v2 migration guidance that new features/functionality will be added to v2 first; it names no feature or delivery date. Treat that as a migration direction, not a roadmap commitment. ([Migration Guide — deprecation timeline](https://developer.whoop.com/docs/developing/v1-v2-migration#deprecation-timeline))
WHOOP's API Terms reserve the right to add, modify, or remove API portions and do not promise compatibility with every update; they also say API access is currently free but would require prior notice before WHOOP starts charging. This is a change/pricing policy, not a roadmap commitment. ([API Terms of Use](https://developer.whoop.com/api-terms-of-use/#1-use-of-whoop-apis))

WHOOP's official **2022-10-27** engineering announcement explains the platform's design priorities—stable, documented external interfaces, a dedicated gateway, and webhooks—and links developers to the platform. The **2022-10-29** announcement explains webhook retries and reconciliation. These are launch/design announcements, not promises of future endpoints. ([Designing the WHOOP API, 2022-10-27](https://engineering.prod.whoop.com/dev-platform/), [Designing a resilient webhook system, 2022-10-29](https://engineering.prod.whoop.com/dev-platform-3/))

## Justified future `whoop-cli` issue candidates

1. **Expose v2 detail lookups.** The CLI currently covers collection reads; add client/CLI access to the documented cycle, sleep, workout, and cycle-related sleep/recovery detail endpoints where useful. ([API reference](https://developer.whoop.com/api/))
2. **Add one-time legacy-ID migration tooling.** A command that resolves stored v1 sleep/workout IDs through `/v1/activity-mapping/{activityV1Id}` and records the UUID would follow WHOOP's documented migration-only use, without making the mapping call part of normal sync. ([Migration Guide](https://developer.whoop.com/docs/developing/v1-v2-migration#2-lookup-v2-ids-from-v1-ids))
3. **Consider an opt-in webhook-assisted refresh mode.** A small receiver/sync workflow could consume the six supported v2 events, validate signatures, fetch changed resources, and reconcile missed events. This should remain separate from the read-only command path. ([Webhooks](https://developer.whoop.com/docs/developing/webhooks/))
4. **Do not file a device-control or continuous-HR feature issue as an API implementation plan yet.** The official surface does not document it, and WHOOP explicitly says continuous HR is unavailable via the API; revisit only if a future changelog or reference entry names such a capability. ([OpenAPI specification](https://api.prod.whoop.com/developer/doc/openapi.json), [Developer Support FAQ](https://developer.whoop.com/docs/developing/support/#does-the-api-offer-access-to-continuous-heart-rate-data))
