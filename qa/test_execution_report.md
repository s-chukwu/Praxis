# Praxis Test Execution Report

**Prepared by:** Nziza (QA Tester)  
**Date:** 2 November 2026  
**Version:** 0.2.0  
**Last updated:** 19 November 2026

## 1. Introduction

This document records the results of the QA testing performed on the Praxis application. It tracks executed test cases, their outcomes, observations, and areas that require further testing.

## 2. Test Environment

| Item | Details |
|---|---|
| Project | Praxis |
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
| Frontend Automated Tests | Playwright (Chromium and mobile Chrome emulation) |
| Test Execution Date | 1 November 2026 (manual), 15 November 2026 (automated) |
| Automated Test Python Version | 3.13.18 |
| Automated Test Runner | pytest 9.2.0 |

## 3. Test Execution Summary

A total of 61 test cases are recorded in `test_case.md`: 13 original cases (TC-001 to TC-013) and 48 cases added on 19 November 2026 (TC-016 to TC-064, TC-052 not used).

| Status | Number of Test Cases |
|---|---:|
| PASS | 60 |
| FAIL | 0 |
| BLOCKED | 0 |
| NOT APPLICABLE | 1 |
| NOT TESTED | 0 |
| **Total** | **61** |

The AI service automated test suite was also run on 15 November 2026. Its results are in section 4.4 and are separate from the manual test cases above.

## 4. Executed Test Results

### 4.1 Frontend Testing

| Test ID | Test Case | Result | Observation |
|---|---|---|---|
| TC-001 | Homepage Loading | PASS | Homepage loaded successfully at localhost:3000. |
| TC-002 | Backend Status Display | PASS | Homepage displayed `Backend: ok`. |
| TC-003 | Skill Catalog Navigation | NOT APPLICABLE | The current app has no standalone Skill Catalog route. Course details are on the landing page and course pages open from the plan. |

### 4.2 Backend Testing

| Test ID | Test Case | Result | Observation |
|---|---|---|---|
| TC-007 | Go Backend Health Check | PASS | `curl` returned HTTP 200 OK. Response showed status `ok`, service `praxis-backend`, and version `0.1.0`. |

Command executed:

```bash
curl -i http://localhost:8080/health
```

### 4.3 AI Service Testing

| Test ID | Test Case | Result | Observation |
|---|---|---|---|
| TC-008 | AI Service Health Check | PASS | `curl` returned HTTP 200 OK. Response showed status `ok`, service `praxis-ai-assessment`, and version `0.3.0`. |
| TC-009 | Swagger UI Accessibility | PASS | Swagger UI loaded successfully and displayed the available API endpoints. |

Command executed:

```bash
curl -i http://localhost:8000/health
```

Swagger UI accessed at:

```text
http://localhost:8000/docs
```

### 4.4 AI Service Automated Test Suite

| Item | Details |
|---|---|
| Location | `AI_Assessment/tests` |
| Command | `python -m pytest -v` |
| Python | 3.13.18 |
| Dependencies | Pinned versions from `service/requirements.txt` |
| Total tests | 313 |
| Passed | 313 |
| Failed | 0 |
| Warnings | 1 |

| Test Group | Tests | Result |
|---|---:|---|
| Integration (API, examples, skills API) | 51 | PASS |
| Quality (API contract, journeys, traceability) | 5 | PASS |
| Security (boundary, dev proxy) | 64 | PASS |
| Unit (allocation, attempt workflow, bot, capstone, comparison, content, signing proxy, Gemini, handoff, key vault, scoring, skills) | 193 | PASS |
| **Total** | **313** | **PASS** |

The one warning is a Starlette deprecation notice (`anyio.abc.BlockingPortal`) raised by the test client. It does not affect the results.

Command executed:

```bash
cd AI_Assessment
python -m pytest -v
```

### 4.5 Frontend Automated Tests (Playwright)
 
| Item | Details |
|---|---|
| Location | `frontend/e2e` |
| Configuration | `frontend/playwright.config.ts` |
| Command | `npx playwright test` |
| Projects | chromium, mobile (Pixel 7) |
| Coverage | Landing page, login page, sign-up form validation (empty fields, invalid email, weak passwords, name field), 404 routes, signed-out dashboard |
| CI | `.github/workflows/e2e.yml` |
| Result | Run on 19 November 2026. All tests passed. |
 
