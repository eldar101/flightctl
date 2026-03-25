# E2E CI scope audit for GitHub Actions and Jenkins

This document inventories the Go E2E coverage currently defined for GitHub Actions and Jenkins, compares the two platforms, and proposes scope boundaries for future placement.

The primary optimization target is time-to-main, with GitHub `merge_group` scope as the biggest controllable lever. Today, even small changes can wait about 1.5 hours in the GitHub merge queue because `merge_group` runs the full current GitHub E2E matrix. That makes GitHub Actions scope an operational bottleneck, not just a coverage question.

The analysis is based on these primary sources:

- `upstream/main` of `flightctl/flightctl` for repository content such as GitHub workflows and `test/e2e`
- Your local [`ocp-edge-ci/jenkins/ci-profiles-new/`](../../ocp-edge-ci/jenkins/ci-profiles-new/) copy for Jenkins profile analysis
- The downloaded GitHub flakiness report under [`eldar_audit_logs/github_flakiness/`](../../eldar_audit_logs/github_flakiness/)

The current repository snapshot used for workflow and discovery validation was commit `e364d4e9fd1490b197f9f67f18a37f974e51cb89`.

## Table of contents

- [Terminology](#terminology)
- [Data reviewed](#data-reviewed)
- [Executive summary](#executive-summary)
- [Method and limitations](#method-and-limitations)
- [GitHub Actions inventory](#github-actions-inventory)
- [Jenkins inventory](#jenkins-inventory)
- [Overlap and gap analysis](#overlap-and-gap-analysis)
- [Runtime comparison](#runtime-comparison)
- [Recommended scope boundaries](#recommended-scope-boundaries)
- [Recommended merge-queue removals](#recommended-merge-queue-removals)
- [Concrete merge-queue proposal](#concrete-merge-queue-proposal)
- [Two implementation options](#two-implementation-options)
- [Decision framework for future placement](#decision-framework-for-future-placement)
- [Implementation outline](#implementation-outline)
- [Open items for stakeholder review](#open-items-for-stakeholder-review)
- [Success metrics](#success-metrics)

## Terminology

- `spec` means one individual Ginkgo test case.
- `bucket` means one CI test grouping defined by a label filter, such as non-agent sanity or agent slow.
- `node` means one parallel Ginkgo worker running part of a bucket.
- `job` means one CI execution unit in GitHub Actions or Jenkins.
- `matrix entry` means one specific GitHub workflow combination, such as an OS variant plus a label-filter bucket.
- `artifact` means saved CI output such as logs or JUnit XML results.
- `wall time` means the total elapsed clock time for a CI run or job.
- `queue wait` means the time a job spends waiting before it actually starts running.
- `slice` means the chosen subset of tests kept together for a specific CI purpose. In this audit, it usually means the tests kept in the required merge-queue path.
- `merge_group` means GitHub's merge queue run. This is the required CI path that must pass before code reaches `main`.
- `required merge-queue path` means the set of checks that block the merge queue. If a test is in this path, it can stop a PR from landing.
- `time-to-main` means the total time from a ready PR until the change is actually available in `main`.
- `sanity` means the current label used for the smaller, higher-priority E2E subset. It does not mean all important tests, only the tests currently tagged that way.
- `control-plane subset` means tests of the central service behavior rather than device-internal behavior. In this proposal that is `cli`, `configuration`, `resourcesync`, and `selectors`.
- `agent smoke` means a very small set of device-agent health tests. In this proposal that is the retained agent specs for device status, OCI prefetch, systemd reporting, update, and rollback.
- `non-agent` means tests that do not exercise device-agent behavior directly. In practice these are mostly service-side and control-plane flows.
- `agent fast` means agent-labeled tests that are not marked `slow`.
- `agent slow` means agent-labeled tests marked `slow`, usually because they involve longer device lifecycle behavior.
- `canary specs` means one or two very small representative tests kept only to catch obvious regressions early, not to provide broad coverage.
- `OS variant` means the device image under test, such as `cs9-bootc` or `cs10-bootc`.
- `broad validation lane` means the larger CI lane intended to catch more scenarios, more environments, and slower regressions. In this audit, that is Jenkins.
- `Jenkins ACM lane` means the main broad Jenkins validation path used in this audit.
- `specialized coverage` means tests for narrower or heavier areas such as backup and restore, TPM, quadlets, disconnected, external DB, or upgrade.
- `post-merge lane` means validation that runs after code reaches `main`, not before.
- `required check` means a CI result that must pass before the merge queue can continue.
- `dry-run discovery` means listing which tests would run without actually executing them.
- `optional PR validation` means a manually triggered GitHub workflow for a specific PR that runs extra tests without expanding the required merge queue.
- `selector` means the test filter used to decide which tests run, such as `sanity && !agent` or `!integration`.

## Data reviewed

This audit compares current GitHub workflow definitions, current Jenkins test selection, fresh Linux dry-run test discovery, recent GitHub merge-queue timing data, GitHub Actions job logs and JUnit artifacts, Jenkins console logs from saved runs, and a separate GitHub flaky-test review.

## Executive summary

- GitHub Actions on `upstream/main` currently runs 68 Go E2E specs, all selected through `sanity`-based label filters.
- The main Jenkins ACM profile currently runs 143 Go E2E specs through the broader selector `!integration`.
- 68 specs run in both systems.
- 0 specs run only in GitHub Actions.
- 75 specs run only in Jenkins.
- GitHub Actions is not actually a small fast lane today. It is a 68-spec merge-queue gate with a measured average `merge_group` wall time of 61.71 minutes.
- The dominant GitHub bottleneck is the non-agent `sanity` leg at 57.11 minutes average.
- Recent `merge_group` log extraction shows that the non-agent bottleneck is primarily driven by `backup_restore`, `quadlets`, `parametrisable_templates`, `rootless`, `containers`, `alertmanager_proxy`, and `tpm`.
- A separate GitHub flakiness review classified 31 of 32 failed GitHub runs as flakes, which means merge-queue pain is coming from both runtime and instability.
- Jenkins is effectively the broad validation lane for non-`integration` Go E2E coverage, plus upgrade, disconnected, external-DB, software-catalog, UI, and quadlet/RPM-specific profile families.
- Parsed successful Jenkins `ocp-flightctl-gotests` logs show ACM-family lanes averaging roughly 212 to 328 minutes of E2E runtime per run, with the broad ACM profiles dominated by full agent, full helm, rollout, CLI, backup/restore, imagebuild, hooks, and parametrisable templates.
- A concrete smaller merge-queue slice can cut GitHub merge-queue Go E2E coverage from 68 specs to 30 specs while keeping core CLI, configuration, selectors, resource sync, and agent smoke coverage.
- If a stricter target of 15 minutes or less to `main` is required, an even smaller blocking slice is possible, but with weaker agent coverage in the required merge-queue path.

## Blockers to lower time-to-main

The main blockers visible in this data are:

- oversized GitHub `merge_group` scope, especially the non-agent `sanity` leg
- queue contention during busy periods
- node and parallelism limits that leave the current suite mix above the desired runtime target
- flaky VM-heavy flows that cause retries and reruns on top of the runtime cost

## Controllable levers

The main levers available from this audit are:

- shrink the blocking GitHub merge-queue slice to the smallest merge-safety subset
- rebalance or add GitHub test nodes if the retained slice still misses the desired runtime target
- keep broad and environment-specific coverage in Jenkins rather than in the required merge-queue path
- fix flaky tests and flaky infrastructure, rather than treating flakiness itself as grounds to drop coverage

## Method and limitations

The comparison uses:

- GitHub Actions workflow configuration from `upstream/main`
- E2E Ginkgo labels from `upstream/main`
- Jenkins profile configuration from your local `ocp-edge-ci/jenkins/ci-profiles-new/` directory
- GitHub Actions run history, job timing, and raw `merge_group` logs exported with `gh`
- Local Jenkins `ocp-flightctl-gotests` console logs saved under [`eldar_audit_logs/`](../../eldar_audit_logs/)
- The downloaded GitHub flakiness report under [`eldar_audit_logs/github_flakiness/`](../../eldar_audit_logs/github_flakiness/)

The analysis now includes measured runtime evidence, but still has some limits:

- Trigger cadence is available where `trigger_day` or workflow triggers are present.
- Exact Jenkins wall-clock schedules are not available in these checked-in profile files.
- Jenkins successful-run suite timing is based on the saved local log sample, not the full historical corpus.
- Some GitHub `merge_group` runs or Jenkins runs were missing parseable logs, so suite-level timing samples are smaller than the full job-history samples.

Even with those limits, the merge-queue reduction proposal below is now grounded in measured GitHub and Jenkins runtime data rather than selector breadth alone.

## GitHub Actions inventory

### Workflows and triggers

GitHub Actions Go E2E execution is defined in:

- [`.github/workflows/pr-e2e-testing.yaml`](../../.github/workflows/pr-e2e-testing.yaml)
- [`.github/workflows/run-e2e-tests.yaml`](../../.github/workflows/run-e2e-tests.yaml)

Current triggers:

- `push` to `main`
- `push` to `release-*`
- `pull_request` for `opened`, `synchronize`, `reopened`, `labeled`
- `merge_group`
- `workflow_dispatch`

PR gating behavior:

- PRs targeting `main` skip E2E unless the PR has the `run-e2e` label.
- PRs targeting `release-*` do run E2E.
- Merge queue runs E2E through `merge_group`.

That means the current GitHub E2E matrix is part of the protected-branch critical path.

### Execution environment

- Runner OS: `ubuntu-24.04`
- Backend type: `helm`
- Namespace: `flightctl-external`
- Agent OS variants:
- `cs9-bootc`
- `cs10-bootc`

### Test selection

GitHub Actions runs three label-filter buckets:

| Bucket | Label filter | Nodes | OS variants |
|---|---|---|---|
| Non-agent sanity | `sanity && !agent` | 2 | `cs9-bootc` |
| Agent sanity fast | `sanity && agent && !slow` | 2 | `cs9-bootc`, `cs10-bootc` |
| Agent sanity slow | `sanity && agent && slow` | 2 | `cs9-bootc`, `cs10-bootc` |

### Current scope by count

GitHub Actions currently covers 68 specs total:

- 43 `sanity && !agent`
- 21 `sanity && agent && !slow`
- 4 `sanity && agent && slow`

These counts were revalidated from a fresh Linux dry-run discovery on commit `e364d4e9fd1490b197f9f67f18a37f974e51cb89`.

High-volume areas in GitHub Actions:

- `cli`: 18
- `agent`: 19
- `quadlets`: 6
- `parametrisable_templates`: 4
- `selectors`: 4
- `configuration`: 3

Notable categories present in GitHub Actions:

- `sanity`
- `agent`
- `slow`
- `quadlets`
- `backup-restore`
- `tpm`
- `rootless`

Notable categories absent from GitHub Actions:

- `imagebuild`
- `certificate-rotation`
- `multiorg`
- `rbac`
- `rollout`
- `authprovider`
- most unlabeled or specialty agent specs

### Merge-queue impact

On `upstream/main`, the merge queue currently inherits the same GitHub Actions E2E scope:

- 68 logical specs
- 5 job matrix entries
- duplicated execution across `cs9-bootc` and `cs10-bootc` for agent buckets

That is the wrong shape for a queue-optimization lane. The merge queue should validate only the smallest set that protects merge safety.

### Observed GitHub Actions metrics

Using recent `merge_group` history exported from GitHub Actions:

- sampled merge-group runs: 33
- completed runs in sample: 31
- successful runs: 21
- failed runs: 9
- cancelled runs: 1
- in-progress at export time: 2
- average queue wait: 2.81 minutes
- maximum queue wait: 92.88 minutes
- average run wall time: 61.71 minutes
- maximum run wall time: 100.33 minutes

The queue itself is not the main average cost. The main average cost is the runtime of the merge-group workflow after it starts. The worst-case wait time, however, confirms that queue contention can still become severe.

The dominant GitHub merge-group bottleneck is the non-agent sanity leg:

- `sanity && !agent` on `cs9-bootc`: 57.11 minutes average, 69.98 minutes max

By comparison:

- `sanity && agent && !slow` legs are about 30 to 31 minutes each
- `sanity && agent && slow` legs are about 24 to 28 minutes each
- `build-backend-containers` averages 13.21 minutes
- agent image build legs average about 8 to 9 minutes each

This is important because it means the current merge-queue delay is driven more by the breadth of the non-agent `sanity` slice than by the agent OS matrix.

Recent non-agent suite timing extraction from 9 completed `merge_group` runs confirms that the non-agent bottleneck is dominated by a small number of suites:

- `backup_restore`: 886.27 seconds per run on average
- `quadlets`: 578.03 seconds per run on average
- `parametrisable_templates`: 483.39 seconds per run on average
- `applications` overall: 473.81 seconds per run on average
- `rootless`: 342.06 seconds per run on average
- `containers`: 165.95 seconds per run on average
- `cli`: 203.86 seconds per run on average
- `configuration`: 146.92 seconds per run on average
- `alertmanager_proxy`: 130.76 seconds per run on average
- `tpm`: 96.07 seconds per run on average
- `resourcesync`: 79.31 seconds per run on average
- `selectors`: 69.61 seconds per run on average

The same extraction also shows that several suites currently present in the non-agent discovery universe are not consuming merge-queue runtime in this lane:

- `imagebuild`
- `certificate_rotation`
- `rollout`
- `rbac`
- `authprovider`
- `observability`

Those suites appeared with zero selected specs in the sampled non-agent logs and therefore are not current time drivers for the GitHub merge queue.

### GitHub flakiness findings

A separate GitHub flakiness review under [`eldar_audit_logs/github_flakiness/report.md`](../../eldar_audit_logs/github_flakiness/report.md) adds another important signal for merge-queue design:

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

This does not change the main recommendation. It strengthens it. The GitHub merge queue is being hurt by both long-running suites and flake-prone VM-heavy flows, but flaky tests should still be fixed rather than removed only because they flake.

Note on data quality:

- a small number of GitHub job records had invalid timestamps such as `0001-01-01T00:00:00Z` or 1-second inversions
- those rows were excluded from per-job duration summaries
- they do not materially change the bottleneck conclusion

## Jenkins inventory

### Current profile families found

The latest repo-local Jenkins flightctl profile set is under `4.20`. Older `4.17` to `4.19` profiles follow the same general pattern with minor selector differences.

Relevant `4.20` profile families:

| Profile | Go E2E | Selector |
|---|---|---|
| `edge-management-flightctl-acm.yaml` | Yes | `!integration` |
| `edge-management-flightctl-acm-externaldb.yaml` | Yes | `!integration` |
| `edge-management-flightctl-acm-disconnected.yaml` | Yes | `!integration` |
| `edge-management-flightctl-acm-d.yaml` | Yes | `!integration` |
| `edge-management-flightctl-acm-d-disconnected.yaml` | Yes | `!integration` |
| `edge-management-flightctl-acm-upgrade.yaml` | Yes | `!integration && sanity` |
| `edge-management-flightctl-standalone-upgrade.yaml` | Yes | `!integration && sanity` |
| `edge-management-flightctl-acm-ui.yaml` | No | UI only |
| `edge-management-flightctl-standalone-ui.yaml` | No | UI only |
| `edge-management-flightctl-quadlets-externaldb.yaml` | No Go E2E | RPM install flow only |

### Execution environment

Current Jenkins flightctl Go E2E profiles use OCP-backed environments rather than GitHub-hosted VMs:

- ACM connected cluster
- ACM connected cluster with external DB
- ACM disconnected cluster
- ACM software catalog variants
- Upgrade validation against previous release, current release, and `main`

Common environment traits:

- OCP deployment via Jenkins profile jobs
- flightctl deployed by Ansible or install job
- optional ACM/MCE install
- optional disconnected install
- optional external DB
- optional software catalog enablement
- optional follow-on VM creation and Cypress UI checks in some ACM-D variants

### Triggers and frequency

The `4.20` profiles are day-based scheduled jobs:

| Profile | Trigger days |
|---|---|
| `edge-management-flightctl-acm.yaml` | Sunday to Saturday |
| `edge-management-flightctl-acm-externaldb.yaml` | Sunday, Tuesday, Thursday |
| `edge-management-flightctl-acm-disconnected.yaml` | Sunday to Saturday |
| `edge-management-flightctl-acm-d.yaml` | Tuesday, Thursday |
| `edge-management-flightctl-acm-d-disconnected.yaml` | Monday, Friday |
| `edge-management-flightctl-acm-upgrade.yaml` | Monday, Wednesday, Friday |
| `edge-management-flightctl-standalone-upgrade.yaml` | Monday, Wednesday, Friday |

### Current scope by count

Using the main ACM selector from [`edge-management-flightctl-acm.yaml`](../../ocp-edge-ci/jenkins/ci-profiles-new/4.20/edge-management-flightctl-acm.yaml), Jenkins covers 143 specs.

Highest-volume areas in the main ACM lane:

- `cli`: 18
- `agent_pruning`: 9
- `agent_update`: 8
- `agent`: 7
- `applications/helm`: 7
- `certificate_rotation`: 7
- `cli_console`: 7
- `multiorg`: 6
- `quadlets`: 6

Selector deltas across Jenkins profiles:

- `acm`, `acm-externaldb`, `acm-disconnected`, `acm-d`, and `acm-d-disconnected` are now equivalent for Go E2E scope because they all use `!integration`.
- `acm-upgrade` and `standalone-upgrade` now use the plain `!integration && sanity` selector with no extra alertmanager exclusions.

### Categories owned by Jenkins today

Jenkins currently owns most of the following categories:

- `imagebuild`
- `certificate-rotation`
- `multiorg`
- `rbac`
- `rollout`
- `authprovider`
- `observability`
- `decommission`
- unlabeled agent and CLI scenarios
- non-`sanity` console and lifecycle flows
- environment-specific variants such as disconnected, external DB, software catalog, and upgrade

### Observed Jenkins metrics

The Jenkins samples below come from the recent build-history JSON you provided for the main flightctl profile families. They are sufficient to characterize whether these lanes are appropriate for merge blocking.

#### `edge-management-flightctl-acm`

- sample size: 17 builds
- visible results:
- `SUCCESS`: 1
- `FAILURE`: 3
- `UNSTABLE`: 9
- `ABORTED`: 4
- `in progress`: 1
- observed duration range:
- short failures or aborts: about 5 to 73 minutes
- longer unstable runs: about 349 to 640 minutes
- visible success: about 198 minutes

#### `edge-management-flightctl-acm-externaldb`

- sample size: 5 builds
- visible results:
- `SUCCESS`: 1
- `FAILURE`: 2
- `UNSTABLE`: 1
- `in progress`: 1
- observed duration range:
- success: about 138 minutes
- unstable or failure runs: about 329 to 641 minutes

#### `edge-management-flightctl-acm-disconnected`

- sample size: 9 builds
- visible results:
- `SUCCESS`: 1
- `FAILURE`: 3
- `UNSTABLE`: 3
- `ABORTED`: 2
- observed duration range:
- short failure or success cases: about 37 to 75 minutes
- longer unstable, failure, or aborted cases: about 544 to 661 minutes

#### `edge-management-flightctl-acm-d`

- sample size: 5 builds
- visible results:
- `SUCCESS`: 1
- `FAILURE`: 2
- `UNSTABLE`: 2
- observed duration range:
- success: about 149 minutes
- failures or unstable runs: about 47 to 368 minutes

#### `edge-management-flightctl-acm-d-disconnected`

- sample size: 3 builds
- visible results:
- `UNSTABLE`: 2
- `ABORTED`: 1
- observed duration range:
- about 373 to 661 minutes

#### `edge-management-flightctl-acm-upgrade`

- sample size: 14 builds
- visible results:
- `FAILURE`: 6
- `UNSTABLE`: 5
- `ABORTED`: 3
- `SUCCESS`: 0 in the sample
- observed duration range:
- about 1 to 517 minutes

#### `edge-management-flightctl-standalone-upgrade`

- sample size: 6 builds
- visible results:
- `ABORTED`: 5
- `UNSTABLE`: 1
- `SUCCESS`: 0 in the sample
- observed duration range:
- about 18 to 661 minutes

### Observed Jenkins suite timing

In addition to build-history metadata, local `ocp-flightctl-gotests` console logs were parsed for the saved successful runs under [`eldar_audit_logs/`](../../eldar_audit_logs/). Those logs provide the same `Running Suite:` and `Ran X of Y Specs in ...` evidence used on the GitHub side.

Measured average E2E runtime per successful run from the saved Jenkins logs:

| Profile | Successful runs parsed | Average E2E minutes per run |
|---|---|---|
| `edge-management-flightctl-acm` | 5 | 328.31 |
| `edge-management-flightctl-acm-disconnected` | 4 | 312.16 |
| `edge-management-flightctl-acm-externaldb` | 3 | 267.45 |
| `edge-management-flightctl-acm-d` | 2 | 211.93 |
| `edge-management-flightctl-acm-d-disconnected` | 1 | 212.40 |
| `edge-management-flightctl-acm-upgrade` | 3 | 100.87 |
| `edge-management-flightctl-standalone-upgrade` | 1 | 14.43 |

The suite-level runtime profile is also clear.

For `edge-management-flightctl-acm`, the dominant suites in the parsed successful runs are:

- `Agent E2E Suite`: 102.18 minutes per run
- `Helm E2E Suite`: 74.14 minutes per run
- `Rollout Suite`: 36.95 minutes per run
- `CLI E2E Suite`: 36.84 minutes per run
- `Backup and Restore E2E Suite`: 18.95 minutes per run
- `ImageBuild E2E Suite`: 13.47 minutes per run
- `Hooks E2E Suite`: 13.16 minutes per run
- `ParametrisableTemplates E2E Suite`: 11.48 minutes per run

For `edge-management-flightctl-acm-disconnected`, the dominant suites in the parsed successful runs are:

- `Agent E2E Suite`: 114.02 minutes per run
- `Helm E2E Suite`: 81.86 minutes per run
- `CLI E2E Suite`: 36.61 minutes per run
- `ImageBuild E2E Suite`: 21.62 minutes per run
- `Backup and Restore E2E Suite`: 16.81 minutes per run
- `microshift ACM enrollment E2E Suite`: 14.67 minutes per run
- `Hooks E2E Suite`: 13.08 minutes per run
- `Quadlets E2E Suite`: 10.51 minutes per run

For `edge-management-flightctl-acm-externaldb`, the dominant suites in the parsed successful runs are:

- `Agent E2E Suite`: 97.81 minutes per run
- `Helm E2E Suite`: 75.56 minutes per run
- `Rollout Suite`: 35.87 minutes per run
- `CLI E2E Suite`: 33.31 minutes per run
- `Certificate Rotation E2E Suite`: 20.08 minutes per run
- `ImageBuild E2E Suite`: 12.79 minutes per run
- `Hooks E2E Suite`: 12.20 minutes per run
- `ParametrisableTemplates E2E Suite`: 10.92 minutes per run

For `edge-management-flightctl-acm-d` and `edge-management-flightctl-acm-d-disconnected`, the saved successful runs show a different broad shape:

- `Agent E2E Suite`: about 77 to 80 minutes per run
- `CLI E2E Suite`: about 63 minutes per run
- `ParametrisableTemplates E2E Suite`: about 14 to 15 minutes per run
- `Hooks E2E Suite`: about 14 to 15 minutes per run
- `microshift ACM enrollment E2E Suite`: about 14 to 16 minutes per run

For `edge-management-flightctl-acm-upgrade`, the parsed successful runs are much narrower:

- `Agent E2E Suite`: 41.89 minutes per run
- `CLI E2E Suite`: 21.89 minutes per run
- `Hooks E2E Suite`: 8.70 minutes per run
- `Resourcesync E2E Suite`: 7.95 minutes per run
- `ParametrisableTemplates E2E Suite`: 7.28 minutes per run

The standalone-upgrade sample is too small to generalize from, because only one successful `ocp-flightctl-gotests` log was available in the saved set.

### Jenkins interpretation

The important takeaway is not only that Jenkins lanes are broader than GitHub Actions. It is that they are also operationally noisy and frequently long-running.

Across the sampled Jenkins profile families:

- successful runs are sparse in the samples
- `UNSTABLE`, `FAILURE`, and `ABORTED` states dominate
- many environment-rich lanes run for several hours
- disconnected and upgrade profiles are especially expensive and noisy

The parsed successful console logs make the runtime tradeoff concrete:

- Jenkins ACM-family lanes are dominated by broad suites such as full agent, full helm, rollout, CLI, backup/restore, imagebuild, hooks, and parametrisable templates
- disconnected and external-DB variants preserve most of that broad runtime shape while adding environment-specific cost
- the narrower upgrade lane is still materially larger than the proposed GitHub merge-safety slice

That makes these profiles appropriate for comprehensive validation and release confidence, but inappropriate as candidates for GitHub merge-queue blocking.

## Overlap and gap analysis

### Comparison basis

The primary comparison below uses:

- GitHub Actions: all three current `sanity` buckets combined
- Jenkins: the main ACM profile selector

### Coverage counts

| Set | Specs |
|---|---|
| GitHub Actions total | 68 |
| Jenkins ACM total | 143 |
| Shared | 68 |
| GitHub Actions only | 0 |
| Jenkins only | 75 |

### Coverage percentages

| Metric | Value |
|---|---|
| GitHub Actions specs also in Jenkins | 100.0% |
| Jenkins ACM specs also in GitHub Actions | 47.6% |
| GitHub Actions unique coverage | 0.0% |
| Jenkins unique coverage | 52.4% |

### Tests running only in GitHub Actions

There are no longer any GitHub-only Go E2E specs relative to the current Jenkins ACM selector.

### Tests running only in Jenkins

Jenkins-only coverage is concentrated in these areas:

- Agent non-sanity flows
- Console extended scenarios
- Certificate rotation
- Image build
- Multiorg and RBAC
- Rollout
- Auth provider
- Observability
- Decommission

Representative Jenkins-only specs:

- `should verify basic image build process - build, export, download, use in an agent`
- `should rotate certificate while device stays online`
- `all users should belong to the same flightctl organization`
- `should select devices correctly based on BatchSequence strategy`
- `logs in via OAuth --web --no-browser with Keycloak and headless browser`

### Redundancy

The main redundancy is deliberate but not documented:

- All 68 GitHub Actions specs also run in Jenkins ACM.
- Jenkins upgrade profiles re-run the full `!integration && sanity` subset before and after upgrades.
- Jenkins software-catalog profiles can run broader selectors than standard ACM, increasing duplication further.

That also supports an important design rule for future placement:

- required merge-queue tests should not be exclusive to GitHub merge queue
- the merge-queue slice should remain a subset of broader Jenkins comprehensive coverage

That redundancy is useful only if the goal is environment validation across:

- GitHub-hosted ephemeral environments
- OCP ACM environments
- upgrade transitions
- disconnected or external DB deployments

Without that framing, the current split looks wasteful.

### Runtime comparison

With measured GitHub and Jenkins data together, the platform split becomes clearer:

- GitHub merge-group runs average about 61.71 minutes wall time
- the dominant GitHub bottleneck is the broad non-agent `sanity` leg at about 57.11 minutes average
- Jenkins environment-rich lanes often run for several hours
- Jenkins upgrade and disconnected profiles are materially noisier than GitHub merge-group lanes

This leads to a practical conclusion: GitHub is too broad today for merge safety, Jenkins is too expensive and too unstable to move into the merge queue, and the right action is to shrink GitHub merge-group scope.

## Recommended scope boundaries

### GitHub Actions scope

GitHub Actions should have two roles:

- Merge queue lane: smallest merge-safety slice only
- Optional PR or post-merge lane: broader `sanity` coverage when queue latency is not the primary concern

Recommended criteria:

- merge-queue validation must optimize for throughput first
- only coverage that materially protects merge safety
- `sanity` coverage only
- cross-agent OS coverage where it matters to developers
- low- to medium-duration suites
- issues likely to regress due to code changes in the same repository

GitHub should keep only the smallest merge-safety slice in `merge_group`.

### Jenkins scope

Jenkins should be the comprehensive and environment-specialized lane.

Why Jenkins can stay broad while GitHub should be trimmed:

- GitHub `merge_group` is the required merge-queue path, so every additional minute there directly increases time-to-main for queued changes.
- Jenkins broad lanes are not the required path for every PR, so they can carry slower, broader, and more environment-specific validation without making every merge wait for it.
- That is why the audit does not recommend shrinking Jenkins to the same minimal slice. The two systems serve different purposes: GitHub protects merge safety with the smallest practical subset, while Jenkins remains the broad regression and environment-validation lane.

Recommended criteria:

- environment-specific validation that GitHub Actions cannot represent well
- broad non-`integration` regression coverage
- upgrade and release-gate coverage
- disconnected and external DB variants
- specialty suites with higher runtime or infrastructure cost

Recommended Jenkins content:

- Keep the broad ACM selector as the comprehensive lane
- Keep disconnected and external DB variants in Jenkins only
- Keep upgrade profiles in Jenkins only
- Keep image build, cert rotation, rollout, multiorg, RBAC, auth provider, and observability in Jenkins only
- Absorb the `sanity` tests removed from the GitHub merge queue
- Keep noisy or hours-long lanes in Jenkins even when they are strategically important, because their operational profile is not suitable for merge-group blocking

### When tests should run in both

A test should run in both systems only if at least one of these is true:

- It is a `sanity` gate for developer feedback.
- It validates behavior that is likely to differ between GitHub Actions and Jenkins environments.
- It is needed before and after upgrade checkpoints.
- It protects a high-frequency regression area where early failure is materially valuable.

That means most dual-platform coverage should remain limited to a small merge-safety subset of `sanity`, not the entire current `sanity` inventory.

This audit does not include exact CPU or memory counters per suite. The recommendation uses measured runtime in GitHub and Jenkins plus the breadth of environment setup required by each lane. In practice, anything that adds many minutes to every merge or requires a broader environment than the retained smoke slice is a poor candidate for the protected-branch critical path.

### Recommended merge-queue removals

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

Run these two logical slices in `merge_group`:

| Slice | Proposed scope | Count |
|---|---|---|
| Control-plane smoke | `cli`, `configuration`, `resourcesync`, `selectors` with `sanity && !agent` | 25 |
| Agent smoke | five explicit critical agent specs on both `cs9-bootc` and `cs10-bootc` | 5 |

Proposed control-plane scope:

- `test/e2e/cli`
- `test/e2e/configuration`
- `test/e2e/resourcesync`
- `test/e2e/selectors`

with:

- `GINKGO_LABEL_FILTER='sanity && !agent'`

Proposed agent smoke IDs:

- `83871` agent prefetch status
- `86238` systemd status reporting
- `75991` device status
- `77671` update with embedded application
- `82481` rollback on broken image

One concrete way to implement this with existing knobs is:

- control-plane job:

  `GO_E2E_DIRS='test/e2e/cli test/e2e/configuration test/e2e/resourcesync test/e2e/selectors'`

  with `GINKGO_LABEL_FILTER='sanity && !agent'`

- agent smoke job:

  `GO_E2E_DIRS='test/e2e/agent'`

  with `GINKGO_LABEL_FILTER='83871 || 86238 || 75991 || 77671 || 82481'`

Run the agent smoke job on both `cs9-bootc` and `cs10-bootc`.

These five specs are all already in the current GitHub `sanity` inventory and represent the smallest agent slice that still covers:

- enrollment and online state
- agent status reporting
- update path
- rollback path
- OCI prefetch path
- application/systemd state visibility

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

### Placement of removed tests

The 38 removed merge-queue specs should go to one of these places:

- Jenkins ACM comprehensive lane
- Jenkins disconnected or external-DB lanes where appropriate
- Jenkins upgrade lanes for upgrade-sensitive `sanity`
- optional post-merge GitHub Actions `sanity-full` lane if a GitHub-only signal is still desired

## Decision framework for future placement

Use this decision order when adding or relabeling E2E coverage:

- Is the test intended to block `merge_group`?

   Put it in the smallest possible GitHub merge-queue slice only if yes.

- Is the test intended for PR feedback within one developer cycle, but not necessarily merge blocking?

   Put it in GitHub Actions outside the merge queue if yes.

- Does the test require disconnected, ACM, external DB, upgrade, or similar environment specialization?

   Put it in Jenkins if yes.

- Is the test `sanity` and broadly representative of user-critical flows?

   Consider it for both, but default to the smaller merge-queue subset unless there is a strong reason.

- Is the test long-running, infrastructure-heavy, or operationally specialized?

   Prefer Jenkins.

- Is the test mostly product-surface smoke coverage with modest setup cost?

   Prefer GitHub Actions if it helps feedback and queue timing is acceptable.

## Implementation outline

### Phase 1: Choose the merge-queue target

Effort: 0.5 to 1 day

- Decide whether the working target is the balanced 30-spec option or the more aggressive sub-15-minute option.
- If the sub-15-minute option is chosen, decide whether any agent canary specs remain in `merge_group`.

### Phase 2: Document and align selectors

Effort: 1 to 2 days

- Write down the canonical selectors for GitHub Actions and Jenkins.
- Decide whether ACM and ACM-D families should intentionally remain on the same `!integration` selector or diverge again.
- Decide whether a broader post-merge GitHub `sanity-full` lane is still needed after merge-queue reduction.

### Phase 3: Normalize Jenkins profile intent

Effort: 1 to 2 days

- Align Jenkins profile families to explicit scope definitions.
- Remove accidental selector differences.
- Add profile comments documenting why each selector exists.
- Split GitHub Actions merge-queue scope from any broader GitHub `sanity` scope.

### Phase 4: Add runtime observability

Effort: 2 to 3 days

- Collect average duration, queue time, and failure-rate data from CI history.
- Add a lightweight generated coverage report if desired.
- Revisit whether any `sanity` subsets should be trimmed or expanded.
- Measure merge-queue wait time before and after reducing the GitHub merge-queue slice.

## Open items for stakeholder review

- Should ACM and ACM-D intentionally stay on the same `!integration` selector, or diverge again?
- Should `rpm-sanity` become an explicit GitHub Actions bucket?
- Are upgrade `sanity` checkpoints sufficient, or should one small non-sanity post-upgrade slice exist in Jenkins?
- Should GitHub keep any broader post-merge `sanity` lane after the merge-queue slice is reduced?

## Success metrics

- GitHub merge-queue E2E scope is clearly bounded and materially smaller than today.
- Merge-queue wait time decreases after scope reduction.
- Additional stretch goal: work toward less than 15 minutes end-to-end from PR creation to code available in `main`.
- Jenkins owns environment-specific and comprehensive validation.
- No accidental selector drift exists between Jenkins profile families.
- Duplicate cross-platform execution is intentional and documented.
- Runtime and cost data become available for future redistribution decisions.
