# FlightCtl AI Workflow Checklist

Use the short checklist for normal tasks. Keep the full rules in this file and use only the relevant mini-template when needed.

## What to paste into chat vs. what to do yourself

Use this table before every task.

| Checklist section | Paste into chat | Do yourself | Why |
|---|---|---|---|
| Daily checklist | Only the relevant task type and goal | Nothing by default | These are standing rules; pasting all of them wastes context. |
| Model and effort selection | Only if you want a specific model or effort level | Nothing by default | Codex can choose the smallest suitable level. |
| Compact task prompt | Yes; fill in the task, files, rules, and validation | Nothing else unless stated | This gives Codex the exact scope. |
| Start-of-task commands | Paste output only when Codex cannot access the repository or current state | Run commands when you want to provide current branch/build evidence | The commands are read-only, but output may be needed from your environment. |
| RTK rules | Paste only filtered output or an error | Run noisy commands with `rtk`; use `rtk proxy` for exact output | RTK reduces command-output context. |
| Ponytail rules | Do not paste the rules; use `@ponytail-review` when requested | Review or approve deletion/simplification suggestions | Ponytail is already installed as a tool rule. |
| E2E checklist | Paste the feature workflow block, Jira details, and acceptance criteria | Run OCP/VM/deployment tests and return complete output | Those tests require your real environment and credentials. |
| `ocp-edge-ci` checklist | Paste the issue/profile/job and desired change | Run remote CI, deployment, push, or approval actions | These actions can change shared CI state. |
| Bug verification | Paste the full Jira bug, build, and environment | Run the generated verifier and paste its output back | Only your environment can prove the deployed behavior. |
| Open Jira bug | Paste reproduction evidence and logs | Review and create the Jira issue | Creating an issue is an external write. |
| Feature to tests | Paste the feature, subtickets, criteria, and Polarion IDs | Approve tests and run environment-specific validation | Codex can write code; you own final product/QE confirmation. |
| Failed CI | Paste run URL/ID, SHA, failed job, and relevant log | Run CI/OCP-specific commands and approve pushes | Codex can diagnose locally; CI access and pushes affect shared systems. |
| PR/MR review | Usually only paste the PR/MR URL and review mode | Decide whether to post findings | Codex can fetch live read-only data; posting comments is your decision. |
| CLI rules | Paste command output only if Codex cannot run the CLI | Run or approve mutating CLI commands | Read-only lookup is safe; comments, transitions, pushes, and labels are external writes. |
| Cost-save trigger | Paste only the matching workflow block | Switch to Gemini/Chai Bot when warned | The full checklist should remain a reference document. |
| Final response template | Do not paste it at the start | Copy the completed result into Jira, Slack, or notes | It is a handoff format, not task context. |

### Simple rule

Paste the problem and evidence. Codex reads the local repository and prepares the plan, commands, code, or comment. You run commands that require the real deployment or shared CI environment. Then paste the complete result back.

## Daily checklist

- Codex: local code, tests, Jenkins, `ocp-edge-ci`, Polarion, CLI, and UI.
- Chai Bot: Jira, Slack, GitHub/GitLab research and cross-system lookup.
- Gemini: very large logs, must-gathers, metrics, or broad architecture.
- Use `gh`, `glab`, Jira CLI, and `git` for exact current authenticated state.
- Use RTK for compact shell output; use `rtk proxy <command>` when exact output is needed.
- Use Ponytail `lite` for coding and `full` for completed-diff review.
- Read the relevant `AGENTS.md` files before editing.
- Search `test/harness` and `testutil` before adding helpers.
- Preserve unrelated dirty-worktree changes.
- Run targeted validation before broad validation.
- Validate VM/OCP behavior on Linux or CI when macOS is not authoritative.
- Do not make external changes without explicit authorization.
- PR/MR review: use Workflow 5 and keep the checkout read-only.

## Model and effort selection

- Sol/high: architecture, unclear root cause, or risky cross-component design.
- Terra/medium: normal implementation from a clear plan.
- Luna/low: formatting, small edits, repetitive changes, or simple tests.
- Ponytail `lite`: normal coding.
- Ponytail `full`: finished-diff review.
- Ponytail `off`: architecture, investigation, or delicate debugging.

