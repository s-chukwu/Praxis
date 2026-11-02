# Praxis Test Execution Report

**Prepared by:** Nziza (QA Tester)  
**Date:** 2 November 2026  
**Version:** 0.1.0

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
| Test Execution Date | 1 November 2026 |

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

## 8. Overall Testing Assessment

Initial testing confirms that the Praxis homepage loads successfully, the Go backend health endpoint responds correctly, and the Python AI assessment service is running.

The AI service's Swagger UI is also accessible, providing documentation for the available API endpoints.

However, further testing is required to verify the individual API operations, frontend navigation, and database integration.

## 9. Conclusion

The initial QA execution has successfully verified that the main application services are running and responding to health checks.

The next testing activities will focus on:

1. Testing individual AI service endpoints.
2. Testing frontend navigation and user-facing pages.
3. Investigating the Skill Catalog 404 observation.
4. Confirming Supabase database connectivity.
5. Testing data storage and retrieval.

This report will be updated as additional test cases are executed.