# Changelog

## [6.0.1](https://github.com/roadrunner-server/config/compare/v6.0.0...v6.0.1) (2026-11-27)

### 🐛 Bug Fixes

- **Included configuration merge**: Apply included leaf keys individually instead of replacing entire top-level sections. Main-only keys, earlier-include keys, and their environment-expanded values remain available when decoding shared sections with `UnmarshalKey`. See [BUG #2445](https://github.com/roadrunner-server/roadrunner/issues/2445) and [PR #121](https://github.com/roadrunner-server/config/pull/121).