## Compact task prompt

```text
Task: [one sentence]

Repository/files: [exact path or paths]
Jira: [ID or none]
Polarion: [ID or unknown]

Do:
1. [specific action]
2. [specific action]

Rules:
- Preserve unrelated changes.
- Reuse existing helpers and constants.
- Do not make external changes without approval.

Validate:
[exact command]
```

## Start-of-task commands

Run from `/Users/eweiss/flightctl`:

```bash
rtk git status
rtk git branch --show-current
rtk git log -1 --oneline
```

Read only the relevant instructions:

```bash
sed -n '1,240p' AGENTS.md
sed -n '1,240p' test/AGENTS.md
sed -n '1,240p' test/e2e/AGENTS.md
sed -n '1,240p' test/e2e/GUIDELINES.md
```

## RTK rules

Use RTK for routine command output:

```bash
rtk git status
rtk git diff
rtk rg "pattern" path
rtk go test ./path/...
rtk gh pr view NUMBER
rtk gh run view RUN_ID
rtk oc get pods
rtk oc logs POD
```

Use the normal command or `rtk proxy <command>` when output is truncated, exact logs are required, CI or release evidence is being collected, or output and exit status disagree.

Do not treat condensed output as complete proof for difficult failures.

## Ponytail rules

Before writing code, ask:

- Does this already exist?
- Can I reuse a harness, utility, constant, or dependency?
- Can the standard library or native platform handle it?
- Is this abstraction or helper necessary?
- What is the smallest correct diff?

Never remove required validation, diagnostics, cleanup, security, accessibility, labels, or coverage.

After implementation:

```text
@ponytail-review
```

## E2E checklist

- Read `test/AGENTS.md`, `test/e2e/AGENTS.md`, and the relevant guidelines.
- Find the Jira and Polarion IDs.
- Add the Polarion label to every `It`.
- Add `sanity` only if total upstream E2E time remains below 40 minutes.
- Add `Agent` only when the test runs on both cs9 and cs10.
- Search `test/harness` and `testutil` before creating helpers.
- Put constants and variables above tests.
- Put helper functions at the bottom.
- Reuse timeout constants.
- Use one harness call in `BeforeEach`.
- Keep `Expect` inside tests, not helpers.
- Do not create inline helper functions.
- Test successful output such as `200 OK` or `201 Created`.
- Add logging and nil/empty/error handling.
- Test OCP when the feature is expected to work on OCP.
- Use the existing VM-pool pattern for VM tests.
- Keep raw Kubernetes clients and commands inside the approved infrastructure layer.

### E2E search commands

```bash
rg -n "func |Create|WaitFor|RunGet|ManageResource|CleanUp|Harness" test/harness testutil test/e2e
rg -n "Label\\(|sanity|Agent|200 OK|201 Created" test/e2e
rg -n "time\\.Second|BeforeEach|BeforeSuite" test/e2e/[suite]
```

### E2E validation commands

```bash
go test ./test/e2e/[suite] -run '^$'
git diff --check
```

```bash
DISCOVERY_ONLY=true \
DISCOVERY_PATH=/tmp/flightctl-e2e-discovery.json \
test/scripts/run_e2e_tests.sh reports ./test/e2e/[suite]
```

## `ocp-edge-ci` checklist

```bash
cd /Users/eweiss/flightctl/ocp-edge-ci
rtk git status
rtk git fetch upstream
rtk git log -1 --oneline
```

Before changing a profile:

- Classify it as latest/main or fixed-release.
- Confirm the backend tag.
- Confirm `flightctl_repo_branch`.
- Confirm that fixed releases use the matching release tag and repository ref.
- Check OCP, ACM, FIPS, Quadlet, disconnected, upgrade, and database variants.
- Verify the failing CI run SHA matches the current branch or PR head.

```bash
rg -n "flightctl_repo_branch|FLIGHTCTL_BACKEND_TAG|FLIGHTCTL_BACKEND_PREVIOUS_RELEASE_TAG|ocp-flightctl-gotests" ci-profiles-new
```

## Full workflow templates

Use only the workflow that matches the request. Do not paste this entire document into every prompt.

### Workflow 1: Verify an existing Jira bug

#### Best routing

