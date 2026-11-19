# Praxis – Quality Assurance Test Plan

**Prepared by:** Nziza (QA Tester)  
**Date:** 1 November 2026  
**Last updated:** 19 November 2026  
**Version:** 0.2.0

## 1. Introduction

This document outlines the Quality Assurance (QA) testing strategy for the Praxis platform.

Praxis is an AI workforce transition platform designed to help learners assess their skills, identify skill gaps, improve their abilities through learning activities, verify their progress, and demonstrate their readiness for employment.

The purpose of this test plan is to ensure that the platform functions as expected, provides a good user experience, and meets its intended requirements.

## 2. Objectives

The main objectives of QA testing are to:

- Verify that the different components of Praxis work correctly.
- Identify bugs, errors, and unexpected behaviour.
- Ensure that users can navigate the platform easily.
- Verify that the backend APIs return the expected responses.
- Test the AI assessment service and its assessment-related functionality.
- Check that user inputs are handled correctly.
- Document identified issues and communicate them to developers.
- Verify that reported bugs have been fixed.

## 3. Scope of Testing

Testing will cover the following components:

### 3.1 Frontend Testing

The frontend is the part of Praxis that users interact with.

Testing will include:

- Homepage display and functionality.
- Navigation between different pages.
- Learner, Reviewer and Admin interfaces.
- Login and signup interfaces.
- Skill Catalog.
- Assessment pages.
- Page responsiveness and user experience.

### 3.2 Backend Testing

The backend handles requests and communicates with other services.

Testing will include:

- Backend health endpoint.
- API responses.
- Request validation.
- Error handling.
- Communication between the frontend and backend.
- Communication between the Go backend and Python AI service.

### 3.3 AI Assessment Service Testing

The AI service is responsible for assessment-related functionality.

Testing will include:

- Assessment-related API endpoints.
- Input validation.
- Assessment scoring.
- Skills comparison.
- Assessment allocation functionality.
- Error handling.

### 3.4 Out of Scope
 
Dedicated database testing (automated store tests and row-level security tests) is out of scope for this release. The database is exercised indirectly through sign-up, login, the baseline flow and account deletion.

## 4. Types of Testing

The following testing methods will be used:

| Testing Type | Description |
|---|---|
| Manual Testing | Manually interacting with the application to identify problems. |
| Functional Testing | Checking whether features perform their intended functions. |
| API Testing | Sending requests to backend endpoints and checking responses. |
| Integration Testing | Checking whether different services communicate correctly. |
| Smoke Testing | Checking whether the main application components start and respond. |
| Regression Testing | Retesting existing functionality after developers make changes. |
| Validation Testing | Checking how the application handles invalid or unexpected inputs. |
| Usability Testing | Evaluating whether the application is easy to understand and navigate. |
| End-to-End Testing | Automated browser tests (Playwright) that act like a user on the real pages. |
| Security Testing | Checking that each role can only reach its own pages and API endpoints. |

## 5. Testing Environment and Tools

Testing will be conducted using the following environment and tools:

| Tool | Purpose |
|---|---|
| Ubuntu Linux | Operating system used for testing. |
| Visual Studio Code | Reviewing project files and documenting test results. |
| Docker | Running the Praxis services. |
| Google Chrome | Testing the frontend interface. |
| cURL | Sending HTTP requests to backend endpoints. |
| FastAPI Swagger UI | Exploring and testing AI service endpoints. |
| Git and GitHub | Version control and sharing QA documentation. |
| Markdown | Writing test plans, test cases, and bug reports. |
| Playwright | Automated end-to-end browser tests for the frontend. |
| GitHub Actions | Runs the automated tests on every pull request. |

### Application Services

The Praxis application consists of three main services:

| Service | Technology | Local Address |
|---|---|---|
| Frontend | Next.js | http://localhost:3000 |
| Backend | Go | http://localhost:8080 |
| AI Assessment Service | Python FastAPI | http://localhost:8000 |

## 6. Testing Approach

Testing will be performed in the following stages:

### Stage 1: Environment and Smoke Testing

- Confirm that Docker builds the application successfully.
- Confirm that all three services start.
- Check the health endpoints.
- Confirm that the frontend is accessible.

### Stage 2: Frontend Testing

- Navigate through the available pages.
- Test navigation links and buttons.
- Check the Learner, Reviewer and Admin interfaces.
- Test login and signup interfaces.
- Check for broken pages and unexpected errors.

### Stage 3: Backend and API Testing

- Test available API endpoints.
- Verify HTTP response codes.
- Check request and response formats.
- Test invalid and missing inputs.
- Check communication between services.

### Stage 4: AI Assessment Testing

- Inspect available AI service endpoints.
- Test assessment-related functionality.
- Check scoring and skills comparison.
- Test invalid assessment inputs.
- Verify expected outputs against the defined requirements.

### Stage 5: Regression Testing

- Retest features after developers fix reported bugs.
- Confirm that previously working functionality still works.
- Record each retest in `regression_log.md`.
- Update test results and bug reports.

## 7. Entry Criteria

Testing can begin when:

- The project source code is available.
- The testing environment is configured.
- The application can be started locally.
- The relevant feature or API endpoint is available for testing.
- Expected behaviour is defined or agreed upon.

## 8. Exit Criteria

A testing activity can be considered complete when:

- All planned test cases have been executed or their blocked status documented.
- Test results have been recorded.
- Identified bugs have been documented.
- Critical issues have been communicated to developers.
- Fixed issues have been retested.
- A test execution report has been prepared.

Testing completion does not automatically mean that every feature is bug-free.

## 9. Bug Reporting and Documentation

All identified bugs will be documented in `bug_reports.md`.

Each bug report should contain:

- Bug ID.
- Bug title.
- Description.
- Steps to reproduce.
- Expected result.
- Actual result.
- Severity.
- Status.
- Supporting screenshots or logs where available.

Test cases and their results will be maintained in `test_case.md`. Bug retests will be recorded in `regression_log.md`.

Overall testing progress will be recorded in `test_execution_report.md`.

## 10. Current Testing Status

Status as of 19 November 2026.
 
| Component | Observation | Status |
|---|---|---|
| Frontend | Homepage loads. Playwright tests for public pages and sign-up validation pass. | Passed |
| Go Backend | Health endpoint returns `ok`. | Passed |
| AI Service | Starts successfully. The automated suite of 313 tests passes. | Passed |
| Skill Catalog | The current app has no standalone route. | Not applicable |
| Login and Signup | Work. Cases TC-016 to TC-022 pass. | Passed |
| Assessments | Baseline bugs found, fixed and retested (`regression_log.md`). | Passed |
| Reviewer and Admin | Cases TC-042 to TC-051 pass. | Passed |

## 11. Conclusion

The QA process will help ensure that Praxis meets its functional requirements and provides a reliable experience for learners and reviewers.
 
Testing will be performed continuously throughout development. Identified issues will be documented, communicated to developers, and retested after fixes are implemented.
 
The QA tester will maintain test documentation to provide evidence of testing activities and track the overall quality of the platform.