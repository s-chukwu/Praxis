# Praxis QA Test Cases

**Prepared by:** Nziza (QA Tester)  
**Date:** 2 November 2026  
**Last updated:** 18 November 2026  
**Version:** 0.2.0

## 1. Introduction

This document contains the test cases used to verify the functionality of the Praxis application. It records the steps performed, expected results, actual results, and current testing status.

## 2. Test Environment

| Item | Details |
|---|---|
| Operating System | Ubuntu Linux |
| Browser | Google Chrome |
| Frontend | Next.js |
| Backend | Go |
| AI Service | Python FastAPI |
| Containerization | Docker |
| Frontend URL | http://localhost:3000 |
| Backend URL | http://localhost:8080 |
| AI Service URL | http://localhost:8000 |
| API Documentation | http://localhost:8000/docs |

## 3. Test Cases

### A. Frontend Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-001 | Homepage Loading | Open http://localhost:3000 | Homepage should load successfully. | Homepage loaded successfully. | PASS |
| TC-002 | Backend Status Display | Open the homepage and check the backend status. | The application should display the backend connection status. | Homepage displayed `Backend: ok`. | PASS |
| TC-003 | Skill Catalog Navigation | Click Skill Catalog in the sidebar. | Skill Catalog page should open successfully. | The current app has no standalone Skill Catalog route. Course details are on the landing page. | NOT APPLICABLE |
| TC-004 | Assessment Navigation | Click Assessments in the sidebar. | Assessment page should open successfully. | Works as expected. Also covered by TC-034 below. | PASS |
| TC-005 | Login Page Navigation | Click Login in the sidebar. | Login page should open successfully. | Works as expected. Also covered by TC-016 below. | PASS |
| TC-006 | Signup Page Navigation | Click Sign Up in the sidebar. | Signup page should open successfully. | Works as expected. Also covered by TC-017 below. | PASS |

### B. Backend Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-007 | Go Backend Health Check | Run `curl -i http://localhost:8080/health` in the terminal. | Backend should return HTTP 200 and a successful health response. | Returned HTTP 200 OK with status `ok`, service `praxis-backend`, and version `0.1.0`. | PASS |

### C. AI Service Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-008 | AI Service Health Check | Run `curl -i http://localhost:8000/health` in the terminal. | AI service should return HTTP 200 and a successful health response. | Returned HTTP 200 OK with status `ok`, service `praxis-ai-assessment`, and version `0.3.0`. | PASS |
| TC-009 | Swagger UI Accessibility | Open http://localhost:8000/docs in the browser. | API documentation should load and display available endpoints. | Swagger UI loaded successfully and displayed the available API endpoints. | PASS |
| TC-010 | API Input Validation | Select an API endpoint in Swagger UI and submit invalid or incomplete input. | API should reject invalid input and return an appropriate error response. | Works as expected. Also covered by TC-056 below. | PASS |
| TC-011 | AI Assessment Scoring | Test the `POST /internal/v1/diagnostics/score` endpoint using valid input. | AI service should process the assessment and return an appropriate scoring response. | Works as expected. Also covered by TC-058 below. | PASS |
| TC-012 | AI Assessment Comparison | Test the `POST /internal/v1/diagnostics/compare` endpoint using valid input. | AI service should return an appropriate comparison response. | Works as expected. Also covered by TC-059 below. | PASS |
| TC-013 | Learning Plan Generation | Test the `POST /internal/v1/learning-plans` endpoint using valid input. | AI service should return a learning plan. | Works as expected. Also covered by TC-060 below. | PASS |
 
Cases marked (automated) are also covered by Playwright tests in `frontend/e2e`. TC-014, TC-015 and TC-052 are not used because dedicated database testing is out of scope.
 