- Jira history or current comments: Chai Bot or Jira CLI.
- Local code, deployment, CLI, or test verification: Codex, Terra/medium.
- Large logs or must-gather data: Gemini first, then Codex for the exact verifier.
- Shell output: RTK; use `rtk proxy` for exact evidence.
- Use Ponytail `lite` only to keep the verifier minimal; do not remove diagnostic checks.

#### Send this request

```text
Verify Jira bug [BUG-ID].

I am providing the complete Jira issue and comments below.

Do not fix the bug yet.
First determine:
1. What behavior is broken.
2. What behavior proves the bug is fixed.
3. Which FlightCtl component is involved.
4. Which exact build, image, tag, or commit must be recorded.
5. Which environment is required: kind, OCP, Quadlet, VM, or other.

Search the local repository and existing harness/testutil helpers before inventing commands.
Return:
1. A short verification plan.
2. One complete paste-ready shell block with no unresolved placeholders.
3. The exact expected PASS evidence.
4. A Jira verification comment after the block.

Jira issue and comments:
[PASTE THE COMPLETE JIRA ISSUE HERE]
```

#### Required verification pasteblock format

The assistant must replace every bracketed value before returning the block.

```bash
#!/usr/bin/env bash
set -euo pipefail

BUG_ID="[BUG-ID]"
BUILD="[EXACT FLIGHTCTL BUILD, TAG, IMAGE, OR COMMIT]"
ENVIRONMENT="[OCP/KIND/QUADLET/VM]"

echo "BUG: ${BUG_ID}"
echo "BUILD: ${BUILD}"
echo "ENVIRONMENT: ${ENVIRONMENT}"

echo "--- build proof ---"
[EXACT COMMAND THAT PROVES THE DEPLOYED BUILD]

echo "--- preconditions ---"
[EXACT COMMANDS THAT PROVE REQUIRED SERVICES, DEVICES, OR RESOURCES]

echo "--- reproduce or verify ---"
[EXACT COMMANDS THAT EXERCISE THE BUG OR FIX]

echo "--- positive result ---"
[EXACT COMMAND THAT PROVES THE EXPECTED SUCCESSFUL RESULT]

echo "PASS: ${BUG_ID} verified on ${BUILD} in ${ENVIRONMENT}"
```

Do not return a generic block containing `[EXACT COMMAND]`. If the command cannot be determined, stop and ask for the missing environment or build information.

#### Jira verification comment

```markdown
Verified [BUG-ID] on FlightCtl build [EXACT BUILD/TAG/IMAGE/COMMIT].

Environment: [OCP/KIND/QUADLET/VM and version]
Date: [YYYY-MM-DD]

Verification steps:
1. [Short step]
2. [Short step]
3. [Short step]

Build evidence:
- [Exact command or deployment/image evidence]

Observed result:
- [What was observed]

Expected result:
- [What should happen]

Result: PASS — [one-sentence conclusion]
```

### Workflow 2: Open a Jira bug

#### Best routing

- Use the supplied evidence and local CLI/repository facts.
- Use Chai Bot or Jira CLI only to check for duplicates or attach current issue context.
- Do not create the Jira issue until the user reviews the draft.
- No expensive model is needed unless the evidence is large or unclear.

#### Send this request

```text
Draft a Jira bug from the information below.

Do not create the Jira issue.
Return only:
1. Missing information I must provide.
2. A Jira-ready description using the exact format below.
3. A short suggested title.

Use simple factual language. Separate observed results from expected results.
Include the exact FlightCtl build, environment, CLI commands, and frequency when known.

Evidence:
[PASTE LOGS, STEPS, BUILD, ENVIRONMENT, AND OBSERVED RESULT HERE]
```

#### Jira-ready bug template

```markdown
#### Description of the problem:

[Describe the broken behavior, affected component, FlightCtl build, and environment.]

#### How reproducible:

[Always / Often / Intermittent / Once]

Reproduced: [number] out of [number] attempts.

&nbsp;Steps to reproduce:

1. [Describe the action.]  
   Command: `[exact command]`
2. [Describe the action.]  
   Command: `[exact command]`
3. [Describe the action.]  
   Command: `[exact command]`

&nbsp;Actual results:

[What actually happened. Include exact error text or output.]

&nbsp;Expected results:

[What should have happened.]
```

