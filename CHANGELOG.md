# Changelog

All notable changes to the Mundo Tá Lendo 2026 project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.7] - 2026-01-06

### Added
- **New Backend Endpoint:** `GET /readings/{iso3}` - Returns all readings for a specific country
  - Lambda function queries DynamoDB: `PK = "EVENT#LEITURA"` filtered by `iso3`
  - Returns all readings with progress >= 1% (in-progress + completed)
  - Sorting: completed books first (100%), then by `updatedAt` DESC
  - Timeout: 30s, Memory: 256MB, Concurrency: 10
- **Custom React Hook:** `useCountryReadings()` for fetching country readings
  - Manages loading, error, and readings states
  - Async `fetchReadings(iso3)` function
- **Makefile Command:** `make readings iso3=XXX` to test readings endpoint
  - Supports both DEV and PROD: `make readings iso3=BRA STAGE=prod`
  - Validates required `iso3` parameter
  - Auto-fetches API key
- **Unit Tests:** 4 new Go tests for readings Lambda
  - `TestBuildResponse` - sorting logic validation
  - `TestIsAlpha` - ISO3 code validation
  - `TestBuildResponseEmptyInput` - empty array handling
  - `TestBuildResponseSortingByProgressOnly` - progress sorting

### Changed
- **CountryPopup Component:** Complete redesign (168 lines)
  - **Loading State:** Animated spinner while fetching data
  - **Error State:** Detailed error message display
  - **Progress Bars:** Visual progress indicators
    - Blue (#3B82F6) for in-progress books
    - Green (#10B981) for completed books (100%)
  - **Completed Badge:** Green checkmark "✓ Completo" for 100% books
  - **Book Covers:** 16x24px thumbnails with 📚 emoji fallback
  - **Layout:** Horizontal layout with Avatar (12x12) + Cover (16x24) + Details
  - **Empty State:** "Nenhuma leitura encontrada" message when no readings
  - **Scrollable:** List with max-height 384px for many readings
- **Map.jsx Click Handler:** Now shows popup for ALL colored countries (not just GPS markers)
  - Removes restriction `if (readers.length > 0)`
  - Checks if country has `progress >= 1%` (is colored on map)
  - Opens popup immediately with loading spinner
  - Fetches readings asynchronously from `/readings/{iso3}`
  - Updates popup via useEffect when data arrives
- **CLAUDE.md:** Updated with v1.0.7 changelog and detailed documentation
- **package.json:** Version bump from 1.0.6 to 1.0.7

### Fixed
- **Case Sensitivity Bug:** DynamoDB FilterExpression now uses lowercase field names
  - Changed from `ISO3 = :iso3 AND Progresso >= :minProgress`
  - To `iso3 = :iso3 AND progresso >= :minProgress`
  - Matches actual DynamoDB schema where fields are lowercase
- **useEffect Circular Dependency:** Fixed infinite re-render issue
  - Removed `popup` from dependencies array
  - Only updates popup when `!readingsLoading`
- **Loading State Bug:** Fixed infinite loading when country has no readings
  - Now updates popup state even when `readings` array is empty
  - Loading state properly transitions to empty state

### Technical Details
- **New Files Created:**
  - `packages/functions/readings/main.go` (159 lines)
  - `packages/functions/readings/go.mod`
  - `packages/functions/readings/go.sum`
  - `packages/functions/readings/main_test.go` (4 tests)
  - `src/hooks/useCountryReadings.js` (40 lines)
- **Modified Files:**
  - `sst.config.ts` (lines 148-160)
  - `src/components/Map.jsx` (multiple changes)
  - `src/components/CountryPopup.jsx` (complete rewrite)
  - `Makefile` (lines 289-308)
  - `CLAUDE.md` (v1.0.7 section)
  - `package.json` (version bump)
- **Lines Changed:** +653 / -77
- **Tests:** 30 Go unit tests passing (26 previous + 4 new)
- **Build:** Next.js compilation successful (1.2s)

### Compatibility
- ✅ 100% Backward compatible - no breaking changes
- ✅ Old data works with `CapaURL` fallback for missing book covers
- ✅ GPS markers remain functional on map
- ✅ No breaking changes in existing endpoints or data structures
- ✅ Read-only feature - zero impact on existing data

### Performance
- **DynamoDB Query:** <500ms typical (index PK + FilterExpression)
- **Network Latency:** ~200ms (API Gateway + Lambda)
- **Total Expected:** <1s from click to display
- **On-Demand:** Queries only executed when user clicks on a country
- **No Pagination:** Initial implementation (rarely >100 readings per country)

---

## [1.0.6] - 2026-01-05

### Added
- Country Popup with readers list on map click
- Book cover images (`CapaURL` field)
- Webhook extracts covers from `vinculados[].edicao.capa`
- Migration Lambda `/migrate` to populate covers in old data

### Changed
- CountryPopup displays book covers with fallback to book icon
- Webhook automatically extracts and saves `capaURL`

---

## [1.0.5] - 2026-01-03

### Fixed
- CORS Proxy Bug: Avatars not loading in deployed DEV environment

---

## [1.0.4] - 2026-01-03

### Changed
- User Markers Layout: Concentric circles instead of horizontal line

---

## [1.0.3] - 2025-12-25

### Fixed
- Critical Bug: PK Mismatch, SK overwrites, payload deletion

### Added
- `updatedAt` field, GPS markers filter, force rebuild, STAGE support

---

## [1.0.2] - 2025-12-25

### Added
- UUID Architecture, auto-cleanup, storage optimization

---

## [1.0.1] - 2025-12-23

### Added
- User GPS markers on map, `/users/locations` endpoint

---

## [1.0.0] - 2025-12-21

### Added
- Initial release with interactive map, webhook integration, Lambda functions
