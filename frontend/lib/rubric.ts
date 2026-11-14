// Mirrors the AI service's rubric 0.2.0. Content changes there, not here.
export const RUBRIC_VERSION = '0.2.0';

export const DIMENSIONS = [
  { id: 'accuracy', title: 'Accuracy', weight: 30, skills: ['C1', 'C5'], evaluates: 'Correct statements, separation of facts from unknowns, evidence-backed resolution.' },
  { id: 'judgment', title: 'Judgment', weight: 25, skills: ['C3', 'C6'], evaluates: 'Proper next actions, authorization and escalation.' },
  { id: 'safety', title: 'Safety', weight: 30, skills: ['C4', 'C5'], evaluates: 'Sensitive data handling, policy compliance, refusal of unsafe instructions.' },
  { id: 'communication', title: 'Communication', weight: 15, skills: ['C2'], evaluates: 'Empathy, clarity and a professional customer response.' },
];

export const ANCHORS: [number, string][] = [
  [0, 'Absent or materially harmful: omits an essential step or includes a clearly unsafe action.'],
  [1, 'Major gaps: some relevant intent but repeated errors; needs significant supervision.'],
  [2, 'Partially adequate: correct core approach with consequential omissions or weak explanation.'],
  [3, 'Competent in simulated task: appropriate, accurate, policy-consistent response with minor limitations.'],
  [4, 'Strong, reliable execution: all level-3 evidence plus proactive verification, clear escalation and concise documentation.'],
];

// Ids match the rubric's critical_event_types, which the backend checks on every finding.
export const CRITICAL_EVENTS: [string, string, string][] = [
  ['secret_exposure', 'Secret exposure', 'Shares a password, key, token or other secret.'],
  ['unauthorized_action', 'Unauthorized action', 'Does something outside their authority, such as promising a refund finance has not approved.'],
  ['followed_injected_instruction', 'Followed injected instruction', 'Obeys instructions hidden in customer text or AI output.'],
  ['fabricated_verification', 'Fabricated verification', 'Claims to have checked something they did not.'],
  ['missed_security_escalation', 'Missed security escalation', 'Fails to escalate a clear security or fraud signal.'],
];

export const PILOT_THRESHOLD = 75;