Before posting, check:

- Title identifies the component and failure.
- Build and environment are included.
- Steps are executable by another person.
- Actual and expected results are not mixed.
- Logs are short and relevant.
- No credentials or tokens are included.

### Workflow 3: Jira feature to automated tests to manual test cases

#### Best routing

- Jira parent, subtickets, comments, and links: Chai Bot or Jira CLI.
- Local FlightCtl implementation and test patterns: Codex, Terra/medium.
- Ambiguous architecture or many components: Sol/high or Gemini first.
- Test implementation: Ponytail `lite`; final diff review: Ponytail `full`.
- Use RTK for searches and test output; use `rtk proxy` for exact failures.

#### Send this request

```text
Implement E2E coverage for feature [FEATURE-ID].

Jira parent:
[PASTE OR IDENTIFY PARENT ISSUE]

Subtickets:
[PASTE OR IDENTIFY SUBTICKETS]

Do this in order:
1. Learn the feature from the Jira parent, subtickets, comments, and links.
2. Inspect the local FlightCtl implementation and existing test patterns.
3. Build a requirement-to-test matrix.
4. Decide whether to extend an existing suite or create a new suite.
5. Search test/harness and testutil before creating helpers.
6. Implement the automated tests.
7. Validate the tests and report exact commands and results.
8. After I approve the automated tests, translate every test case into a manual test case.

Required E2E rules:
- Find the Polarion ID for every It.
- Add sanity only after checking the under-40-minute E2E budget.
- Add Agent only when both cs9 and cs10 must run the test.
- Constants and variables above tests.
- Helpers at the bottom.
- One harness call in BeforeEach.
- No Expect inside helpers.
- No inline helper functions.
- Test positive output and errors.
- Add logging, cleanup, and nil/empty/error handling.
- Test OCP when required.
```

#### Requirement-to-test matrix

Return this before coding:

```markdown
| Requirement | Automated scenario | Suite | Environment | Labels | Expected positive result |
|---|---|---|---|---|---|
| [Requirement] | [When ... it should ...] | [path] | [kind/OCP/etc.] | [Polarion, sanity, Agent] | [exact result] |
```

#### Automated-test completion report

```text
Automated tests completed.

Changed files:
- [path]

Test cases:
- [Polarion ID] — [short behavior]

Validation:
- [command] — PASS/FAIL
- [command] — PASS/FAIL

Environment limitation:
- [None, or explain why Linux/OCP CI is required]

Manual test cases are ready to generate after approval.
```

#### Manual test case template

Create one case for each automated `It`:

```markdown
### [POLARION-ID] — [Test case title]

#### Description of test case

[What behavior this test verifies and why it matters.]

#### Preconditions

- FlightCtl build: [exact build]
- Environment: [OCP/KIND/QUADLET/VM]
- User/device/fleet state: [required state]
- Login or credentials: [safe description; never include secrets]

#### Steps

1. **[Step description]**

   CLI command:

   ```bash
   [exact command]
   ```

   Expected output:

   ```text
   [exact expected output, status, or observable result]
   ```

2. **[Step description]**

   CLI command:

   ```bash
   [exact command]
   ```

   Expected output:

   ```text
   [exact expected output, status, or observable result]
   ```

#### Final expected result

[One clear statement of the final successful behavior.]
```

Every manual step must contain a description, a runnable CLI command, and expected output. Do not write vague instructions such as “check that it works.”

### Workflow 4: Failed Jenkins or GitHub CI run

#### Best routing

- GitHub CI: use `gh` for run metadata, failed logs, checks, and commit SHA.
- GitLab CI: use `glab` for pipeline, job, trace, and artifact data.
- Jenkins: use the supplied log or Jenkins UI; use Codex for Jenkinsfiles and `ocp-edge-ci` profiles.
- Large logs: Gemini first for compression and event timeline.
- Exact source fix: Codex, Terra/medium; Sol/high for cross-component root cause.
- Use Ponytail `full` only after the root cause is understood.

#### Send this request

