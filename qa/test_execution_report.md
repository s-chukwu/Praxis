# Praxis Test Execution Report

**Prepared by:** Nziza (QA Tester)  
**Date:** 2 November 2026  
**Version:** 0.1.0  
**Last updated:** 15 November 2026

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
| Test Execution Date | 1 November 2026 (manual), 15 November 2026 (automated) |
| Automated Test Python Version | 3.13.18 |
| Automated Test Runner | pytest 9.2.0 |

## 3. Test Execution Summary

A total of 15 test cases have been identified in the test case document.

| Status | Number of Test Cases |
|---|---:|
| PASS | 5 |
| FAIL | 0 |
| BLOCKED | 2 |
| PENDING INVESTIGATION | 1 |
| NOT TESTED | 7 |
| **Total** | **15** |

The AI service automated test suite was also run on 15 November 2026. Its results are in section 4.4 and are separate from the 15 manual test cases above.

## 4. Executed Test Results

### 4.1 Frontend Testing

| Test ID | Test Case | Result | Observation |
|---|---|---|---|
| TC-001 | Homepage Loading | PASS | Homepage loaded successfully at localhost:3000. |
| TC-002 | Backend Status Display | PASS | Homepage displayed `Backend: ok`. |
| TC-003 | Skill Catalog Navigation | PENDING INVESTIGATION | Clicking Skill Catalog displayed a 404 page. Further investigation is required. |

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

## 5. Blocked and Pending Tests

| Test ID | Test Case | Status | Reason |
|---|---|---|---|
| TC-014 | Database Connectivity | BLOCKED | Supabase database connectivity has not yet been confirmed. |
| TC-015 | Data Integrity | BLOCKED | Requires database connectivity to test saving and retrieving data. |
| TC-003 | Skill Catalog Navigation | PENDING INVESTIGATION | A 404 page was observed, but the expected implementation needs confirmation. |

## 6. Tests Not Yet Executed

The following tests still need to be performed:

| Test ID | Test Case | Testing Area |
|---|---|---|
| TC-004 | Assessment Navigation | Frontend |
| TC-005 | Login Page Navigation | Frontend |
| TC-006 | Signup Page Navigation | Frontend |
| TC-010 | API Input Validation | AI Service |
| TC-011 | AI Assessment Scoring | AI Service |
| TC-012 | AI Assessment Comparison | AI Service |
| TC-013 | Learning Plan Generation | AI Service |

## 7. Issues and Observations

### Observation 1: Skill Catalog Navigation

- **Test ID:** TC-003
- **Observation:** Clicking Skill Catalog displays a 404 page.
- **Current Status:** Pending Investigation.
- **Next Action:** Confirm whether the Skill Catalog page has been implemented and whether the navigation route is correct.

### Observation 2: Database Testing

- **Test IDs:** TC-014 and TC-015
- **Observation:** Database connectivity has not been confirmed.
- **Current Status:** Blocked.
- **Next Action:** Confirm Supabase configuration and connectivity with the backend development team.

### Observation 3: Missing Azure packages when running the AI service tests

- **Test Area:** AI service automated tests (`tests/unit/test_key_vault.py`)
- **Observation:** Collection stopped with `ModuleNotFoundError: No module named 'azure.keyvault'`, so only part of the suite was collected.
- **Cause:** `azure-identity` and `azure-keyvault-secrets` were not installed in the test environment.
- **Current Status:** Resolved. The packages were installed and all 313 tests were collected.

### Observation 4: Failures on Python 3.14

- **Test Area:** AI service automated tests
- **Observation:** On Python 3.14, `pip install -r service/requirements.txt` failed while building `pydantic-core==2.27.2`, which does not support Python 3.14. With newer unpinned libraries (FastAPI 0.162.0, Pydantic 2.14.2, Starlette 1.10.0), `tests/quality/test_contract.py::test_api_contract_only_changes_on_purpose` failed with 312 of 313 tests passing.
- **Cause:** The environment used newer library versions than the ones pinned in `service/requirements.txt` (FastAPI 0.115.6, Pydantic 2.10.4, Starlette 0.41.3). The contract test fingerprints the OpenAPI schema, which depends on those libraries. No API change was identified.
- **Current Status:** Resolved. On Python 3.13.18 with the pinned dependencies, all 313 tests pass, including the contract test.
- **Next Action:** Run AI service tests on Python 3.12 or 3.13. Do not update `CONTRACT_SHA256` to match the Python 3.14 result.

## 8. Overall Testing Assessment

Initial testing confirms that the Praxis homepage loads successfully, the Go backend health endpoint responds correctly, and the Python AI assessment service is running.

The AI service's Swagger UI is also accessible, providing documentation for the available API endpoints.

The AI service automated test suite (313 tests) passes on Python 3.13 with the pinned dependencies, covering the API, scoring, comparison, content handling, security boundaries, and the API contract.

However, further manual testing is required to verify the individual API operations through Swagger UI, frontend navigation, and database integration.

## 9. Conclusion

The initial QA execution has successfully verified that the main application services are running and responding to health checks.

The next testing activities will focus on:

1. Testing individual AI service endpoints.
2. Testing frontend navigation and user-facing pages.
3. Investigating the Skill Catalog 404 observation.
4. Confirming Supabase database connectivity.
5. Testing data storage and retrieval.

This report will be updated as additional test cases are executed.