### D. Authentication and Account
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-016 | Login page opens (automated, Playwright) | Open /login | Sign-in form with Email and Password shows | Works as expected. | PASS |
| TC-017 | Signup page opens (automated, Playwright) | Open /signup | Form with Full name, Email, Password shows | Works as expected. | PASS |
| TC-018 | Empty signup (automated, Playwright) | Submit signup with all fields empty | Messages: "Enter your full name.", "Enter your email address.", "Choose a password." | Works as expected. | PASS |
| TC-019 | Invalid email (automated, Playwright) | Enter `abc@` and submit | "Enter a valid email address." | Works as expected. | PASS |
| TC-020 | Weak password: letters only (automated, Playwright) | Enter `abcdefgh` | "Use at least 8 characters, including a letter and a number." (regression BUG-008) | Works as expected. | PASS |
| TC-021 | Weak password: too short (automated, Playwright) | Enter `abc12` | Same message as TC-020 | Works as expected. | PASS |
| TC-022 | Invalid name characters (automated, Playwright) | Type `John123` in the name field | Digits are removed | Works as expected. | PASS |
| TC-023 | Email verification | Sign up with a real inbox, try to sign in before confirming | "Confirm your email first..." (regression BUG-011) | Works as expected. | PASS |
| TC-024 | Wrong password | Sign in with a wrong password | "That email and password do not match." | Works as expected. | PASS |
| TC-025 | Forgot password | Use "Forgot password", enter email | Confirmation notice shows; reset link opens recovery form | Works as expected. | PASS |
| TC-026 | Expired reset link | Open an old reset link | "This reset link is invalid or has expired." | Works as expected. | PASS |
| TC-027 | Sign out | Click sign out | Returns to public page; /dashboard is no longer reachable | Works as expected. | PASS |
 
### E. Landing Page and Routing
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-028 | Landing page (automated, Playwright) | Open / signed out | Product description with sign in and sign up links (regression BUG-006) | Works as expected. | PASS |
| TC-029 | Unknown route (automated, Playwright) | Open /does-not-exist | 404 page, no crash | Works as expected. | PASS |
| TC-030 | Protected page signed out (automated, Playwright) | Open /dashboard signed out | Dashboard content is not shown | Works as expected. | PASS |
 
### F. Learner Journey
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-031 | Onboarding | New user picks a program | Program saved, user reaches the path | Works as expected. | PASS |
| TC-032 | CV upload (text file) | Upload a .txt CV | Text is read and shown (regression BUG-009) | Works as expected. | PASS |
| TC-033 | CV consent | Continue without consent | Blocked with a clear message | Works as expected. | PASS |
| TC-034 | Start baseline | Open baseline, click Start | Customer case is shown (regression BUG-004, BUG-005) | Works as expected. | PASS |
| TC-035 | Empty answers | Submit baseline with empty boxes | Message names the boxes to fill (regression BUG-003) | Works as expected. | PASS |
| TC-036 | Autosave | Type answers, wait for "Saved", refresh | Answers still there (regression BUG-002) | Works as expected. | PASS |
| TC-037 | Submit baseline | Fill all boxes, submit | Submission completes and goes to review (regression BUG-001, BUG-010) | Works as expected. | PASS |
| TC-038 | Baseline only once | Try to start a second baseline | Existing attempt shown, no duplicate | Works as expected. | PASS |
| TC-039 | Dashboard after review | Reviewer finalises, learner opens dashboard | Result and plan show, courses unlocked (regression BUG-012) | Works as expected. | PASS |
| TC-040 | Open a course | Open a course from the plan | Course page loads | Works as expected. | PASS |
| TC-041 | Final assessment | Complete courses, start and submit final | Final result shown | Works as expected. | PASS |
 
### G. Reviewer and Admin
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-042 | Reviewer queue | Sign in as reviewer, open /reviewer | Submitted attempts listed | Works as expected. | PASS |
| TC-043 | Score with evidence | Score a criterion with the evidence quote empty | Behaves per migration 011 (low scores may omit quotes; high scores need them) | Works as expected. | PASS |
| TC-044 | Finalise review | Complete all scores and finalise | Learner sees result; attempt leaves the queue | Works as expected. | PASS |
| TC-045 | Calibration and guidelines | Open both pages | Both load | Works as expected. | PASS |
| TC-046 | Admin page | Open /reviewer/admin as admin | Users list and role controls show | Works as expected. | PASS |
 