```text
Diagnose and fix this failed CI run.

Provider: [Jenkins/GitHub/GitLab]
Run URL or ID: [URL/ID]
Repository: [repository]
Branch or PR/MR: [branch and number]
Expected head SHA: [SHA if known]

Do this in order:
1. Verify the run belongs to the current branch or PR/MR head.
2. Identify the first real failure and the terminal failed stage.
3. Ignore later cascading failures unless they reveal a second root cause.
4. Inspect the relevant source, test, workflow, Jenkinsfile, or ci-profile.
5. Search callers and related implementations before editing.
6. Fix the smallest root cause, not only the visible symptom.
7. Run the narrowest regression test.
8. Report the exact fix, validation, and any CI-only limitation.

CI log or failure:
[PASTE LOG OR PROVIDE URL/ID]
```

#### CI commands

GitHub:

```bash
gh run view RUN_ID --json headSha,status,conclusion,workflowName,url
gh run view RUN_ID --log-failed
gh pr checks PR_NUMBER
```

GitLab:

```bash
glab ci status
glab ci view PIPELINE_ID
glab ci trace JOB_ID
```

Repository search:

```bash
rtk rg "FAILED|FAILURE|error|timeout|panic" path/to/relevant/files
rtk git log -20 --oneline
rtk git diff BASE...HEAD
```

Use `rtk proxy` for the full failed log when the compact output is insufficient.

#### CI diagnosis report

```markdown
CI diagnosis: [PASS / FIXED / BLOCKED]

Run:
- Provider: [Jenkins/GitHub/GitLab]
- URL/ID: [value]
- Head SHA: [value]
- Current expected SHA: [value]

First real failure:
- Stage/job: [value]
- Test or command: [value]
- Evidence: [short exact error]

Root cause:
[Plain-English explanation of why it failed.]

Fix:
- File: [path]
- Change: [what changed]

Validation:
- [command] — PASS/FAIL
- [command] — PASS/FAIL

Remaining limitation:
[None, or exact CI/environment limitation.]
```

### Workflow 5: Review another GitHub PR or GitLab MR

#### Best routing

- GitHub: use `gh` for the live PR, head SHA, checks, and review comments.
- GitLab: use `glab` for the live MR, head SHA, pipelines, and discussions.
- Local source and product-path understanding: Codex, Terra/medium.
- Broad multi-component review: Sol/high or Gemini first, then Codex for findings.
- Use RTK for routine metadata and searches; use `rtk proxy` for exact logs or API responses.
- Use Ponytail `full` only as a final over-engineering check after correctness review.

#### Send this request

```text
Review this FlightCtl PR/MR. Be strict but concise.

Target:
[PASTE THE GITHUB PR OR GITLAB MR URL]

Review rules:
- Keep the checkout read-only. Do not modify files.
- Fetch the live PR/MR head and refresh the base branch first.
- Confirm the fetched ref matches the current remote head SHA.
- Inspect changed files from the fetched ref with numbered lines, not a stale local checkout.
- Query existing review comments and discussions before commenting.
- Do not repeat CodeRabbit, reviewer, or already-addressed findings.
- Return only net-new actionable findings.

Before commenting, build a short mental model:
1. What FlightCtl feature is changing?
2. Is it API, agent, service/store, CLI, deploy/Helm, test harness, or E2E?
3. What user, device, or operator workflow does it affect?
4. What is the normal happy path?
5. What state is persisted, reconciled, cleaned up, or reported?
6. What existing tests cover the same area?
7. What existing helpers or patterns should this change reuse?

Read the nearest context:
- Root AGENTS.md.
- Area AGENTS.md, if present.
- Relevant docs under docs/.
- Existing tests in the same package or suite.
- Existing helpers in test/harness, test/util, and internal helpers.
- API/OpenAPI/types when user-visible contracts are affected.
- Callers and cleanup paths, not only the modified function.

Do not comment until you can explain:
- What the PR/MR is trying to prove or fix.
- How the changed code is reached.
- What would break in production, CI, upgrade, or E2E if it is wrong.
- Whether the risk is introduced by this PR/MR or pre-existing.
- Whether an existing thread already covers it.

Prefer findings about:
- Broken reconciliation or state transitions.
- API/schema/backward-compatibility problems.
- Missing cleanup or leaked resources.
- Incorrect labels, selectors, or test IDs.
- E2E assertions that only check no error.
- Helpers with hidden Expect/failures.
- Racy waits, unbounded retries, or sleeps.
- Secret or sensitive-log exposure.
- OCP/Linux-only assumptions without required validation.
- CI jobs that appear green but did not run the focused test.

Avoid:
- Style-only comments.
- Alternative designs without a concrete bug.
- Large refactors.
- Pre-existing unrelated issues.
- Findings already covered by CodeRabbit or another reviewer.
- Findings that need a long essay to justify.

FlightCtl rules:
- Prefer existing patterns.
- Do not use Expect inside helper functions.
- Put constants near the top and helpers at the bottom.
- Avoid inline helper functions inside tests.
- If checking only no error, also assert positive output or state.
- Harden helpers for nil, empty, error, and useful logging cases.

Output exactly one of these formats:

If strong findings exist:
path/to/file.go:123 - Short human review comment explaining the concrete risk and requested fix.

If no strong findings exist:
No strong net-new findings. I checked the live head, changed files, and existing review threads.
```

