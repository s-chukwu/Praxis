# Praxis QA Test Cases

**Prepared by:** Nziza (QA Tester)  
**Date:** 2 November 2026  
**Version:** 0.1.0

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
| TC-003 | Skill Catalog Navigation | Click Skill Catalog in the sidebar. | Skill Catalog page should open successfully. | Browser displayed a 404 page. Requires investigation to confirm whether the page is implemented. | PENDING INVESTIGATION |
| TC-004 | Assessment Navigation | Click Assessments in the sidebar. | Assessment page should open successfully. | Not yet tested. | NOT TESTED |
| TC-005 | Login Page Navigation | Click Login in the sidebar. | Login page should open successfully. | Not yet tested. | NOT TESTED |
| TC-006 | Signup Page Navigation | Click Sign Up in the sidebar. | Signup page should open successfully. | Not yet tested. | NOT TESTED |

### B. Backend Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-007 | Go Backend Health Check | Run `curl -i http://localhost:8080/health` in the terminal. | Backend should return HTTP 200 and a successful health response. | Returned HTTP 200 OK with status `ok`, service `praxis-backend`, and version `0.1.0`. | PASS |

### C. AI Service Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-008 | AI Service Health Check | Run `curl -i http://localhost:8000/health` in the terminal. | AI service should return HTTP 200 and a successful health response. | Returned HTTP 200 OK with status `ok`, service `praxis-ai-assessment`, and version `0.3.0`. | PASS |
| TC-009 | Swagger UI Accessibility | Open http://localhost:8000/docs in the browser. | API documentation should load and display available endpoints. | Swagger UI loaded successfully and displayed the available API endpoints. | PASS |
| TC-010 | API Input Validation | Select an API endpoint in Swagger UI and submit invalid or incomplete input. | API should reject invalid input and return an appropriate error response. | Not yet tested. | NOT TESTED |
| TC-011 | AI Assessment Scoring | Test the `POST /internal/v1/diagnostics/score` endpoint using valid input. | AI service should process the assessment and return an appropriate scoring response. | Not yet tested. | NOT TESTED |
| TC-012 | AI Assessment Comparison | Test the `POST /internal/v1/diagnostics/compare` endpoint using valid input. | AI service should return an appropriate comparison response. | Not yet tested. | NOT TESTED |
| TC-013 | Learning Plan Generation | Test the `POST /internal/v1/learning-plans` endpoint using valid input. | AI service should return a learning plan. | Not yet tested. | NOT TESTED |

### D. Database Testing

| Test ID | Test Case | Steps | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-014 | Database Connectivity | Verify that the application can connect to the Supabase database. | Application should establish a successful database connection. | Database integration testing is pending configuration. | BLOCKED |
| TC-015 | Data Integrity | Test saving and retrieving application data from the database. | Data should be stored and retrieved correctly without unexpected changes. | Cannot test until database integration is available. | BLOCKED |

## 4. Test Status Definitions

| Status | Meaning |
|---|---|
| PASS | The actual result matches the expected result. |
| FAIL | The test did not produce the expected result. |
| BLOCKED | Testing cannot continue because a dependency or requirement is unavailable. |
| NOT TESTED | The test has not yet been performed. |
| PENDING INVESTIGATION | An observation requires clarification before it can be classified as a confirmed failure. |

## 5. Test Summary

| Testing Area | Total Test Cases |
|---|---:|
| Frontend | 6 |
| Backend | 1 |
| AI Service | 6 |
| Database | 2 |
| **Total** | **15** |

### Current Results

| Status | Number of Test Cases |
|---|---:|
| PASS | 5 |
| FAIL | 0 |
| BLOCKED | 2 |
| PENDING INVESTIGATION | 1 |
| NOT TESTED | 7 |
| **Total** | **15** |

## 6. Conclusion

Initial testing confirms that the Praxis homepage loads successfully, the Go backend is responding correctly, and the Python AI assessment service is running.

The AI service Swagger UI is also accessible and displays the available API endpoints.

Further testing is required to verify individual AI endpoints, frontend navigation, input validation, and database integration. The Skill Catalog navigation has displayed a 404 page and requires further investigation before it can be classified as a confirmed bug.