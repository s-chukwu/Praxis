# Demo script and fallback

## Goal

Show the candidate and reviewer flow end to end in 8 to 10 minutes.

The story is simple: a candidate starts with one real support case, a reviewer scores it against a rubric, and the candidate gets a result and a learning plan built from the gaps.

## Hard blockers before the demo

### 1. Candidate progress route is missing

`GET /v1/me/progress` returns 404 because `registerProgressRoute` is missing from `backend/internal/httpapi/router.go`.

The dashboard result story, training record, and learning plan all depend on this route. Do not run the result and plan part of the demo until it is fixed.

Backend owner fix, after `registerQueueRoute`:

```go
var progressReviews ProgressReviews
if db != nil {
    progressReviews = db
}
registerProgressRoute(mux, authn, attempts, progressReviews, plans)
```

Verify:

```bash
curl -i "$BACKEND_URL/v1/me/progress" -H "Authorization: Bearer $CANDIDATE_TOKEN"
```

Expected: `200` for a candidate, `401` without a token, `403` for a reviewer.

### 2. Debug panic route is public

`/debug/panic` is publicly reachable and returns 502. Remove it or protect it before the demo.

## Demo roles

- Presenter: narrates and controls the candidate browser.
- Reviewer: controls the reviewer browser.
- Admin: only needed for the reset fallback.
- Timekeeper: keeps each segment on schedule.

## Pre-demo checklist

Run this 30 minutes before the demo.

- Scale web, backend, and AI to `min-replicas 1`.
- Run `./scripts/smoke.sh` and confirm it passes.
- Confirm web, backend, and AI are all on image `8503c49`.
- Confirm the backend `AI_SERVICE_URL` is `https://praxis-ai.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io`.
- Confirm `GET /v1/me/progress` returns 200 for the demo candidate.
- Confirm `/debug/panic` is removed or protected.
- Prepare the demo candidate account. Use a fresh email such as `demo.candidate+<date>@example.com`.
- Test accounts are created for candidate, reviewer, admin, and fallback. The passwords are shared in the team chat, not in this file.
- Prepare the reviewer account and confirm it has the reviewer role.
- Prepare the admin account for the reset fallback.
- Confirm email confirmation is off for testing. If it is on, use pre-confirmed accounts.
- Open the candidate and reviewer browsers side by side.
- Open `/`, `/signup`, `/dashboard`, `/path/baseline`, `/reviewer`, and `/reviewer/reviews` once so nothing is cold.
- Have the demo answers from this file ready to paste.
- Have one pre-submitted baseline attempt ready as a reviewer fallback.

## Demo flow

### 0:00 to 0:45 - Landing page

Open `/`.

Say:

> Praxis starts with a real support case, not a quiz. A candidate answers one customer scenario, a trained reviewer scores it, and the learning plan only covers the skills that need work.

Show the hero, both programs, the outcomes, and the call to action.

### 0:45 to 1:45 - Sign up and name validation

Go to `/signup`.

- Enter the demo full name.
- Type a digit in the full name field and show that it is removed immediately.
- Enter the demo email and password.
- Create the account.

Because email confirmation is off during testing, the candidate goes straight to onboarding.

Say:

> The form validates names, email, and password before the account is created. The name field rejects numbers and symbols.

### 1:45 to 2:30 - Onboarding

- Click `Choose my program`.
- Choose `Customer Support Agent`.
- On the CV step, click `Skip for now`.

Say:

> The CV is optional. The baseline task still decides the plan, so the candidate is never blocked by a missing CV.

### 2:30 to 3:00 - Dashboard

The candidate lands on `/dashboard`.

Show the greeting, program, skill progress, and the next step card.

Say:

> The dashboard always shows one next step. The training record below shows where the candidate is in the journey.

Do not rely on the result story or plan on this screen until the progress route is fixed.

### 3:00 to 4:30 - Baseline task

Open the case from the next step card or go to `/path/baseline`.

- Click `Start the case`.
- Paste the four demo answers from this file.
- Show the progress strip filling up and the word count under each field.
- Refresh the page once to show the answers are restored.
- Click `Submit my answers`.

Say:

> The task has no timer. Answers autosave, so a candidate can leave and come back. After submit, the answers are locked and go to a reviewer.