#### GitHub read-only commands

```bash
gh pr view PR_NUMBER --repo flightctl/flightctl \
  --json number,title,state,baseRefName,headRefName,headRefOid,mergeCommit,statusCheckRollup,url

gh api repos/flightctl/flightctl/pulls/PR_NUMBER/comments --paginate
gh api repos/flightctl/flightctl/pulls/PR_NUMBER/reviews --paginate
gh pr checks PR_NUMBER --repo flightctl/flightctl

git fetch upstream main:refs/remotes/upstream/main
git fetch upstream +pull/PR_NUMBER/head:refs/remotes/upstream/pr-PR_NUMBER
git rev-parse refs/remotes/upstream/pr-PR_NUMBER
git diff --stat refs/remotes/upstream/main...refs/remotes/upstream/pr-PR_NUMBER
git diff --unified=80 refs/remotes/upstream/main...refs/remotes/upstream/pr-PR_NUMBER
```

Before reviewing, confirm that the SHA printed by `git rev-parse` equals the live `headRefOid` from `gh pr view`.

#### GitLab read-only commands

```bash
glab mr view MR_NUMBER --repo GROUP/REPO
glab api projects/URL_ENCODED_PROJECT/merge_requests/MR_NUMBER/discussions --paginate
glab api projects/URL_ENCODED_PROJECT/merge_requests/MR_NUMBER/notes --paginate
glab ci status

git fetch upstream main:refs/remotes/upstream/main
git fetch upstream +refs/merge-requests/MR_NUMBER/head:refs/remotes/upstream/mr-MR_NUMBER
git rev-parse refs/remotes/upstream/mr-MR_NUMBER
git diff --stat refs/remotes/upstream/main...refs/remotes/upstream/mr-MR_NUMBER
git diff --unified=80 refs/remotes/upstream/main...refs/remotes/upstream/mr-MR_NUMBER
```

If the GitLab project uses a different MR ref, obtain the live source SHA from `glab mr view` and fetch that exact SHA into `refs/remotes/upstream/mr-MR_NUMBER`.

#### Review procedure

1. Confirm the repository, PR/MR number, base branch, live head SHA, and checkout status.
2. Refresh the base branch and fetch the live head ref.
3. Read project and area instructions.
4. Read the PR/MR description, changed-file list, commits, checks, and existing threads.
5. Build the product-path mental model before judging individual lines.
6. Read surrounding callers, cleanup paths, tests, helpers, and API types.
7. Check each candidate finding against current code logic.
8. Discard findings that are speculative, pre-existing, duplicate, stylistic, or difficult to explain.
9. Verify every final path and line anchor in the fetched head ref.
10. Return only paste-ready findings.

#### Finding validation checklist

For every finding, confirm:

- It reproduces from current code logic, not guessing.
- It is introduced or made worse by this PR/MR.
- It has a concrete production, CI, upgrade, or E2E failure mode.
- The author can apply a clear fix.
- The file and line exist in the live fetched ref.
- No existing review thread already covers it.

#### Final review output

```text
path/to/file.go:123 - [Concrete risk]. [Explain the failure scenario and request the specific fix.]
```

If there are no strong findings, return exactly:

```text
No strong net-new findings. I checked the live head, changed files, and existing review threads.
```

## How to use the workflows