## 5. Not Applicable Tests
 
| Test ID | Test Case | Status | Reason |
|---|---|---|---|
| TC-003 | Skill Catalog Navigation | NOT APPLICABLE | The current app has no standalone Skill Catalog route. Course details are on the landing page and course pages open from the plan. |

## 6. Tests Executed After the First Round

The following tests were not run in the first round. They have now been run and passed:

| Test ID | Test Case | Testing Area |
|---|---|---|
| TC-004 | Assessment Navigation | Frontend |
| TC-005 | Login Page Navigation | Frontend |
| TC-006 | Signup Page Navigation | Frontend |
| TC-010 | API Input Validation | AI Service |
| TC-011 | AI Assessment Scoring | AI Service |
| TC-012 | AI Assessment Comparison | AI Service |
| TC-013 | Learning Plan Generation | AI Service |

These are also covered by the following cases added to `test_case.md`: TC-004 by TC-034, TC-005 by TC-016, TC-006 by TC-017, TC-010 by TC-056, TC-011 by TC-058, TC-012 by TC-059 and TC-013 by TC-060.

## 7. Issues and Observations

### Observation 1: Skill Catalog Navigation

- **Test ID:** TC-003
- **Observation:** Clicking Skill Catalog displayed a 404 page in an earlier build.
- **Current Status:** Not applicable. The current app has no standalone Skill Catalog route; course details are on the landing page and course pages open from the plan.
- **Next Action:** None. Use the landing page and plan course links for the course flow.

### Observation 2: Missing Azure packages when running the AI service tests

- **Test Area:** AI service automated tests (`tests/unit/test_key_vault.py`)
- **Observation:** Collection stopped with `ModuleNotFoundError: No module named 'azure.keyvault'`, so only part of the suite was collected.
- **Cause:** `azure-identity` and `azure-keyvault-secrets` were not installed in the test environment.
- **Current Status:** Resolved. The packages were installed and all 313 tests were collected.

### Observation 3: Failures on Python 3.14

- **Test Area:** AI service automated tests
- **Observation:** On Python 3.14, `pip install -r service/requirements.txt` failed while building `pydantic-core==2.27.2`, which does not support Python 3.14. With newer unpinned libraries (FastAPI 0.162.0, Pydantic 2.14.2, Starlette 1.10.0), `tests/quality/test_contract.py::test_api_contract_only_changes_on_purpose` failed with 312 of 313 tests passing.
- **Cause:** The environment used newer library versions than the ones pinned in `service/requirements.txt` (FastAPI 0.115.6, Pydantic 2.10.4, Starlette 0.41.3). The contract test fingerprints the OpenAPI schema, which depends on those libraries. No API change was identified.
- **Current Status:** Resolved. On Python 3.13.18 with the pinned dependencies, all 313 tests pass, including the contract test.
- **Next Action:** Run AI service tests on Python 3.12 or 3.13. Do not update `CONTRACT_SHA256` to match the Python 3.14 result.

## 8. Overall Testing Assessment
 
Initial testing confirms that the Praxis homepage loads successfully, the Go backend health endpoint responds correctly, and the Python AI assessment service is running.
 
The AI service's Swagger UI is also accessible, providing documentation for the available API endpoints.
 
The AI service automated test suite (313 tests) passes on Python 3.13 with the pinned dependencies, covering the API, scoring, comparison, content handling, security boundaries, and the API contract.
 
On 19 November 2026 the QA documents were extended with 48 new test cases for the sign-up, baseline, review, plan, profile and role-access flows (added to `test_case.md`), a regression log for the 13 reported bugs (`regression_log.md`), and Playwright end-to-end tests for the frontend. They have all been run and passed.
 
Dedicated database testing was not carried out. The database is exercised indirectly through sign-up, login and account deletion, which were confirmed working during team testing (BUG-011, BUG-013).
 
## 9. Conclusion
 
The initial QA execution has successfully verified that the main application services are running and responding to health checks.
 
All 61 test cases have been executed: 60 passed and 1 (TC-003) is not applicable. All 13 reported bugs were retested and verified (see `regression_log.md`).
 
Going forward, the regression pass in `regression_log.md` should be repeated after each release, and the Playwright tests should stay in CI.
 
This report will be updated as additional test cases are executed.