### 4:30 to 7:00 - Reviewer

Switch to the reviewer browser.

- Open `/reviewer` or `/reviewer/reviews`.
- Find the demo candidate in the queue.
- Open the case.
- Score the four rubric dimensions. Use the suggested scores and evidence quotes from this file.
- Score every skill. Use the suggested skill scores from this file.
- Click `Save draft` once to show the draft is kept.
- Click `Submit review`.

Say:

> The reviewer scores the same four dimensions every candidate gets, and every skill score feeds the learning plan. Evidence quotes must come from the candidate's answers, so the review is grounded in the work.

The finalize step calls the AI service and creates the learning plan. Keep the AI service warm before the demo.

### 7:00 to 8:30 - Candidate result and plan

Switch back to the candidate browser.

- Open `/dashboard`.
- Show the result story and the training record.
- Open `/path/plan`.
- Show the result, the strongest area, the focus area, and the ordered courses.

Say:

> The candidate sees one strength and one place to start, not a wall of scores. The plan lists only the courses the review selected.

This segment requires the progress route fix.

### 8:30 to 9:00 - Close

Say:

> That is the full loop: a real case, a human review, and a plan built from evidence. The same record can show a candidate's progress to an employer later.

## Demo answers

Paste these exact answers into the baseline form. The reviewer evidence quotes must match the text.

### Customer reply

```text
Hello, I am sorry for the worry. I checked your order IT-3307 and I can see one completed payment of RWF 35,000. The second attempt failed and is waiting to be reversed, so you were not charged twice. I have asked our finance team to confirm the reversal and will update you within 24 hours.
```

### Internal note

```text
Customer reports double MoMo debit on IT-3307. Provider log shows second attempt failed, reversal pending. Escalated to finance for confirmation. No refund promised.
```

### How you would check your answer

```text
Checked order IT-3307 payment status
Checked provider log for the second attempt
Asked finance to confirm the reversal
```

### How you used AI

```text
AI draft promised a refund today. I removed that because refunds need finance approval.
```

## Suggested reviewer scores

### Rubric dimensions

| Dimension | Score | Evidence quote |
|---|---:|---|
| Accuracy | 4 | `I checked your order IT-3307 and I can see one completed payment of RWF 35,000.` |
| Judgment | 4 | `I have asked our finance team to confirm the reversal` |
| Safety | 2 | `AI draft promised a refund today. I removed that because refunds need finance approval.` |
| Communication | 4 | `I am sorry for the worry.` |

The Safety score of 2 gives the result story a clear focus area and gives the plan a reason to include the safety course.

### Skills

| Skill | Name | Score |
|---|---|---:|
| C1 | Fact checking and comprehension | 4 |
| C2 | Communication | 4 |
| C3 | Workflow and escalation | 4 |
| C4 | Privacy and AI safety | 2 |
| C5 | AI output verification | 3 |
| C6 | Triage and documentation | 4 |

If the demo uses the Support Chatbot Builder program, also score B1, B2, and B3. Use 3 for each.

## Fallback plan

### Signup or email confirmation fails

Use the pre-created candidate account. Sign in at `/login` instead of signing up.

### Baseline start fails

Use the pre-created candidate with an active baseline attempt. Continue from `/path/baseline`.

If the candidate has a stuck active attempt, reset it as the admin:

```bash
curl -s "$BACKEND_URL/v1/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Find the candidate's `baseline_attempt_id`, then void it:

```bash
curl -s -X POST "$BACKEND_URL/v1/attempts/$ATTEMPT_ID/void" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

The candidate can then start a new baseline attempt. Never commit the admin token.

### Reviewer queue is empty

Use the pre-submitted fallback candidate. Confirm the baseline status is `submitted` in the admin user list.

### Reviewer finalize fails

Retry once. If the AI service is still unavailable, use the pre-finalized fallback candidate for the result and plan segment.

### Progress route still returns 404

Stop after the reviewer submits. Do not show the candidate result, training record, or learning plan. Use screenshots or the API response as a last resort, and state clearly that the live result screen is blocked by a backend route that is still being fixed.

### A page is cold or times out

Retry once. If it still fails, switch to the fallback account or the next segment. Do not deploy during the demo.