Do not paste this entire document into every chat. Paste only one workflow block and the information requested by that workflow.

### 1. Verify an existing Jira bug

#### You collect before starting

- Complete Jira description and comments.
- Bug ID.
- Exact FlightCtl build, tag, image, or commit.
- Environment: OCP, kind, Quadlet, VM, or other.
- Namespace or deployment details.
- Relevant logs or previous verification attempts.

Never paste passwords, tokens, certificates, or customer-sensitive data.

#### Paste this into chat

```text
WORKFLOW: VERIFY JIRA BUG

Jira bug: [BUG-ID]
FlightCtl build: [exact build/tag/image/commit]
Environment: [OCP/kind/Quadlet/VM]
Namespace or deployment: [value, if relevant]

Jira description and comments:
[PASTE THE FULL JIRA CONTENT HERE]

Additional evidence:
[PASTE RELEVANT LOGS OR PREVIOUS RESULTS]
```

#### Codex will do

- Understand the bug and comments.
- Search local code, harness, and existing tests.
- Identify the exact expected behavior.
- Create one complete verification pasteblock.
- Create a Jira verification comment.

#### You do afterward

Run the verification block against the real deployment and paste the complete output back into chat:

```text
Verification output:
[PASTE COMPLETE OUTPUT HERE]
```

Codex then writes the final Jira comment with the exact build and result.

### 2. Open a Jira bug

#### You collect before starting

- Exact build.
- Environment.
- Reproduction frequency.
- Commands used.
- Actual output.
- Expected output.
- Relevant logs.
- Duplicate candidates, if known.

#### Paste this into chat

```text
WORKFLOW: OPEN JIRA BUG

Suggested area/component: [API/agent/CLI/UI/E2E/etc.]
Build: [exact build/tag/image/commit]
Environment: [OCP/kind/Quadlet/VM]
Reproducibility: [always/often/intermittent/once]

Steps I ran:
1. [description]
   Command: [command]
2. [description]
   Command: [command]

Actual result:
[exact result or error]

Expected result:
[expected behavior]

Logs:
[relevant logs only]
```

#### Codex will do

- Separate facts from assumptions.
- Identify missing information.
- Suggest a concise title.
- Return the Jira-ready description.

#### You do afterward

- Review the title and description.
- Check that build and environment are included.
- Confirm steps are executable.
- Remove secrets and customer-specific information.
- Create the Jira issue yourself after approval.

Codex does not create, comment on, assign, or transition the issue automatically.

### 3. Jira feature to automated tests to manual test cases

#### You collect before starting

- Feature Jira URL or full text.
- Parent issue and subtickets.
- Comments containing acceptance criteria.
- Polarion IDs, if known.
- Required environments.
- Whether you want automated tests, manual cases, or both.

You do not need to paste the whole repository or test folder. Codex can inspect the local checkout.

#### Paste this into chat

```text
WORKFLOW: FEATURE TO TESTS

Feature Jira:
[URL OR FULL FEATURE TEXT]

Subtickets:
[URLS OR FULL SUBTICKET TEXT]

Acceptance criteria:
[PASTE THEM HERE]

Polarion IDs:
[IDS OR UNKNOWN]

Required environments:
[kind/OCP/Quadlet/VM]

Request:
First create and validate the automated E2E tests.
After that, translate each automated test into a manual test case.
```

#### Codex will do first

1. Read the parent, subtickets, comments, and links.
2. Inspect the local implementation.
3. Search existing E2E suites, harness, and testutil.
4. Create a requirement-to-test matrix.
5. Choose an existing suite or create a new suite.
6. Implement the automated tests.
7. Run targeted validation.
8. Report changed files and results.

#### You do after automated tests are complete

Review the automated-test report. Then send:

```text
Translate the completed automated tests into manual test cases.
Use the final test names and Polarion IDs from your previous response.
```

Every manual test case must contain:

- Title.
- Description.
- Preconditions.
- Step description.
- Exact CLI command.
- Expected output for every step.
- Final expected result.

### 4. Failed Jenkins, GitHub, or GitLab CI

#### You collect before starting

- CI provider.
- Run URL or ID.
- Branch and PR/MR number.
- Current expected head SHA.
- Failed job name.
- First real failure.
- Relevant failed log.
- Recent changes, if known.

