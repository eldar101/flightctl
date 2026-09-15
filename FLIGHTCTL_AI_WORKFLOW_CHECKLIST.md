# FlightCtl AI Workflow Checklist

Choose one workflow. Paste only that workflow's data block into chat.

## Contents

- [How to use this document](#how-to-use-this-document)
- [Workflow 1: Verify a Jira bug](#workflow-1-verify-a-jira-bug)
- [Workflow 2: Open a Jira bug](#workflow-2-open-a-jira-bug)
- [Workflow 3: Feature to automated and manual tests](#workflow-3-feature-to-automated-and-manual-tests)
- [Workflow 4: Fix failed CI](#workflow-4-fix-failed-ci)
- [Workflow 5: Review a GitHub PR or GitLab MR](#workflow-5-review-a-github-pr-or-gitlab-mr)
- [Common FlightCtl rules](#common-flightctl-rules)

## How to use this document

1. Choose the workflow.
2. Get only the data listed under "Get the data".
3. Paste the workflow's "Paste this into chat" block.
4. Codex inspects the local repository and prepares the plan, commands, code, or comment.
5. You run commands that require the real deployment, VM, OCP cluster, Jenkins, or shared CI.
6. Paste the complete result back.
7. Approve all external Jira, GitHub, GitLab, push, comment, or transition actions.

Do not paste this whole document into every chat.

For every command you run, paste the command and its output together. If a field says "manual", it cannot be safely inferred by a command; type the value yourself. If a command returns secrets or customer-specific data, redact them before pasting.

### Default routing

- Codex/ChatGPT: local FlightCtl code, tests, CLI, Jenkinsfiles, ocp-edge-ci, Polarion, and UI.
- Chai Bot: live Jira, Slack, GitHub/GitLab lookup, and cross-system research.
- Gemini: very large logs, must-gathers, metrics, or broad multi-component analysis.
- Claude/Opus: optional independent second opinion for high-risk design or review.
- Scripts: repeated or bulk operations.

### Effort routing

- Sol/high or Claude/Opus: architecture, unclear root cause, or high-risk review.
- Terra/medium: normal implementation, diagnosis, or review.
- Luna/low: formatting, manual-test conversion, and repetitive work.
- Ponytail lite: normal implementation.
- Ponytail full: final diff or over-engineering review.

### Cost rule

Switch to Gemini or Chai Bot when the task requires more than roughly 200 lines of logs, a whole repository, many packages, broad architecture, or repeated cross-system research. Return to Codex after the findings are reduced to specific files and actions.

## Workflow 1: Verify a Jira bug

### Purpose

Create a paste-ready verifier, run it against the real deployment, and produce a Jira verification comment.

### Get the data

- Jira bug description and comments.
- Bug ID.
- Exact FlightCtl build, tag, image, or commit.
- Environment: OCP, kind, Quadlet, VM, or other.
- Namespace/deployment details.
- Relevant previous logs.

### Commands to get the data

Jira issue and comments:

    jira issue view BUG-ID --comments 100 --plain

For machine-readable Jira data:

    jira issue view BUG-ID --comments 100 --raw

Build from a local CLI:

    bin/flightctl version

Build from an OpenShift deployment:

    oc get deployment DEPLOYMENT -n NAMESPACE -o jsonpath='{.spec.template.spec.containers[*].image}{"\n"}'

Environment and namespace:

    oc version
    oc project
    oc get pods -n NAMESPACE

Relevant pod logs:

    rtk oc logs POD -n NAMESPACE --since=30m

Use rtk proxy oc logs POD -n NAMESPACE --since=30m when exact unfiltered logs are needed.

If the full Jira text is already available, skip the Jira command above. Use RTK for normal logs and rtk proxy for exact output. Never paste secrets or customer-specific data.

### Paste this into chat

    WORKFLOW: VERIFY JIRA BUG

    Jira bug: [BUG-ID]
    FlightCtl build: [exact build/tag/image/commit]
    Environment: [OCP/kind/Quadlet/VM]
    Namespace/deployment: [value or unknown]

    Jira description and comments:
    [PASTE FULL CONTENT OR PROVIDE THE JIRA URL]

    Previous evidence:
    [PASTE RELEVANT LOGS OR RESULTS]

### Rules for the chat

- Do not fix the bug yet.
- Identify the broken behavior and the proof of success.
- Search existing harness/testutil helpers before inventing commands.
- Return one complete verifier with no unresolved placeholders.
- Include build proof, preconditions, reproduction, and a positive result.
- After verification output is returned, write the Jira comment.

### Switch at these stages

1. Jira lookup:
   - If the issue is pasted, stay in Codex and skip lookup.
   - Otherwise run: jira issue view BUG-ID --comments 100 --plain
   - Or send Chai Bot: "Read Jira BUG-ID and return only the problem, acceptance criteria, affected component, build references, environment, and relevant comments. Do not modify Jira."
2. Normal local inspection: stay in Codex/Terra.
3. Cross-component analysis: switch to Codex/Sol or Claude/Opus with: "Trace this bug across the affected FlightCtl components. Return the execution path, state changes, likely failure point, and exact evidence needed. Do not edit files."
4. More than 200 lines of logs: send Gemini: "Analyze this FlightCtl verification log. Identify the timeline, first relevant failure, environment/build evidence, and the minimum commands needed to prove the bug. Do not propose code changes. Return a concise handoff for Codex." Then paste only Gemini's handoff into Codex.
5. Final verifier simplification: run @ponytail-review or use Ponytail lite. Never remove proof or diagnostics.

### You run afterward

Run the verifier on the real deployment and paste all output back:

    Verification output:
    [PASTE COMPLETE OUTPUT HERE]

## Workflow 2: Open a Jira bug

### Purpose

Turn a reproduced failure into a Jira-ready title and description. Do not create the issue automatically.

### Get the data

You must reproduce the failure and collect the build, environment, frequency, exact commands, actual output, expected output, and relevant logs.

Use rtk proxy when exact failure text matters. Use Jira CLI or Chai Bot only to check duplicates.

### Commands to get the data

Build:

    bin/flightctl version

Environment:

    oc version
    oc project
    oc get pods -n NAMESPACE

Run each reproduction command and save the exact output:

    set -o pipefail
    [YOUR REPRODUCTION COMMAND] 2>&1 | tee /tmp/flightctl-bug-output.txt

View the saved output:

    sed -n '1,240p' /tmp/flightctl-bug-output.txt

For a deployed component, collect logs:

    rtk oc logs POD -n NAMESPACE --since=30m

Record reproducibility manually by repeating the same command and writing down the number of passes and failures. Do not invent a loop that changes the test conditions.

### Paste this into chat

    WORKFLOW: OPEN JIRA BUG

    Area/component: [API/agent/CLI/UI/E2E/etc.]
    Build: [exact build/tag/image/commit]
    Environment: [OCP/kind/Quadlet/VM]
    Reproducibility: [always/often/intermittent/once]

    Steps:
    1. [description]
       Command: [command]
    2. [description]
       Command: [command]

    Actual result:
    [exact result or error]

    Expected result:
    [expected behavior]

    Relevant logs:
    [logs only]

### Rules for the chat

- Return missing information, a suggested title, and Jira-ready Markdown.
- Separate actual from expected results.
- Include build and environment.
- Do not create, comment on, assign, or transition the issue.

### Switch at these stages

1. Normal drafting: stay in Codex/Luna or Codex/Terra.
2. Duplicate lookup:
   - Run: jira issue list --jql 'text ~ "UNIQUE_ERROR_OR_FEATURE"' --plain --no-headers --no-truncate
   - Or send Chai Bot: "Search Jira for possible duplicates of this bug. Return issue keys, summaries, status, and matching evidence only. Do not create or modify issues."
3. Large or confusing logs: send Gemini: "Summarize this bug evidence. Separate reproduction steps, actual result, expected result, build/environment, and exact error text. Do not invent facts. Return only information needed for a Jira draft."
4. No Ponytail review is needed unless code or a reproduction script is being written.

### You do afterward

Review the title and body, remove secrets/customer data, then create the Jira issue yourself.

## Workflow 3: Feature to automated and manual tests

### Purpose

Learn a Jira feature, implement E2E coverage, validate it, and translate each automated test into a manual case.

### Get the data

- Feature Jira URL or full text.
- Parent issue, subtickets, comments, and acceptance criteria.
- Polarion IDs, if known.
- Required environments: kind, OCP, Quadlet, or VM.

### Commands to get the data

Parent Jira issue:

    jira issue view FEATURE-ID --comments 100 --plain

Each known subtask:

    jira issue view SUBTASK-ID --comments 100 --plain

If you do not know the subtask IDs:

    jira issue list --jql 'parent = FEATURE-ID' --plain --no-headers --no-truncate

Then run jira issue view for each returned key.

Local implementation and existing tests:

    rg -n "FEATURE_TERM|RESOURCE_NAME|API_OPERATION" api internal pkg cmd test
    rg -n "Describe|Context|It|Label\(" test/e2e
    rg -n "func |Create|WaitFor|RunGet|ManageResource|CleanUp" test/harness testutil

Polarion references in the repository or Jira export:

    rg -n -i "polarion|POLARION-ID|FEATURE-ID" test docs .

Environment:

    oc version
    oc project
    oc get pods -n NAMESPACE

If the feature text or Jira comments are already pasted, do not run another Jira lookup. Do not paste the whole repository; Codex can read the local files.

### Paste this into chat

    WORKFLOW: FEATURE TO TESTS

    Feature Jira:
    [URL OR FULL FEATURE TEXT]

    Subtickets/comments:
    [PASTE OR IDENTIFY THEM]

    Acceptance criteria:
    [PASTE THEM HERE]

    Polarion IDs:
    [IDS OR UNKNOWN]

    Required environments:
    [kind/OCP/Quadlet/VM]

    Request:
    Create and validate the automated E2E tests first.
    After approval, translate every automated test into a manual test case.

### Rules for the chat

- Read Jira requirements and the local implementation before coding.
- Search test/harness and testutil before adding helpers.
- Prefer an existing suite; create a suite only when necessary.
- Add a Polarion label to every It.
- Add sanity only if total upstream E2E time stays below 40 minutes.
- Add Agent only when both cs9 and cs10 must run the test.
- Put constants above tests and helpers at the bottom.
- Use one harness call in BeforeEach.
- Keep Expect inside tests, not helpers.
- Test positive output/state as well as errors.
- Add logging, cleanup, and nil/empty/error handling.
- Test OCP when required.

### Switch at these stages

1. Learn the feature:
   - If the feature text is pasted, stay in Codex and skip lookup.
   - Otherwise run: jira issue view FEATURE-ID --comments 100 --plain
   - For subtasks run: jira issue view SUBTASK-ID --comments 100 --plain
   - Or send Chai Bot: "Read FEATURE-ID and its subtasks. Return the user story, acceptance criteria, dependencies, API/CLI behavior, state transitions, required environments, and Polarion references. Do not modify Jira."
2. Broad architecture: send Gemini: "Analyze this FlightCtl feature and the supplied architecture/source summary. Return only the component flow, state transitions, risks, and test scenarios. Do not write code." Then give Codex the short summary. Use Sol/high or Claude/Opus only for unresolved high-risk design questions.
3. Implement tests: stay in Codex/Terra with Ponytail lite. Run the repository searches listed above before adding helpers.
4. Review the diff: start a fresh Codex reviewer. For high-risk changes, send Claude/Opus: "Independently review this test diff for missing behavior, race conditions, cleanup problems, and false-positive assertions. Return only actionable findings." Then run @ponytail-review.
5. Convert tests to manual cases: use Codex/Luna and paste: "Translate the completed automated tests into manual test cases with title, description, preconditions, exact CLI command, and expected output for every step."
6. Run OCP/VM tests: use Linux/OCP CI or the real environment. Do not treat Mac-only VM evidence as product evidence.

### After automated tests pass

Paste:

    Translate the completed automated tests into manual test cases.
    Use the final test names and Polarion IDs from your previous response.

Every manual case must contain a title, description, preconditions, step description, exact CLI command, expected output for every step, and final expected result.

## Workflow 4: Fix failed CI

### Purpose

Find the first real failure, identify the root cause, make the smallest fix, and validate it.

### Get the data

Collect the provider, run URL/ID, branch, PR/MR, expected head SHA, failed job, first real failure, and relevant failed log. Do not paste successful log noise.

### Commands to get the data

GitHub run metadata, including the actual head SHA:

    gh run view RUN_ID --json headSha,status,conclusion,workflowName,url

GitHub branch and PR head:

    gh pr view PR_NUMBER --repo OWNER/REPO --json headRefName,headRefOid,url

Recent GitHub changes:

    gh pr diff PR_NUMBER --repo OWNER/REPO

GitHub failed log saved locally:

    gh run view RUN_ID --log-failed > /tmp/failed-github-run.log
    rtk rg "FAILED|FAILURE|error|timeout|panic" /tmp/failed-github-run.log
    sed -n '1,240p' /tmp/failed-github-run.log

GitLab MR and current SHA:

    glab mr view MR_NUMBER --repo GROUP/REPO

Recent GitLab changes:

    glab mr diff MR_NUMBER --repo GROUP/REPO --raw

GitLab pipeline and job output:

    glab ci status
    glab ci trace JOB_ID > /tmp/failed-gitlab-job.log
    rtk rg "FAILED|FAILURE|error|timeout|panic" /tmp/failed-gitlab-job.log
    sed -n '1,240p' /tmp/failed-gitlab-job.log

Jenkins: copy the failed stage name, build URL, head SHA, and the console-log section around the first failure from the Jenkins UI. If the Jenkins log is long, paste only the first failure and surrounding context.

### Paste this into chat

    WORKFLOW: FIX FAILED CI

    Provider: [Jenkins/GitHub/GitLab]
    Run URL or ID: [value]
    Repository: [value]
    Branch and PR/MR: [value]
    Expected head SHA: [value]
    Failed job/stage: [value]

    First failure:
    [PASTE FIRST REAL FAILURE]

    Relevant log:
    [PASTE FAILED LOG SECTION]

    Recent change:
    [PR/MR description or changed files]

### Rules for the chat

- Verify the run belongs to the current head SHA.
- Identify the first real failure, not only the last error.
- Separate root cause from cascading failures.
- Inspect callers, tests, workflows, Jenkinsfiles, and CI profiles.
- Fix the root cause, not only the visible symptom.
- Do not push or comment without approval.

### Switch at these stages

1. Get exact run state:
   - GitHub terminal: gh run view RUN_ID --json headSha,status,conclusion,workflowName,url
   - GitLab terminal: glab ci status and glab ci trace JOB_ID
   - Jenkins: copy the failed stage and console section from Jenkins.
2. More than 200 lines of logs: send Gemini: "Analyze this CI failure log. Return the run timeline, head SHA, first real failure, failed stage, likely root cause, and cascading failures. Do not suggest code changes. Keep the handoff under 30 lines." Then paste only the handoff into Codex.
3. Normal source fix: stay in Codex/Terra.
4. Cross-component root cause: use Codex/Sol or Claude/Opus with: "Independently trace this CI failure through the affected FlightCtl components and identify the smallest root-cause fix. Do not edit files. Return evidence and affected paths."
5. Final simplification/review: run @ponytail-review or use Ponytail full.
6. OCP/deployment validation: run in the real environment or Linux/OCP CI; do not use Mac-only VM results as final proof.

### ocp-edge-ci check

Before changing a profile, classify it as latest/main or fixed-release. Fixed releases must use the matching release tag and flightctl_repo_branch.

    cd /Users/eweiss/flightctl/ocp-edge-ci
    rtk git status
    rg -n "flightctl_repo_branch|FLIGHTCTL_BACKEND_TAG|FLIGHTCTL_BACKEND_PREVIOUS_RELEASE_TAG|ocp-flightctl-gotests" ci-profiles-new

### You do afterward

Run CI-only or deployment-specific commands provided by Codex. Paste the complete result back. Approve any push, comment, or PR/MR update separately.

## Workflow 5: Review a GitHub PR or GitLab MR

### Purpose

Review the live change read-only and return only strong, net-new, paste-ready findings.

### Get the data

Usually only the PR/MR URL and review mode are needed. Do not paste the full diff unless CLI access fails.

### Commands to get the data

GitHub PR URL, branch, and live head SHA:

    gh pr view PR_NUMBER --repo OWNER/REPO --json url,baseRefName,headRefName,headRefOid

Existing GitHub review comments:

    gh api repos/OWNER/REPO/pulls/PR_NUMBER/comments --paginate
    gh api repos/OWNER/REPO/pulls/PR_NUMBER/reviews --paginate

GitLab MR details:

    glab mr view MR_NUMBER --repo GROUP/REPO

Existing GitLab discussions:

    glab api projects/URL_ENCODED_PROJECT/merge_requests/MR_NUMBER/discussions --paginate

The review mode and specific concern are not command output. Type them directly in the paste block.

### Paste this into chat

    WORKFLOW: REVIEW PR/MR

    Target URL:
    [GITHUB PR OR GITLAB MR URL]

    Review mode:
    [read-only findings / paste-ready comments]

    Specific concern:
    [optional]

### Rules for the chat

- Keep the checkout read-only.
- Fetch the live head and refresh the base branch.
- Confirm the fetched ref matches the live remote head SHA.
- Read project instructions, docs, callers, cleanup paths, tests, helpers, and API types.
- Query existing review comments/discussions before finding issues.
- Do not repeat existing or already-addressed findings.
- Comment only on concrete correctness, security, cleanup, CI, upgrade, E2E, or convention risks.
- Verify every final line anchor in the live fetched ref.
- Return only paste-ready findings.

### Switch at these stages

1. Fetch PR/MR and threads:
   - GitHub terminal: gh pr view PR_NUMBER --repo OWNER/REPO --json headRefOid,baseRefName,headRefName,url
   - GitHub comments: gh api repos/OWNER/REPO/pulls/PR_NUMBER/comments --paginate
   - GitLab terminal: glab mr view MR_NUMBER --repo GROUP/REPO
   - GitLab discussions: glab api projects/URL_ENCODED_PROJECT/merge_requests/MR_NUMBER/discussions --paginate
2. Broad cross-system research: send Chai Bot: "Research this PR/MR and linked Jira/issues. Return only product context, prior decisions, related incidents, and existing review coverage. Do not comment or modify anything."
3. Very large diff/log set: send Gemini: "Summarize this PR/MR diff for review. Identify changed product paths, state transitions, risk areas, missing tests, and files that need close inspection. Do not produce style comments. Keep the handoff concise." Then paste only the risk summary into Codex.
4. Normal product-path review: stay in Codex/Terra.
5. High-risk architecture/security review: send Claude/Opus or use Sol/high: "Independently review this live PR/MR for correctness, security, compatibility, cleanup, and production failure modes. Return only actionable findings with file and line anchors."
6. Final over-engineering check: run @ponytail-review or use Ponytail full.

GitHub read-only commands:

    gh pr view PR_NUMBER --repo flightctl/flightctl --json headRefOid,baseRefName,headRefName,url
    gh api repos/flightctl/flightctl/pulls/PR_NUMBER/comments --paginate
    gh api repos/flightctl/flightctl/pulls/PR_NUMBER/reviews --paginate
    git fetch upstream main:refs/remotes/upstream/main
    git fetch upstream +pull/PR_NUMBER/head:refs/remotes/upstream/pr-PR_NUMBER
    git rev-parse refs/remotes/upstream/pr-PR_NUMBER

GitLab read-only commands:

    glab mr view MR_NUMBER --repo GROUP/REPO
    glab api projects/URL_ENCODED_PROJECT/merge_requests/MR_NUMBER/discussions --paginate
    git fetch upstream main:refs/remotes/upstream/main
    git fetch upstream +refs/merge-requests/MR_NUMBER/head:refs/remotes/upstream/mr-MR_NUMBER
    git rev-parse refs/remotes/upstream/mr-MR_NUMBER

### Review output

With findings:

    path/to/file.go:123 - [Concrete risk]. [Failure scenario and requested fix.]

Without findings:

    No strong net-new findings. I checked the live head, changed files, and existing review threads.

## Common FlightCtl rules

- Preserve unrelated dirty-worktree changes.
- Use RTK for normal shell output and rtk proxy for exact evidence.
- Never paste secrets or customer-specific data.
- Search existing helpers before adding code.
- Validate positive results, not only absence of errors.
- Use Linux/OCP CI for definitive VM/OCP evidence.
- Do not perform external writes without explicit approval.
- Use @ponytail-review when a final diff simplification review is useful.
