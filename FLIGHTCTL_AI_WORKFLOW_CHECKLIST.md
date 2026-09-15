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

If the full Jira text is already available, do not look it up again. Otherwise run:

    jira issue view BUG-ID

Use RTK for normal logs and rtk proxy for exact output. Never paste secrets or customer-specific data.

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

1. Jira lookup: Chai Bot or Jira CLI. Skip this if the issue is pasted.
2. Normal local inspection: Codex/Terra.
3. Cross-component analysis: Codex/Sol or Claude/Opus.
4. More than 200 lines of logs: Gemini first; give Codex only its short findings.
5. Final verifier simplification: Ponytail lite; never remove proof or diagnostics.

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

1. Normal drafting: Luna/low or Codex/Terra.
2. Duplicate lookup: Jira CLI or Chai Bot.
3. Large or confusing logs: Gemini first.
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

If the issue is not pasted, run:

    jira issue view FEATURE-ID

Do not paste the whole repository or test folder. Codex can inspect the local checkout.

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

1. Learn the feature: Chai Bot/Jira CLI. If the issue is pasted, stay in Codex.
2. Broad architecture: Sol/high or Gemini; return only a short design summary.
3. Implement tests: Codex/Terra with Ponytail lite.
4. Review the diff: fresh Codex reviewer, optional Claude/Opus, then Ponytail full.
5. Convert tests to manual cases: Luna/low.
6. Run OCP/VM tests: Linux/OCP CI or the real environment, not Mac-only VM evidence.

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

GitHub:

    gh run view RUN_ID --json headSha,status,conclusion,workflowName,url
    gh run view RUN_ID --log-failed
    gh pr checks PR_NUMBER

GitLab:

    glab mr view MR_NUMBER --repo GROUP/REPO
    glab ci status
    glab ci trace JOB_ID

Jenkins: copy the failed stage and surrounding console log from Jenkins.

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

1. Get exact run state: gh, glab, or supplied Jenkins data.
2. More than 200 lines of logs: Gemini; return a short event timeline.
3. Normal source fix: Codex/Terra.
4. Cross-component root cause: Codex/Sol or Claude/Opus.
5. Final simplification/review: Ponytail full.
6. OCP/deployment validation: user or Linux/OCP CI.

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

1. Fetch PR/MR and threads: gh or glab.
2. Broad cross-system research: Chai Bot.
3. Very large diff/log set: Gemini; return a short risk summary.
4. Normal product-path review: Codex/Terra.
5. High-risk architecture/security review: Sol/high or Claude/Opus.
6. Final over-engineering check: Ponytail full.

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