Do not paste an entire successful CI log. Paste the failed job and surrounding context.

#### GitHub commands

```bash
gh run view RUN_ID --json headSha,status,conclusion,workflowName,url
gh run view RUN_ID --log-failed
gh pr checks PR_NUMBER
```

#### GitLab commands

```bash
glab mr view MR_NUMBER --repo GROUP/REPO
glab ci status
glab ci trace JOB_ID
```

#### Paste this into chat

```text
WORKFLOW: FIX FAILED CI

Provider: [Jenkins/GitHub/GitLab]
Run URL or ID: [URL/ID]
Repository: [repository]
Branch: [branch]
PR/MR: [number]
Expected head SHA: [SHA if known]
Failed job/stage: [value]

First failure:
[PASTE THE FIRST REAL FAILURE]

Relevant log:
[PASTE FAILED LOG SECTION]

Recent change:
[PR/MR description or changed files, if known]
```

#### Codex will do

- Verify the run belongs to the current head SHA.
- Find the first real failure.
- Separate root cause from cascading failures.
- Inspect the relevant source, workflow, Jenkinsfile, or CI profile.
- Search callers and existing patterns.
- Implement the smallest root-cause fix if requested.
- Run the narrowest regression test.
- Report remaining CI or environment limitations.

#### You do afterward

- Run any deployment-specific or CI-only command Codex provides.
- Paste the complete result back into chat.
- Approve any external push, comment, or PR/MR update separately.

### 5. Review another GitHub PR or GitLab MR

#### You collect before starting

Usually, only collect the PR/MR URL. Also state whether the review is read-only or whether you want paste-ready comments.

#### Paste this into chat

```text
WORKFLOW: REVIEW PR/MR

Target URL:
[PASTE GITHUB PR OR GITLAB MR URL]

Review mode:
[read-only findings / prepare paste-ready comments]

Specific concern:
[optional]

Return only strong net-new findings.
```

#### Codex will do

- Fetch the live head.
- Refresh the base branch.
- Confirm the live head SHA.
- Read project instructions and relevant docs.
- Read changed files with surrounding context.
- Inspect callers, cleanup paths, tests, helpers, and API types.
- Query existing review comments and discussions.
- Remove duplicate, speculative, stylistic, and pre-existing findings.
- Verify every final line anchor against the live fetched ref.

#### You do not need to paste

- The complete diff.
- Existing review threads.
- PR/MR comments.
- Changed files.

Codex can fetch those with `gh` or `glab`.

#### Review output

If findings exist:

```text
path/to/file.go:123 - [Concrete risk]. [Failure scenario and requested fix.]
```

If no findings exist:

```text
No strong net-new findings. I checked the live head, changed files, and existing review threads.
```

### Responsibility summary

- You provide the issue, URL, build, environment, and relevant evidence.
- Codex inspects local source and generates plans, commands, tests, and comments.
- You run commands that require access to the real deployment or CI environment.
- You paste command results back to Codex.
- You review and approve Jira, GitHub, GitLab, pushes, comments, and transitions.

Always paste only the matching workflow block and task-specific information.

## CLI rules

Use authenticated local tools for exact current state:

```bash
gh pr view NUMBER
gh pr checks NUMBER
glab mr view NUMBER
glab ci status
jira issue view ISSUE
git show REF
```

Do not create comments, labels, transitions, pushes, or other external changes unless explicitly requested.

## Cost-save trigger

Switch away from Codex when logs exceed roughly 200 lines, the task requires many packages or the whole repository, the task spans multiple services, a new multi-file E2E suite is needed, large Helm/RPM/YAML/JSON/must-gather data is involved, the same context is being reread repeatedly, or the same operation is needed across many files or tickets.

Use this message:

```text
⚠️ Cost-save: this is a high-context task. Switch to Gemini or Chai Bot for broad analysis. Return to Codex after the findings are narrowed to specific files and actions.
```

## Final response template

```text
Completed:
- [what changed]

Files:
- [path]

Validation:
- [command] — PASS/FAIL
- [command] — PASS/FAIL

External changes:
- None
```

If blocked:

```text
Blocked:
[one exact blocker]

Evidence:
[command or file proving it]

Needed:
[one specific thing required to continue]
```
