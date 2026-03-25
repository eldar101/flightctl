# E2E CI scope audit demo

This audit is for a 5 to 10 minute walkthrough. It focuses on the decision, the evidence behind it, and the two implementation options.

## Table of contents

- [Terminology](#terminology)
- [Data reviewed](#data-reviewed)
- [What the audit found](#what-the-audit-found)
- [Why this matters](#why-this-matters)
- [Recommended direction](#recommended-direction)
- [Two implementation options](#two-implementation-options)
- [Suggested talk track](#suggested-talk-track)
- [FAQ](#faq)

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

## What the audit found

- GitHub `merge_group` currently runs 68 Go E2E specs.
- The main Jenkins ACM lane runs 143 Go E2E specs.
- All 68 GitHub specs also run in Jenkins ACM.
- GitHub therefore has no unique required coverage relative to the main Jenkins ACM lane.
- Current GitHub scope was revalidated from Linux dry-run discovery on commit `e364d4e9fd1490b197f9f67f18a37f974e51cb89`.

Current GitHub split:

| Bucket | Filter | Specs | Nodes |
|---|---|---:|---:|
| Non-agent sanity | `sanity && !agent` | 43 | 2 |
| Agent sanity fast | `sanity && agent && !slow` | 21 | 2 |
| Agent sanity slow | `sanity && agent && slow` | 4 | 2 |
| Total | all three buckets | 68 | - |

## Why this matters

- The optimization target is time-to-main.
- Recent GitHub `merge_group` runs average 61.71 minutes wall time.
- The biggest GitHub bottleneck is the non-agent `sanity` leg at 57.11 minutes average.
- The heaviest contributors inside that leg are `backup_restore`, `quadlets`, `parametrisable_templates`, `rootless`, `containers`, `alertmanager_proxy`, and `tpm`.
- A separate GitHub flakiness review classified 31 of 32 failed runs as flakes, so merge-queue pain is being driven by both runtime and instability.
- Jenkins remains the broad validation lane and the local sample continues to show it is noisy and expensive, which makes it a poor candidate for the required merge-queue path.

Bottom line:

- GitHub `merge_group` is too broad today.
- Jenkins should stay broad.
- Required merge-queue coverage should remain a subset of broader Jenkins coverage.

## Recommended direction

- Keep GitHub `merge_group` as the smallest merge-safety slice.
- Keep broader `sanity`, environment-heavy, and specialized coverage in Jenkins.
- Fix flaky tests and flaky infrastructure, but do not use flakiness alone as the reason to remove coverage.
- If developers need extra GitHub validation for a PR, provide an optional manual GitHub workflow outside `merge_group`.

Why GitHub should be trimmed but Jenkins can stay broad:

- GitHub `merge_group` is the required merge-queue path, so every added minute there directly delays code reaching `main`.
- Jenkins broad lanes are not the required path for every PR, so they can carry slower, broader, and more environment-specific validation without slowing every merge.
- The two systems therefore have different jobs: GitHub protects merge safety with the smallest practical slice, while Jenkins remains the broad regression and environment-validation lane.

## Two implementation options

### Option 1: balanced reduction to 30 specs

This is the recommended default.

Keep in `merge_group`:

- control-plane smoke from `sanity && !agent`:
  - `cli` `(about 3.4 min total)`
    Keeps the most common control-plane create, update, delete, and validation flows in the required merge-queue path.
  - `configuration` `(about 2.4 min total)`
    Keeps basic config application and reconciliation coverage in the required merge-queue path.
  - `resourcesync` `(about 1.3 min total)`
    Keeps repository-to-fleet sync behavior in the required merge-queue path.
  - `selectors` `(about 1.2 min total)`
    Keeps core fleet and device selection logic in the required merge-queue path.
- agent smoke on both `cs9-bootc` and `cs10-bootc`:
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

Result:

- reduce GitHub merge-queue scope from 68 specs to 30 specs
- preserve meaningful merge-time coverage for both control-plane and agent behavior
- estimated E2E critical path after trimming: about `15 to 17 min`
- estimated full merge-queue wall time after trimming: about `19 to 21 min`

Estimated runtime breakdown:

- retained non-agent test time is about `8.3 min`, but the current data also shows about `4.1 min` of environment setup, suite setup, `BeforeEach`, and similar overhead in that lane
- that puts the trimmed non-agent lane at about `12.4 min`
- the retained agent smoke specs sum to about `16.1 min` of raw spec time per OS, but with the current 2-node split the slowest node is about `12.4 min`
- after adding roughly `2 to 4 min` of environment setup, suite startup, and per-spec overhead, the retained agent lane is expected to land around `14.5 to 16.5 min` per OS
- because the agent smoke jobs on `cs9-bootc` and `cs10-bootc` run in parallel, the E2E critical path is expected to be gated by one agent lane rather than by the non-agent lane
- current full `merge_group` wall time averages `61.71 min`, while the current non-agent bottleneck alone averages `57.11 min`, so the remaining workflow overhead outside that bottleneck is about `4.6 min`
- taken together, that implies Solution 1 is not a 30-minute merge queue: it is roughly `4.6 min` of non-E2E workflow overhead plus a `14.5 to 16.5 min` E2E critical path, or about `19 to 21 min` total

Pros:

- better balance between latency and merge-time signal
- removes the biggest measured non-agent runtime drivers
- lower regression risk than a more aggressive cut

Cons:

- may still miss a sub-15-minute target without extra node rebalancing
  The full retained set still includes agent update, rollback, prefetch, and systemd status checks on both `cs9-bootc` and `cs10-bootc`, so this option keeps more required merge-queue work than the smallest possible slice.
- keeps some agent runtime in the required merge-queue path
  It still pays for required pre-merge coverage of device status, OCI prefetch, embedded-app update, rollback, and systemd reporting.

### Option 2: aggressive sub-15-minute merge queue

Use this only if the priority is a strict sub-15-minute target.

Keep in `merge_group`:

- only the control-plane smoke subset:
  - `cli` `(about 3.4 min total)`
    Keeps the most common control-plane create, update, delete, and validation flows in the required merge-queue path.
  - `configuration` `(about 2.4 min total)`
    Keeps basic config application and reconciliation coverage in the required merge-queue path.
  - `resourcesync` `(about 1.3 min total)`
    Keeps repository-to-fleet sync behavior in the required merge-queue path.
  - `selectors` `(about 1.2 min total)`
    Keeps core fleet and device selection logic in the required merge-queue path.
- optionally 1 to 2 very small agent canary specs, for example:
  - `75991` `(about 1.0 min per spec)`
    Keeps one minimal device-status signal in the required merge-queue path.
  - `82481` `(about 4.2 min per spec)`
    Keeps one minimal rollback signal in the required merge-queue path.

Why this is smaller:

- it keeps only the most common control-plane flows that developers hit frequently
- it removes the broader agent smoke slice from the required merge-queue path, or reduces it to a minimal canary
- it is the cleanest option if the top priority is a very small required merge-queue path

Result:

- best chance of holding the required merge-queue path under 15 minutes
- materially weaker merge-time agent coverage

Pros:

- fastest possible merge queue
- easiest required merge-queue path to keep stable and predictable

Cons:

- more agent regressions will be caught after merge rather than before merge
  If agent smoke is removed or reduced to a canary, pre-merge coverage can miss OCI prefetch, systemd status reporting, embedded-app update, and part of rollback behavior.
- higher dependence on Jenkins or optional post-merge GitHub validation
  Broader pre-merge coverage for rootless applications, backup and restore, quadlets, containers, alertmanager proxy, TPM, and most agent flows is no longer in the required merge-queue path.

## Suggested talk track

1. GitHub is currently blocking merges with a 68-spec E2E slice, and all of it already exists in Jenkins ACM.
2. The problem is not lack of coverage. The problem is that the required merge-queue path is too broad.
3. The biggest bottleneck is the non-agent GitHub sanity leg, not the agent OS matrix.
4. Jenkins should stay the broad validation lane, but not become the merge-queue lane.
5. We therefore have two choices:
   - balanced 30-spec merge-safety slice
   - more aggressive sub-15-minute slice
6. The recommended default is the 30-spec option because it cuts latency without dropping too much merge-time signal.
7. If developers need one removed suite on a PR, run it through an optional GitHub workflow outside `merge_group`, not by dynamically changing the queue.

## FAQ

**Why is `sanity` still too broad if it is supposed to be the smaller subset?**

`sanity` is smaller than the full E2E inventory, but in practice it still includes several expensive and specialized suites. In the current GitHub data, that smaller subset is still large enough to create about an hour of merge-queue runtime.

**Why keep any agent coverage in `merge_group` at all?**

Because some device-agent failures are core merge-safety risks, not just broad regression risks. Keeping a very small agent slice lets GitHub still catch obvious status, update, prefetch, and rollback regressions before code reaches `main`.

**Why were those 5 specific agent specs chosen?**

They are the smallest explicit set that still covers the core agent paths this audit treats as merge safety: device status, OCI prefetch, systemd or application state reporting, update, and rollback. They also already exist in the current GitHub `sanity` inventory.

**Why keep both `cs9-bootc` and `cs10-bootc` in the balanced option?**

Because the agent slice is already small, and cross-OS differences are more likely to matter for agent behavior than for the retained control-plane groups. Keeping both OS variants preserves that protection while the slice is still small enough to be practical.

**Why is `cli` kept even though it also appears in flaky runs?**

Because `cli` is still core merge-safety coverage. The audit treats flakiness as a problem to fix, not as the main rule for what stays or goes.

**Why not just fix flakes and keep the current scope?**

Because the current problem is not only flakiness. The measured merge-queue runtime is already too high even before accounting for flakes, and flakes add extra retries on top of that.

**Why not trim Jenkins too if it is also slow and noisy?**

Because Jenkins is not the required merge-queue path for every PR. GitHub should be optimized for time-to-main, while Jenkins can remain the broader lane for slower and more environment-heavy validation.

**Why is Jenkins considered acceptable as the broad lane despite instability?**

Because the audit is not recommending Jenkins as the path every merge must wait on. Its instability is a reason not to move that breadth into GitHub merge queue, not a reason to remove broad validation altogether.

**If GitHub has no unique coverage, why keep any GitHub E2E at all?**

Because GitHub still needs a small pre-merge merge-safety signal. The fact that Jenkins also runs that coverage means GitHub should keep only a small subset, not that GitHub should have no E2E protection before merge.

**What exactly is lost before merge if we move to the 30-spec option?**

The broad non-agent suites move out of the required merge-queue path, including backup and restore, quadlets, rootless, containers, alertmanager proxy, TPM, and similar heavier areas. Broad agent coverage also moves out, but the small agent smoke slice remains.

**What exactly is lost before merge if we move to the sub-15-minute option?**

Pre-merge agent protection becomes much smaller. OCI prefetch, systemd reporting, embedded-app update, and part of rollback behavior may no longer be covered before merge unless a canary for that area is kept.

**How realistic is the `<15 min` target without adding nodes?**

It is much more realistic only with the more aggressive slice. The balanced 30-spec option may still need node rebalancing or additional capacity to stay under that target consistently.

**Are the proposed runtimes based on one run or multiple runs?**

The group timings in the audit come from multiple sampled GitHub `merge_group` runs. The per-spec brackets in the proposal are approximate values based on saved GitHub JUnit results for those retained specs.

**How much confidence do we have in the measured timing data?**

Enough to support the direction of the recommendation. The exact minute values can move run to run, but the broad conclusion is stable: the current GitHub slice is too expensive, and the biggest pain is concentrated in a small set of suites.

**Could the retained slice still regress over time and become too slow again?**

Yes. That is why the audit includes success metrics and recommends treating the retained slice as something to monitor and adjust, not as a permanent truth.

**Why use fixed slices instead of dynamically choosing tests per PR?**

Because `merge_group` is a shared queue lane, not a custom path per PR. Making it dynamic would make the queue harder to reason about and would let one PR expand the cost for everyone behind it.

**Why not run broader tests as optional GitHub PR validation instead of Jenkins?**

That is a valid supplement and the audit recommends it. But optional GitHub PR validation should sit outside the required merge-queue path; it does not replace the need for a broad validation lane somewhere.

**What should be the decision rule for future tests: GitHub or Jenkins?**

If a test is part of the smallest practical merge-safety subset, keep it in GitHub `merge_group`. If it is broader, slower, more environment-specific, or specialized, keep it in Jenkins or an optional GitHub validation path.

**What concrete workflow changes are needed to implement Option 1?**

Limit the non-agent GitHub scope to `cli`, `configuration`, `resourcesync`, and `selectors`, and limit the agent GitHub scope to the 5 retained agent smoke IDs across `cs9-bootc` and `cs10-bootc`.

**What concrete workflow changes are needed to implement Option 2?**

Keep only the retained control-plane groups in GitHub `merge_group`, optionally keep 1 or 2 agent canaries, and move the rest of the current GitHub `sanity` breadth to Jenkins or optional GitHub validation outside the required merge-queue path.
