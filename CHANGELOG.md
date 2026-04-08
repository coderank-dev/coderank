# Changelog

## [0.1.3](https://github.com/coderank-dev/coderank/compare/v0.1.2...v0.1.3) (2026-04-08)


### Features

* add project wiki memory + rename fetch→query ([#45](https://github.com/coderank-dev/coderank/issues/45)) ([f9cac24](https://github.com/coderank-dev/coderank/commit/f9cac2488d9f4293655fbf43b5040a3240fc5247))


### Bug Fixes

* read Go version from go.mod in release workflow ([#43](https://github.com/coderank-dev/coderank/issues/43)) ([471d10c](https://github.com/coderank-dev/coderank/commit/471d10c10d6bc25e159ceda3df46fc27e0de9a11))

## [0.1.2](https://github.com/coderank-dev/coderank/compare/v0.1.1...v0.1.2) (2026-04-07)


### Features

* inject skills by default, --surface flag appends full surface ([#41](https://github.com/coderank-dev/coderank/issues/41)) ([125104c](https://github.com/coderank-dev/coderank/commit/125104c61f796a2c6f26e2d2cb1031e0344e6a99))

## [0.1.1](https://github.com/coderank-dev/coderank/compare/v0.1.0...v0.1.1) (2026-04-04)


### Features

* add AES-256-GCM encrypted offline cache ([#20](https://github.com/coderank-dev/coderank/issues/20)) ([ece087e](https://github.com/coderank-dev/coderank/commit/ece087e16d22ac947423905e1a95abffecb913ef))
* add agent detection and context writer with marker-based section management ([#21](https://github.com/coderank-dev/coderank/issues/21)) ([ac35ff8](https://github.com/coderank-dev/coderank/commit/ac35ff8451650c8659c8d4b9005cd8e46800da04))
* add automated releases with release-please and changelog generator ([#38](https://github.com/coderank-dev/coderank/issues/38)) ([5afdea5](https://github.com/coderank-dev/coderank/commit/5afdea50ef2aa6b6c72e96781377a59f1c353947))
* add coderank inject command with auto agent detection ([#22](https://github.com/coderank-dev/coderank/issues/22)) ([0674513](https://github.com/coderank-dev/coderank/commit/067451385bcfb235e8f6df44387b4a87abdc3d06))
* add file watcher for inject --watch mode and gitignore guidance ([#24](https://github.com/coderank-dev/coderank/issues/24)) ([0bf6038](https://github.com/coderank-dev/coderank/commit/0bf603890b4739b0a625583a1bb5089276bbd856))
* add login command with browser-based OAuth flow ([#25](https://github.com/coderank-dev/coderank/issues/25)) ([5b37320](https://github.com/coderank-dev/coderank/commit/5b37320f0b15de9e9f708cb73e85877813987071))
* add multi-agent skill installer with root skill + optional surfaces ([#34](https://github.com/coderank-dev/coderank/issues/34)) ([6ad60e4](https://github.com/coderank-dev/coderank/commit/6ad60e4077b1aea57524603942f2851b1ee38301))
* add query, topic, search, gotchas, topics commands (UOW_36) ([#26](https://github.com/coderank-dev/coderank/issues/26)) ([576c611](https://github.com/coderank-dev/coderank/commit/576c611cce980d584d9432ecc5a7eaeaba00b871))
* add relevance score (0-100%) to query results ([#29](https://github.com/coderank-dev/coderank/issues/29)) ([ed99170](https://github.com/coderank-dev/coderank/commit/ed9917051de15a688da58e2a262116a54b2900fb))
* add VHS demo tape for terminal recording ([#19](https://github.com/coderank-dev/coderank/issues/19)) ([b2818da](https://github.com/coderank-dev/coderank/commit/b2818dadbfa17075be5f9ee9937c7dabff67567c))
* enrich terminal output to match marketing page aesthetic ([fb5b340](https://github.com/coderank-dev/coderank/commit/fb5b3409b734f8c31fe6d66e77c714abb37ca59c))
* expand root skill with all CLI commands ([3abf483](https://github.com/coderank-dev/coderank/commit/3abf4833eeea874e5050506a0be46b1f4c2abdfb))
* **render:** rewrite health display with lipgloss table ([#40](https://github.com/coderank-dev/coderank/issues/40)) ([5e7cbdd](https://github.com/coderank-dev/coderank/commit/5e7cbdd11b350863481453b0919b6cff16e0cbac))


### Bug Fixes

* align health score bars using lipgloss width ([a951056](https://github.com/coderank-dev/coderank/commit/a9510564174e238ba300cab73bd8fb2a9f2b7e01))
* de-indent and strip language from 4-space indented fenced code blocks ([#33](https://github.com/coderank-dev/coderank/issues/33)) ([17ee3b6](https://github.com/coderank-dev/coderank/commit/17ee3b6e0b9aebdc0eb3faac2d45b210875c007f))
* hide token count in footer when zero ([#37](https://github.com/coderank-dev/coderank/issues/37)) ([b98f6cf](https://github.com/coderank-dev/coderank/commit/b98f6cf6b7bb613661f0906ce99f0ec4310ee40b))
* increase default max-tokens from 5000 to 10000 ([#31](https://github.com/coderank-dev/coderank/issues/31)) ([46491d8](https://github.com/coderank-dev/coderank/commit/46491d85bf62d8cb8258775bcd4147320603a403))
* remove redundant query command (use fetch instead) ([#27](https://github.com/coderank-dev/coderank/issues/27)) ([dea42fc](https://github.com/coderank-dev/coderank/commit/dea42fc747f90785f018414e130cb770d6cf0df2))
* replace ♻️ with 🔄 in health display ([bd72adc](https://github.com/coderank-dev/coderank/commit/bd72adcb4aa20f0b26a0348cd826a2226d37b9fa))
* respect api-url config in coderank install --with-surfaces ([be1e564](https://github.com/coderank-dev/coderank/commit/be1e564a4aac7a0670c3010722a07cb9ba7e63cc))
* skip rendering results with empty content in query output ([#30](https://github.com/coderank-dev/coderank/issues/30)) ([7b21388](https://github.com/coderank-dev/coderank/commit/7b213880d1d6fa2acac96fee0ac7c1f3934431fc))
* strip language identifiers from code fences before Glamour rendering ([#32](https://github.com/coderank-dev/coderank/issues/32)) ([9be0b49](https://github.com/coderank-dev/coderank/commit/9be0b49d41e2171a356d2c9757716c29efd593e6))
* strip YAML frontmatter before rendering query/topic output ([#36](https://github.com/coderank-dev/coderank/issues/36)) ([f0dd24c](https://github.com/coderank-dev/coderank/commit/f0dd24ced29459d1a8a7849ccd53ac9ae24df01f))
* upgrade GitHub Actions to Node.js 24 compatible versions ([#17](https://github.com/coderank-dev/coderank/issues/17)) ([f8bbfbc](https://github.com/coderank-dev/coderank/commit/f8bbfbc3f5a20bb8429e1fe15bcf8b411f432635))
* use coderank query (not fetch) in root skill ([ff9f763](https://github.com/coderank-dev/coderank/commit/ff9f7639f3e93859724bacfc9c24913cd4041acc))
