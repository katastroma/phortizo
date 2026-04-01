# Changelog
All notable changes to this project will be documented in this file. See [conventional commits](https://www.conventionalcommits.org/) for commit guidelines.

- - -
## 0.6.5 - 2026-04-01
#### Bug Fixes
- better logging - (9031ec7) - Nathan Blair

- - -

## 0.6.4 - 2026-04-01
#### Bug Fixes
- use rendertype metadata key from keleustes - (088e7b5) - Nathan Blair

- - -

## 0.6.3 - 2026-04-01
#### Bug Fixes
- deduplicate tenant attribute - (0d69d21) - Nathan Blair
#### Tests
- better test hygiene - (fb6d867) - Nathan Blair
#### Refactoring
- do detection in source handler - (fde68e9) - Nathan Blair

- - -

## 0.6.2 - 2026-03-31
#### Bug Fixes
- use client-streaming - (8f0cb7b) - Nathan Blair

- - -

## 0.6.1 - 2026-03-31
#### Bug Fixes
- FaF but correctly - (815578d) - Nathan Blair
- wait for render stream to finish reading - (322b285) - Nathan Blair
#### Miscellaneous Chores
- add more logging - (8459926) - Nathan Blair

- - -

## 0.6.0 - 2026-03-30
#### Features
- renderer now handles renderer detection - (84e9dfe) - Nathan Blair
#### Bug Fixes
- make render paths relative to root - (ddae360) - Nathan Blair
- use tar archive to stream data out to renderer - (c192e80) - Nathan Blair

- - -

## 0.5.3 - 2026-03-30
#### Bug Fixes
- use new source target phrasing - (d9fa706) - Nathan Blair
#### Documentation
- update variable usage and terminology - (b695452) - Nathan Blair
#### Refactoring
- webhook -> push - (fbf561f) - Nathan Blair
- use closure pattern - (581167b) - Nathan Blair
- use MockObject in credential reader test - (934eb78) - Nathan Blair
- object package is generic - (3af0786) - Nathan Blair
- k8s package now only knows about k8s - (928ffa6) - Nathan Blair
- fully decouple k8s configmap from source target - (7f061ee) - Nathan Blair
- WIP while removing coupling of "watch targets" from k8s/configmap package - (1d5ac18) - Nathan Blair
- use store interface for lease - (7080284) - Nathan Blair
- break up tracing package - (0990d1a) - Nathan Blair
- rip all interfaces out of pipeline - (4b04a74) - Nathan Blair
- move credentials to appropriate package - (bf91fbe) - Nathan Blair
- rename auth package - (006ac3c) - Nathan Blair
- rename Credential interface - (9e3999e) - Nathan Blair
- more source package renames - (d3e5b88) - Nathan Blair
- move find to webhook as match - (67bcc00) - Nathan Blair
- move webhook secret knowledge to webhook domain - (db1df16) - Nathan Blair
- event traces own multiple watch target spans - (79aef6b) - Nathan Blair
#### Miscellaneous Chores
- fix import in main - (d1b8818) - Nathan Blair
- TEMPO_ADDRESS is required to support replayability - (ee291a8) - Nathan Blair
- support b64 encoded keys - (f47d8ae) - Nathan Blair
- don't care what go thinks, periods in comments are weird - (3b94b70) - Nathan Blair
- allow listing configmap resources - (6eb3641) - Nathan Blair
- refactor watchtarget to its own package - (2821b10) - Nathan Blair

- - -

## 0.5.2 - 2026-03-27
#### Bug Fixes
- PR workflow - (f7937c2) - Nathan Blair
#### Documentation
- use org CONTRIBUTING.md - (794c37f) - Nathan Blair
#### Continuous Integration
- add PR workflow - (ac44a64) - Nathan Blair

- - -

## 0.5.1 - 2026-03-26
#### Bug Fixes
- (**cd**) don't trigger on ignored files - (06326ca) - Nathan Blair

- - -

## 0.5.0 - 2026-03-26
#### Features
- phortizo ready for integration testing - (8d1d928) - Nathan Blair
#### Bug Fixes
- (**cd**) cog handles CD triggering - (7565cf3) - Nathan Blair
#### Refactoring
- (**structure**) rename some files - (5396a60) - Nathan Blair
- use credential registration approach - (4e15364) - Nathan Blair
- package name change - (a8ddc18) - Nathan Blair
- fail wrapper for webhook handling - (d1bd954) - Nathan Blair
- more restructuring and code cleanup - (308cbf9) - Nathan Blair
- extract auth resolution - (25a89b2) - Nathan Blair
- cleanup handleMatch for readability - (62c8d52) - Nathan Blair
- cleanup Find for readability - (bbf30f9) - Nathan Blair

- - -

## 0.4.0 - 2026-03-26
#### Features
- implement platform health checking again - (aa1932e) - Nathan Blair

- - -

## 0.3.0 - 2026-03-26
#### Features
- initial release - (4dd5d2f) - Nathan Blair
#### Documentation
- update docs - (0bff9a0) - Nathan Blair
#### Tests
- fix webhook tests to use k8s interface - (fa02772) - Nathan Blair
- add tests for apps package - (e03ef91) - Nathan Blair
- cleanup - (bb2b962) - Nathan Blair
- implement credential tests - (f112dc3) - Nathan Blair
- add tests for tempo - (fff81c9) - Nathan Blair
#### Refactoring
- eliminate the useless and confusing Result type - (c287353) - Nathan Blair
- further leveraging the github SDK - (0eec533) - Nathan Blair
- leverage github SDK more for push event data - (3e99588) - Nathan Blair
- consolidate streaming - (be50296) - Nathan Blair
- read webhook secret from k8s - (2167c84) - Nathan Blair
- discoverability cleanup - (a04a8ec) - Nathan Blair
- this is looking much better but still WIP - (97ecf4f) - Nathan Blair
- more WIP - (dc4cadd) - Nathan Blair
- FUNDAMENTALLY BROKEN STILL - (3f55bf2) - Nathan Blair
- ergonomics and semantics across codebase - (3108d5f) - Nathan Blair
- break up the handlers package - (3da4a1d) - Nathan Blair
#### Miscellaneous Chores
- (**regression**) fix attribute names for watch targets - (47fc618) - Nathan Blair
- implement retrieve - (b97c16a) - Nathan Blair
- add GetWatchTarget - (bd222ed) - Nathan Blair
- go mod update - (a6e90d9) - Nathan Blair
- implement leasing - (07c24e3) - Nathan Blair
- work on implementing replay - (3d5796e) - Nathan Blair
- implement pipeline/match handler/runner tests - (75f22b7) - Nathan Blair
- wire up retrieval of watch targets - (993e34d) - Nathan Blair
- cleaning up more stale references - (f6e8fe9) - Nathan Blair
- consolidate multiple names with what they actually are - (e3909ce) - Nathan Blair
- more cleanup - (7558c8f) - Nathan Blair
- support watchtarget configmap - (3357359) - Nathan Blair
- webhook secret plumbing done - (ee4b9cd) - Nathan Blair

- - -

## 0.2.0 - 2026-03-25
#### Features
- implement retriever and replayer - (0b0b27e) - Nathan Blair
#### Tests
- more test wiring - (c49fb60) - Nathan Blair
- improve tests - (304f0c1) - Nathan Blair
- add some initial tests - (61dbc24) - Nathan Blair
#### Miscellaneous Chores
- wire up streaming - (36686b4) - Nathan Blair

- - -

## 0.1.0 - 2026-03-24
#### Features
- initial release - (64ed1cc) - Nathan Blair
#### Documentation
- update docs - (47744c4) - Nathan Blair
- make phortizo purpose more clear - (91049f2) - Nathan Blair
- expand docs - (ace5ee1) - Nathan Blair
- update docs - (e2de678) - Nathan Blair
- add docs - (f7cfd11) - Nathan Blair
#### Continuous Integration
- reorder CI workflow - (5aeb9ae) - Nathan Blair
- add CI workflow - (e478283) - Nathan Blair
#### Refactoring
- use otel as coordination point - (dce91e9) - Nathan Blair
- committing progress before next major refactor - (542d27c) - Nathan Blair
#### Miscellaneous Chores
- go mod init - (a910c70) - Nathan Blair
- add mise and cog - (f8571cc) - Nathan Blair

- - -

Changelog generated by [cocogitto](https://github.com/cocogitto/cocogitto).