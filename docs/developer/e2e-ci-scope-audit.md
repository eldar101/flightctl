# E2E CI scope audit for GitHub Actions and Jenkins

This document compares the current Go E2E scope in GitHub Actions and Jenkins, identifies overlap and waste, and recommends a clearer split between merge-safety coverage and comprehensive validation.

The primary optimization target is time-to-main, with GitHub `merge_group` scope as the biggest controllable lever. Today, even small changes can wait about 1.5 hours because `merge_group` runs too much E2E coverage in GitHub Actions.

## Table of contents

- [Terminology](#terminology)
- [Data reviewed](#data-reviewed)
- [Sources](#sources)
- [Summary and recommendation](#summary-and-recommendation)
- [Current state](#current-state)
- [Overlap and gaps](#overlap-and-gaps)
- [Measured bottlenecks](#measured-bottlenecks)
- [Recommended scope boundaries](#recommended-scope-boundaries)
- [Recommended merge-queue removals](#recommended-merge-queue-removals)
- [Concrete merge-queue proposal](#concrete-merge-queue-proposal)
- [Two implementation options](#two-implementation-options)
- [Success metrics](#success-metrics)
- [Implementation outline](#implementation-outline)

## Terminology

- `spec` = one individual Ginkgo test case.
- `bucket` = one CI test grouping defined by a label filter, such as non-agent sanity or agent slow.
- `node` = one parallel Ginkgo worker running part of a bucket.
- `job` = one CI execution unit in GitHub Actions or Jenkins.
- `matrix entry` = one specific GitHub workflow combination, such as an OS variant plus a label-filter bucket.
- `artifact` = saved CI output such as logs or JUnit XML results.
- `wall time` = the total elapsed clock time for a CI run or job.
- `queue wait` = the time a job spends waiting before it actually starts running.
- `slice` = the chosen subset of tests kept together for a specific CI purpose. In this audit, it usually means the tests kept in the required merge-queue path.
- `merge_group` = GitHub's merge queue run. This is the required CI path that must pass before code reaches `main`.
- `required merge-queue path` = the set of checks that block the merge queue. If a test is in this path, it can stop a PR from landing.
- `time-to-main` = the total time from a ready PR until the change is actually available in `main`.
- `sanity` = the current label used for the smaller, higher-priority E2E subset. It does not mean all important tests, only the tests currently tagged that way.
- `control-plane subset` = tests of the central service behavior rather than device-internal behavior. In this proposal that is `cli`, `configuration`, `resourcesync`, and `selectors`.
- `agent smoke` = a very small set of device-agent health tests. In this proposal that is the retained agent specs for device status, OCI prefetch, systemd reporting, update, and rollback.
- `non-agent` = tests that do not exercise device-agent behavior directly. In practice these are mostly service-side and control-plane flows.
- `agent fast` = agent-labeled tests that are not marked `slow`.
- `agent slow` = agent-labeled tests marked `slow`, usually because they involve longer device lifecycle behavior.
- `canary specs` = one or two very small representative tests kept only to catch obvious regressions early, not to provide broad coverage.
- `OS variant` = the device image under test, such as `cs9-bootc` or `cs10-bootc`.
- `broad validation lane` = the larger CI lane intended to catch more scenarios, more environments, and slower regressions. In this audit, that is Jenkins.
- `Jenkins ACM lane` = the main broad Jenkins validation path used in this audit.
- `specialized coverage` = tests for narrower or heavier areas such as backup and restore, TPM, quadlets, disconnected, external DB, or upgrade.
- `post-merge lane` = validation that runs after code reaches `main`, not before.
- `required check` = a CI result that must pass before the merge queue can continue.
- `dry-run discovery` = listing which tests would run without actually executing them.
- `optional PR validation` = a manually triggered GitHub workflow for a specific PR that runs extra tests without expanding the required merge queue.
- `selector` = the test filter used to decide which tests run, such as `sanity && !agent` or `!integration`.

## Data reviewed

This audit compares current GitHub workflow definitions, current Jenkins test selection, fresh Linux dry-run test discovery, recent GitHub merge-queue timing data, GitHub Actions job logs and JUnit artifacts, Jenkins console logs from saved runs, and a separate GitHub flaky-test review.

## Sources

This audit uses:

- `upstream/main` of `flightctl/flightctl` for GitHub workflows and `test/e2e`
- your local [`ocp-edge-ci/jenkins/ci-profiles-new/`](../../ocp-edge-ci/jenkins/ci-profiles-new/) copy for Jenkins profile configuration
- GitHub Actions run history and raw `merge_group` logs exported with `gh`
- local Jenkins `ocp-flightctl-gotests` console logs saved under [`eldar_audit_logs/`](../../eldar_audit_logs/)
- the downloaded GitHub flakiness report under [`eldar_audit_logs/github_flakiness/`](../../eldar_audit_logs/github_flakiness/)

The current repository snapshot used for workflow and discovery validation was commit `e364d4e9fd1490b197f9f67f18a37f974e51cb89`.

## Summary and recommendation

- GitHub Actions currently runs 68 Go E2E specs, all selected from `sanity` buckets.
- The main Jenkins ACM profile currently runs 143 Go E2E specs through the broader `!integration` selector.
- 68 specs run in both systems.
- 0 specs run only in GitHub Actions, while 75 run only in Jenkins.
- GitHub `merge_group` runs average 61.71 minutes wall time.
- The dominant GitHub bottleneck is the non-agent `sanity` leg at 57.11 minutes average.
- A separate GitHub flakiness review classified 31 of 32 failed GitHub runs as flakes, so merge-queue pain is being driven by both runtime and instability.
- Successful Jenkins ACM-family runs are much broader and typically average about 212 to 328 minutes of E2E runtime per run.
- Preferred recommendation: reduce GitHub merge-queue Go E2E from 68 specs to the proposed 30-spec slice and keep Jenkins as the broad validation lane.
- More aggressive alternative: if the real target is to hold the entire merge queue at 15 minutes or less, shrink the blocking slice further and rebalance or add GitHub test nodes.

### Blockers to lower time-to-main

The main blockers visible in this data are:

- oversized GitHub `merge_group` scope, especially the non-agent `sanity` leg
- queue contention during busy periods
- node and parallelism limits that leave the current suite mix above the desired runtime target
- flaky VM-heavy flows that cause retries and reruns on top of the runtime cost

### Controllable levers

The main levers available from this audit are:

- shrink the blocking GitHub merge-queue slice to the smallest merge-safety subset
- rebalance or add GitHub test nodes if the retained slice still misses the desired runtime target
- keep broad and environment-specific coverage in Jenkins rather than in the required merge-queue path
- fix flaky tests and flaky infrastructure, rather than treating flakiness itself as grounds to drop coverage

## Current state

### GitHub Actions

GitHub Actions Go E2E is defined in:

- [`.github/workflows/pr-e2e-testing.yaml`](../../.github/workflows/pr-e2e-testing.yaml)
- [`.github/workflows/run-e2e-tests.yaml`](../../.github/workflows/run-e2e-tests.yaml)

Current triggers:

- `push` to `main`
- `push` to `release-*`
- `pull_request`
- `merge_group`
- `workflow_dispatch`

Current execution shape:

| Bucket | Label filter | Nodes | OS variants |
|---|---|---:|---|
| Non-agent sanity | `sanity && !agent` | 2 | `cs9-bootc` |
| Agent sanity fast | `sanity && agent && !slow` | 2 | `cs9-bootc`, `cs10-bootc` |
| Agent sanity slow | `sanity && agent && slow` | 2 | `cs9-bootc`, `cs10-bootc` |

Current GitHub scope:

- 43 `sanity && !agent`
- 21 `sanity && agent && !slow`
- 4 `sanity && agent && slow`
- 68 total specs

These counts were revalidated from a fresh Linux dry-run discovery on commit `e364d4e9fd1490b197f9f67f18a37f974e51cb89`.

Notable GitHub categories:

- `cli`
- `configuration`
- `resourcesync`
- `selectors`
- `agent`
- `backup_restore`
- `quadlets`
- `parametrisable_templates`
- `rootless`
- `tpm`

### Jenkins

Relevant Jenkins flightctl profile families:

| Profile | Go E2E | Selector |
|---|---|---|
| `edge-management-flightctl-acm` | Yes | `!integration` |
| `edge-management-flightctl-acm-externaldb` | Yes | `!integration` |
| `edge-management-flightctl-acm-disconnected` | Yes | `!integration` |
| `edge-management-flightctl-acm-d` | Yes | `!integration` |
| `edge-management-flightctl-acm-d-disconnected` | Yes | `!integration` |
| `edge-management-flightctl-acm-upgrade` | Yes | `!integration && sanity` |
| `edge-management-flightctl-standalone-upgrade` | Yes | `!integration && sanity` |

Jenkins owns most broad and environment-specific coverage, including:

- `imagebuild`
- `certificate_rotation`
- `multiorg`
- `rbac`
- `rollout`
- `authprovider`
- `observability`
- `decommission`
- disconnected
- external DB
- upgrade

## Overlap and gaps

Primary comparison basis:

- GitHub Actions: all three current `sanity` buckets combined
- Jenkins: the main ACM profile selector

### Coverage counts

| Set | Specs |
|---|---:|
| GitHub Actions total | 68 |
| Jenkins ACM total | 143 |
| Shared | 68 |
| GitHub Actions only | 0 |
| Jenkins only | 75 |

### Coverage percentages

| Metric | Value |
|---|---:|
| GitHub Actions specs also in Jenkins | 100.0% |
| Jenkins ACM specs also in GitHub Actions | 47.6% |
| GitHub Actions unique coverage | 0.0% |
| Jenkins unique coverage | 52.4% |

GitHub-only coverage:

- none relative to the current Jenkins ACM selector

Jenkins-only coverage is concentrated in:

- agent non-sanity flows
- certificate rotation
- image build
- multiorg and RBAC
- rollout
- auth provider
- observability
- decommission

Main redundancy:

- all 68 GitHub Actions specs also run in Jenkins ACM
- Jenkins upgrade profiles re-run the full `!integration && sanity` subset before and after upgrades

Design rule:

- required merge-queue tests should not be exclusive to GitHub merge queue
- the merge-queue slice should remain a subset of broader Jenkins comprehensive coverage

## Measured bottlenecks

### GitHub runtime

Recent `merge_group` data:

- sampled runs: 33
- completed runs: 31
- average queue wait: 2.81 minutes
- maximum queue wait: 92.88 minutes
- average run wall time: 61.71 minutes
- maximum run wall time: 100.33 minutes

Main GitHub bottleneck:

- `sanity && !agent` on `cs9-bootc`: 57.11 minutes average, 69.98 minutes max

Per-suite timing extracted from 9 completed `merge_group` runs shows the main non-agent contributors are:

| Group | Average sec/run |
|---|---:|
| `backup_restore` | 886.27 |
| `quadlets` | 578.03 |
| `parametrisable_templates` | 483.39 |
| `rootless` | 342.06 |
| `containers` | 165.95 |
| `cli` | 203.86 |
| `configuration` | 146.92 |
| `alertmanager_proxy` | 130.76 |
| `tpm` | 96.07 |
| `resourcesync` | 79.31 |
| `selectors` | 69.61 |

### GitHub flakiness

A separate GitHub flakiness review under [`eldar_audit_logs/github_flakiness/report.md`](../../eldar_audit_logs/github_flakiness/report.md) adds another important signal:

- 32 failed GitHub runs were analyzed
- 31 were classified as flakes
- 1 was classified as a real product failure

The most relevant flaky buckets for merge-queue scope are:

- `infra-flake/vm-disk-io-error`
  hit `backup_restore` and `cli`, showing that some VM-heavy suites hurt merge queue through instability as well as runtime
- `test-flake/agent-spec-version-stuck-after-restore`
  hit backup/restore IDs `84934` and `84938`, reinforcing that backup/restore is a poor fit for the required merge-queue path
- `test-flake/device-not-updated-timeout` and related revert/update flakes
  hit several agent update and revert flows, reinforcing that broader agent coverage should stay out of `merge_group`
- `infra-flake/vm-ssh-not-ready`
  hit `cli`, which means `cli` is still worth keeping only because it is core merge-safety coverage, not because it is especially stable

Conclusion from flakiness data: the GitHub merge queue is being hurt by both long-running suites and flaky VM-heavy flows. This is additional evidence that the current merge-queue lane is operationally expensive, but flaky tests should be fixed rather than removed only because they flake.

Important nuance:

- at least one sampled recent GitHub failure was isolated to the `cs10-bootc` agent fast lane, so `cs10-bootc` should not be removed from merge queue without stronger evidence

### Jenkins runtime

Parsed successful `ocp-flightctl-gotests` logs show:

| Profile | Successful runs parsed | Average E2E minutes/run |
|---|---:|---:|
| `edge-management-flightctl-acm` | 5 | 328.31 |
| `edge-management-flightctl-acm-disconnected` | 4 | 312.16 |
| `edge-management-flightctl-acm-externaldb` | 3 | 267.45 |
| `edge-management-flightctl-acm-d` | 2 | 211.93 |
| `edge-management-flightctl-acm-d-disconnected` | 1 | 212.40 |
| `edge-management-flightctl-acm-upgrade` | 3 | 100.87 |
| `edge-management-flightctl-standalone-upgrade` | 1 | 14.43 |

Dominant suites in broad ACM-family profiles:

- full `Agent E2E Suite`
- full `Helm E2E Suite`
- `Rollout Suite`
- `CLI E2E Suite`
- `Backup and Restore E2E Suite`
- `ImageBuild E2E Suite`
- `Hooks E2E Suite`
- `ParametrisableTemplates E2E Suite`

## Recommended scope boundaries

### GitHub Actions

GitHub Actions should have two roles:

- merge-queue lane: smallest merge-safety slice only
- optional PR or post-merge lane: broader `sanity` coverage when queue latency is less important

GitHub should keep only the smallest merge-safety slice in `merge_group`.

### Jenkins

Jenkins should remain the comprehensive and environment-specialized lane.

Why Jenkins can stay broad while GitHub should be trimmed:

- GitHub `merge_group` is the required merge-queue path, so extra coverage there directly increases time-to-main for every queued merge.
- Jenkins broad lanes are not the required path for every PR, so they can absorb slower and more specialized coverage without putting that cost on every merge.
- The two systems therefore have different jobs: GitHub protects merge safety with the smallest practical slice, while Jenkins provides comprehensive regression coverage across broader environments.

Keep in Jenkins:

- the broad ACM selector
- disconnected and external DB variants
- upgrade profiles
- image build, cert rotation, rollout, multiorg, RBAC, auth provider, observability
- the broader `sanity` coverage removed from GitHub merge queue

## Recommended merge-queue removals

Remove these groups from the GitHub merge queue and keep them in broader Jenkins coverage:

- `backup_restore`
  Adds about 14.8 minutes to each GitHub `merge_group` run by itself, making it the single largest measured non-agent delay and therefore a poor fit for the required merge-queue path.
- `quadlets`
  Adds about 9.6 minutes to each GitHub `merge_group` run, which is too much for a specialized deployment surface in the required merge-queue path.
- `parametrisable_templates`
  Adds about 8.1 minutes to each GitHub `merge_group` run, which is too expensive for a suite outside the retained control-plane core.
- `rootless`
  Adds about 5.7 minutes to each GitHub `merge_group` run for application-environment behavior rather than core merge-safety smoke.
- `containers`
  Adds about 2.8 minutes to each GitHub `merge_group` run and belongs with the broader application-environment coverage.
- `alertmanager_proxy`
  Adds about 2.2 minutes to each GitHub `merge_group` run for a relatively narrow surface, so it is poor merge-queue value.
- `tpm`
  Adds about 1.6 minutes to each GitHub `merge_group` run for specialized hardware-oriented behavior.
- `imagebuild`
  Not a current GitHub non-agent time driver, but successful Jenkins runs still spend about 12.8 to 21.6 minutes per run here, so it fits the broad validation lane.
- `certificate_rotation`
  Successful externaldb Jenkins runs spend about 20.1 minutes per run here, which is too heavy for every merge.
- `rollout`
  Successful Jenkins runs spend about 35.9 to 37.0 minutes per run here, making it one of the biggest suites outside the merge-safety profile.
- `multiorg`
  This is broader platform-behavior coverage that should stay in Jenkins rather than consume merge-queue time.
- `rbac`
  This is broader platform-behavior coverage that should stay in Jenkins rather than consume merge-queue time.
- `authprovider`
  This is broader platform-behavior coverage that should stay in Jenkins rather than consume merge-queue time.
- disconnected-only validation
  These Jenkins-style environment-specific lanes average well over 100 minutes per successful run and require broader environment setup than merge-safety smoke.
- external-DB-only validation
  These Jenkins-style environment-specific lanes average well over 100 minutes per successful run and require broader environment setup than merge-safety smoke.
- upgrade validation
  Upgrade lanes are release-style validation and still average about 100.9 minutes per successful Jenkins run in Jenkins.
- broad agent coverage beyond the smoke subset
  The full Jenkins agent lane runs roughly 42 to 114 minutes depending on profile, so keeping it in merge queue would greatly lengthen the required merge-queue path. The current agent flakes should still be fixed separately.

## Concrete merge-queue proposal

### Goal

Reduce GitHub merge-queue Go E2E scope from 68 specs to 30 specs on `upstream/main`, a reduction of 38 specs or about 56%.

### Proposed merge-queue slice

| Slice | Proposed scope | Count |
|---|---|---:|
| Control-plane smoke | `cli`, `configuration`, `resourcesync`, `selectors` with `sanity && !agent` | 25 |
| Agent smoke | five explicit critical agent specs on both `cs9-bootc` and `cs10-bootc` | 5 |

Proposed control-plane job:

- `GO_E2E_DIRS='test/e2e/cli test/e2e/configuration test/e2e/resourcesync test/e2e/selectors'`
- `GINKGO_LABEL_FILTER='sanity && !agent'`

Proposed agent smoke job:

- `GO_E2E_DIRS='test/e2e/agent'`
- `GINKGO_LABEL_FILTER='83871 || 86238 || 75991 || 77671 || 82481'`

Run the agent smoke job on both `cs9-bootc` and `cs10-bootc`.

These five agent specs preserve:

- enrollment and online state
- agent status reporting
- update path
- rollback path
- OCI prefetch path

## Two implementation options

### Option 1: Balanced reduction to 30 specs

This is the current recommended option:

- keep the following control-plane smoke groups in the non-agent lane:
  - `cli` `(about 3.4 min total)`
    Keeps the most common control-plane create, update, delete, and validation flows in the required merge-queue path.
  - `configuration` `(about 2.4 min total)`
    Keeps basic config application and reconciliation coverage in the required merge-queue path.
  - `resourcesync` `(about 1.3 min total)`
    Keeps repository-to-fleet sync behavior in the required merge-queue path.
  - `selectors` `(about 1.2 min total)`
    Keeps core fleet and device selection logic in the required merge-queue path.
- keep these 5 explicit agent smoke specs on both `cs9-bootc` and `cs10-bootc`:
  - `83871` `(about 0.9 min per spec)`
    Keeps OCI prefetch status coverage.
  - `86238` `(about 1.9 min per spec)`
    Keeps systemd status reporting coverage.
  - `75991` `(about 1.0 min per spec)`
    Keeps basic device status reporting coverage.
  - `77671` `(about 8.1 min per spec)`
    Keeps the update path with an embedded application in the required merge-queue path.
  - `82481` `(about 4.2 min per spec)`
    Keeps rollback-on-bad-image coverage in the required merge-queue path.
- keep broader `sanity`, lifecycle, and environment-heavy coverage in Jenkins

Estimated runtime:

- retained non-agent test time is about `8.3 min`, but the current data also shows about `4.1 min` of environment setup, suite setup, `BeforeEach`, and similar overhead in that lane
- that puts the trimmed non-agent lane at about `12.4 min`
- the retained agent smoke specs sum to about `16.1 min` of raw spec time per OS, but with the current 2-node split the slowest node is about `12.4 min`
- after adding roughly `2 to 4 min` of environment setup, suite startup, and per-spec overhead, the retained agent lane is expected to land around `14.5 to 16.5 min` per OS
- because the agent smoke jobs on `cs9-bootc` and `cs10-bootc` run in parallel, the E2E critical path is expected to be gated by one agent lane rather than by the non-agent lane
- current full `merge_group` wall time averages `61.71 min`, while the current non-agent bottleneck alone averages `57.11 min`, so the remaining workflow overhead outside that bottleneck is about `4.6 min`
- practical estimate for Solution 1:
  about `15 to 17 min` for the E2E critical path and about `19 to 21 min` for the full merge-queue wall time if the rest of the workflow stays similar

Pros:

- preserves a meaningful merge-safety slice across both control-plane and agent behavior
- removes the biggest measured non-agent runtime drivers from the required merge-queue path
- lower risk of missing agent regressions than a more aggressive cut
- can be implemented with the current workflow structure and then tuned further if needed

Cons:

- may still miss a strict 15-minute end-to-end target without additional node rebalancing
  The full retained set still includes agent update, rollback, prefetch, and systemd status checks on both `cs9-bootc` and `cs10-bootc`, so this option keeps more required merge-queue work than the smallest possible slice.
- keeps some agent runtime in `merge_group`
  It still pays for required pre-merge coverage of device status, OCI prefetch, embedded-app update, rollback, and systemd reporting.
- likely improves latency materially, but not to the most aggressive possible floor

### Option 2: Aggressive sub-15-minute merge queue

If the real goal is to keep the whole merge queue at 15 minutes or less, use a smaller blocking slice than the 30-spec proposal:

- keep only the control-plane smoke subset in `merge_group`:
  - `cli` `(about 3.4 min total)`
    Keeps the most common control-plane create, update, delete, and validation flows in the required merge-queue path.
  - `configuration` `(about 2.4 min total)`
    Keeps basic config application and reconciliation coverage in the required merge-queue path.
  - `resourcesync` `(about 1.3 min total)`
    Keeps repository-to-fleet sync behavior in the required merge-queue path.
  - `selectors` `(about 1.2 min total)`
    Keeps core fleet and device selection logic in the required merge-queue path.
- either move agent smoke entirely out of the required merge-queue path or keep only 1 to 2 agent canary specs, for example:
  - `75991` `(about 1.0 min per spec)`
    Keeps one minimal device-status signal in the required merge-queue path.
  - `82481` `(about 4.2 min per spec)`
    Keeps one minimal rollback signal in the required merge-queue path.
- rebalance or add GitHub test nodes so the retained blocking slice stays below the target
- run the removed agent and broader `sanity` coverage in optional GitHub post-merge/PR lanes or Jenkins

Why this is smaller:

- it keeps only the most common control-plane flows that developers hit frequently
- it removes the broader agent smoke slice from the required merge-queue path, or reduces it to a minimal canary
- it is the cleanest option if the top priority is a very small required merge-queue path

Pros:

- best chance of reaching a strict sub-15-minute time-to-main target
- maximizes merge-queue throughput
- makes the required merge-queue path much easier to reason about and keep stable

Cons:

- materially weaker merge-time agent coverage
  If agent smoke is removed or reduced to a canary, pre-merge coverage can miss OCI prefetch, systemd status reporting, embedded-app update, and part of rollback behavior.
- increases dependence on post-merge GitHub or Jenkins signals to catch agent regressions
  Broader pre-merge coverage for rootless applications, backup and restore, quadlets, containers, alertmanager proxy, TPM, and most agent flows is no longer in the required merge-queue path.
- higher risk that a regression affecting device flows lands in `main` before the broader lanes report it

Recommended choice:

- use Option 1 if the goal is to materially reduce merge latency while keeping broader required merge-queue coverage
- use Option 2 only if the priority is the smallest possible required merge-queue path

### Optional GitHub PR validation outside merge queue

To avoid re-expanding `merge_group`, GitHub should also offer one optional targeted PR validation path for suites removed from the merge queue.

Suggested flow:

1. A developer opens or updates a PR.
2. The fixed `merge_group` slice stays unchanged and continues to run only the smaller 30-spec merge-safety set.
3. If the developer wants extra GitHub validation for that PR, they manually trigger a separate GitHub workflow.
4. That workflow accepts runtime inputs such as:
   - `go_e2e_dirs`
   - `label_filter`
   - optional `os_id`
5. The developer enters the removed suite or subset they care about, for example:
   - `test/e2e/backup_restore`
   - `test/e2e/quadlets`
   - `test/e2e/alertmanager_proxy`
6. GitHub Actions runs that extra suite only for the PR branch.
7. The result is visible in GitHub for that PR, but it does not change or enlarge the merge queue for everyone else.

Why not make merge queue dynamic per PR:

- `merge_group` is a shared merge-queue lane, not a per-PR custom lane.
- Letting one PR add `backup_restore` or another removed group would change the required merge-queue path for the queue rather than just validating that PR in isolation.
- That would reintroduce the same merge-queue latency the reduction is trying to remove.
- The right model is a fixed small merge queue plus optional extra GitHub validation for a specific PR.

Important rule:

- do not make `merge_group` dynamic per PR
- keep the extra validation workflow optional and outside the required merge-queue path

## Success metrics

Use these to judge the new split:

- lower average `merge_group` wall time
- lower maximum queue wait during busy periods
- lower rate of flake-driven `merge_group` failures and reruns
- additional stretch goal: work toward less than 15 minutes end-to-end from PR creation to code available in `main`
- no increase in missed regressions from the retained merge-safety slice
- Jenkins continues to own broad, disconnected, external-DB, and upgrade validation
- reduced duplicate execution in GitHub for suites that add large runtime with low merge-safety value

## Implementation outline

1. Choose between Option 1 (balanced 30-spec slice) and Option 2 (aggressive sub-15-minute slice).
2. Update GitHub `merge_group` scope to the chosen slice.
3. Rebalance or add GitHub test nodes if the retained slice still misses the desired runtime target.
4. Keep existing broad Jenkins lanes unchanged initially.
5. Monitor GitHub merge-queue wall time and failure rate for 2 weeks.
6. Continue fixing flaky tests and flaky infrastructure in parallel; do not use flakiness alone as the criterion for removing coverage.
7. Reassess any remaining borderline suites after observing the smaller merge queue in production.