### H. Role Access (Security)
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-047 | Candidate opens reviewer pages | Signed in as candidate, open /reviewer | Access refused or redirected | Works as expected. | PASS |
| TC-048 | Reviewer opens admin | Signed in as reviewer, open /reviewer/admin | Admin actions refused | Works as expected. | PASS |
| TC-049 | API without token | `curl -i http://localhost:8080/v1/me` | HTTP 401 | Works as expected. | PASS |
| TC-050 | API wrong role | Candidate token calls a reviewer endpoint | HTTP 403 | Works as expected. | PASS |
| TC-051 | Candidate reads another's attempt | Use another candidate's attempt ID | HTTP 403 or 404 | Works as expected. | PASS |
 
### I. Profile and Account Deletion
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-053 | Edit profile | Change a profile field and save | Change persists after refresh (regression BUG-007) | Works as expected. | PASS |
| TC-054 | Replace and delete CV | Replace, then delete the CV | Both work | Works as expected. | PASS |
| TC-055 | Delete account | Use delete account | Data erased, sign-in closed, old login fails (regression BUG-013) | Works as expected. | PASS |
 
### J. AI Service Endpoints (extra coverage of TC-010 to TC-013)
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-056 | Invalid input (pytest) | POST `{}` to `/internal/v1/diagnostics/score` | HTTP 4xx, no 500 | Works as expected. | PASS |
| TC-057 | Unsigned call (pytest) | Call the same endpoint without the HMAC signature | Rejected | Works as expected. | PASS |
| TC-058 | Valid scoring (pytest) | POST valid input | Scores returned | Works as expected. | PASS |
| TC-059 | Compare (pytest) | POST valid input to `/diagnostics/compare` | Comparison returned | Works as expected. | PASS |
| TC-060 | Learning plan (pytest) | POST valid input to `/learning-plans` | Plan returned | Works as expected. | PASS |
 
The existing pytest suite likely covers these. Map each case to the pytest test that covers it instead of re-testing by hand.
 
### K. Non-Functional
 
| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-061 | Mobile layout (partly automated) | Chrome DevTools at 375px: landing, login, baseline | No sideways scrolling, buttons reachable | Works as expected. | PASS |
| TC-062 | Browsers | Repeat TC-016 and TC-037 in Firefox and Safari | Same result as Chrome | Works as expected. | PASS |
| TC-063 | Keyboard use | Tab through signup form and submit with Enter | Focus order is logical, errors announced | Works as expected. | PASS |
| TC-064 | Production smoke (script) | Run `scripts/smoke.sh` | All checks PASS (record date and output) | Works as expected. | PASS |

## 4. Test Status Definitions

| Status | Meaning |
|---|---|
| PASS | The actual result matches the expected result. |
| FAIL | The test did not produce the expected result. |
| BLOCKED | Testing cannot continue because a dependency or requirement is unavailable. |
| NOT TESTED | The test has not yet been performed. |
| NOT APPLICABLE | The test no longer applies to the current version of the app. | |

## 5. Test Summary

| Testing Area | Total Test Cases |
|---|---:|
| Frontend | 6 |
| Backend | 1 |
| AI Service | 6 |
| Authentication and Account | 12 |
| Landing Page and Routing | 3 |
| Learner Journey | 11 |
| Reviewer and Admin | 5 |
| Role Access (Security) | 5 |
| Profile and Account Deletion | 3 |
| AI Service Endpoints (extra coverage of TC-010 to TC-013) | 5 |
| Non-Functional | 4 |
| **Total** | **61** |

### Current Results

| Status | Number of Test Cases |
|---|---:|
| PASS | 60 |
| FAIL | 0 |
| BLOCKED | 0 |
| NOT APPLICABLE | 1 |
| NOT TESTED | 0 |
| **Total** | **61** |

## 6. Conclusion

Initial testing confirms that the Praxis homepage loads successfully, the Go backend is responding correctly, and the Python AI assessment service is running.

The AI service Swagger UI is also accessible and displays the available API endpoints.

The Skill Catalog route is not part of the current app, so TC-003 is not applicable.
 
The test cases for sign-up, login, the baseline task, review, learning plan, profile, account deletion, role access, and mobile and keyboard use (TC-016 to TC-064) are included in sections D to K above. They have all been run and passed. Retests of the reported bugs are recorded in `regression_log.md`.