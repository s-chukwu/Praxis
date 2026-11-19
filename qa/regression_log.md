# Praxis Regression Log

Each fixed bug needs a retest on the build that contains the fix. A bug is only changed to **Verified** in `bug_reports.md` after the test results Passed and indicated in the resukt column.

| Bug | Title | Retest case(s) | Build / commit | Retest date | Tester | Result |
|---|---|---|---|---|---|---|
| BUG-001 | Baseline submit not connected | TC-037 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-002 | Refresh deletes all answers | TC-036 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-003 | No message on empty boxes | TC-035 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-004 | Customer case shows empty | TC-034 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-005 | No start button for baseline | TC-034 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-006 | No landing page | TC-028 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-007 | No profile editing page | TC-053 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-008 | Weak passwords accepted | TC-020, TC-021 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-009 | Text file upload not working | TC-032 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-010 | Assessment service not connected | TC-037, TC-058 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-011 | No email verification | TC-023 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-012 | Reviewed work still under review | TC-039, TC-044 | Latest build | 19 November 2026 | Nziza | PASS |
| BUG-013 | No delete account option | TC-055 | Latest build | 19 November 2026 | Nziza | PASS |

## Regression pass after each release

1. Run Go tests, pytest and Playwright (CI does this).
2. Run TC-016 to TC-030 (auth and routing).
3. Run one full learner journey (TC-031 to TC-041) with a reviewer in the middle (TC-042 to TC-044).
4. Run TC-047 to TC-052 (role access).
5. Record the date, build and tester above.