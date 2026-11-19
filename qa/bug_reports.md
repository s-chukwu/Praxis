# Praxis Bug Reports

**Date:** 15 November 2026
**Last updated:** 19 November 2026
**Version:** 0.2.0

## 1. Introduction

This document records the bugs found during QA testing of the Praxis application. Bugs were reported by the team during end-to-end testing.

Status values: Open, In progress, Fixed, Verified. Fixed means the developers made the change. Verified means QA retested it and recorded the result in `regression_log.md`. The retest cases for each bug are listed there.

## 2. Summary

| Bug ID | Title | Severity | Status |
|---|---|---|---|
| BUG-001 | Baseline submit not connected | Critical | Verified |
| BUG-002 | Refresh deletes all answers | High | Verified |
| BUG-003 | No message on empty boxes | Medium | Verified |
| BUG-004 | Customer case shows empty | Critical | Verified |
| BUG-005 | No start button for the baseline task | High | Verified |
| BUG-006 | No landing page | Medium | Verified |
| BUG-007 | No profile editing page | Medium | Verified |
| BUG-008 | Weak passwords accepted | Medium | Verified |
| BUG-009 | Text file upload not working | Low | Verified |
| BUG-010 | Assessment service not connected | Critical | Verified |
| BUG-011 | No email verification | High | Verified |
| BUG-012 | Reviewed work still shows under review | Critical | Verified |
| BUG-013 | No delete account option | High | Verified |

## 3. Bug Details

### BUG-001: Baseline submit not connected

- **Description:** Submitting the baseline task does not complete, so a candidate cannot finish it.
- **Steps to reproduce:**
  1. Sign in as a candidate.
  2. Open the baseline task.
  3. Answer the questions and select Submit.
- **Expected result:** The answers are submitted and sent to a reviewer.
- **Actual result:** The submission is not connected, so the candidate cannot finish the task.
- **Severity:** Critical
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-002: Refresh deletes all answers

- **Description:** Refreshing the baseline page removes all answers the candidate entered.
- **Steps to reproduce:**
  1. Open the baseline task.
  2. Type answers in the boxes.
  3. Refresh the page.
- **Expected result:** The answers are saved and shown again after the refresh.
- **Actual result:** All answers are deleted.
- **Severity:** High
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-003: No message on empty boxes

- **Description:** No message appears when the baseline form is submitted with empty boxes.
- **Steps to reproduce:**
  1. Open the baseline task.
  2. Leave one or more boxes empty.
  3. Select Submit.
- **Expected result:** A message tells the candidate which boxes must be filled in.
- **Actual result:** No message appears.
- **Severity:** Medium
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-004: Customer case shows empty

- **Description:** The customer case on the baseline task is blank, so the tester does not know what to answer.
- **Steps to reproduce:**
  1. Open the baseline task.
- **Expected result:** The customer case is displayed above the questions.
- **Actual result:** The customer case is empty.
- **Severity:** Critical
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-005: No start button for the baseline task

- **Description:** There is no button to start the task and see the customer case.
- **Steps to reproduce:**
  1. Open the baseline task.
- **Expected result:** A start button opens the task and shows the customer case.
- **Actual result:** No start button is shown.
- **Severity:** High
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-006: No landing page

- **Description:** The application has no landing page.
- **Steps to reproduce:**
  1. Open the application home page.
- **Expected result:** A landing page describes the product and links to sign in and sign up.
- **Actual result:** No landing page is shown.
- **Severity:** Medium
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-007: No profile editing page

- **Description:** There is no page where a user can edit their profile.
- **Steps to reproduce:**
  1. Sign in.
  2. Look for a profile or settings page.
- **Expected result:** The user can open a profile page and edit their details.
- **Actual result:** No profile editing page exists.
- **Severity:** Medium
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)
- **Resolution:** The profile page now supports editing profile fields, replacing or deleting the CV, and deleting the account. See `frontend/app/profile/page.tsx` and `backend/internal/httpapi/learner.go`.

### BUG-008: Weak passwords accepted

- **Description:** The sign-up form accepts weak passwords. It only checks that the password has 8 characters.
- **Steps to reproduce:**
  1. Open the sign-up page.
  2. Enter a password with 8 characters and no number, for example letters only.
  3. Create the account.
- **Expected result:** The form rejects the password and asks for a stronger one.
- **Actual result:** The account is created.
- **Severity:** Medium
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-009: Text file upload not working

- **Description:** The text file upload feature does not work and is marked coming soon.
- **Steps to reproduce:**
  1. Open the CV upload step.
  2. Choose a text file.
- **Expected result:** The text from the file is read and shown.
- **Actual result:** The feature shows coming soon.
- **Severity:** Low
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-010: Assessment service not connected

- **Description:** The assessment service is not connected, so answers cannot be scored and blank answers cannot be checked.
- **Steps to reproduce:**
  1. Submit a baseline task.
- **Expected result:** The assessment service checks the answers.
- **Actual result:** The assessment service is not connected.
- **Severity:** Critical
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-011: No email verification

- **Description:** An account created with a fake email address is accepted without email verification.
- **Steps to reproduce:**
  1. Open the sign-up page.
  2. Enter a fake email address and a valid password.
  3. Create the account.
- **Expected result:** The user must confirm the email address before using the account.
- **Actual result:** The account is accepted without verification.
- **Severity:** High
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)

### BUG-012: Reviewed work still shows under review

- **Description:** After a reviewer finishes the review, the candidate profile still says the work is under review, so the candidate cannot continue with the courses.
- **Steps to reproduce:**
  1. Submit a baseline task as a candidate.
  2. Complete the review as a reviewer.
  3. Sign in as the candidate and open the dashboard.
- **Expected result:** The result and the learning plan are shown, and the courses are available.
- **Actual result:** The profile still says the work is under review and the courses stay locked.
- **Severity:** Critical
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)
- **Resolution:** `registerProgressRoute` is now registered in `backend/internal/httpapi/router.go`, so the dashboard reads `/v1/me/progress` and no longer leaves reviewed work under review.

### BUG-013: No delete account option

- **Description:** There is no option to delete an account.
- **Steps to reproduce:**
  1. Sign in.
  2. Look for a delete account option.
- **Expected result:** The user can delete their account.
- **Actual result:** No delete account option exists.
- **Severity:** High
- **Status:** Verified
- **Retested:** 19 November 2026 by Nziza (see `regression_log.md`)
- **Resolution:** The profile page has a delete-account action. It calls `DELETE /v1/me`, which erases the learner data and closes the sign-in account.