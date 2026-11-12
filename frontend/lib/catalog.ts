// Mirrors the AI service's role standard 1.0.0 and curriculum 1.0.0. Content changes there, not here.

export type ProgramId = 'operator' | 'builder';

export const programs: Record<ProgramId, { name: string; summary: string; skills: string[] }> = {
  operator: {
    name: 'Customer Support Agent',
    summary: 'Handle chats and tickets with AI tools: check AI drafts, protect customer data, escalate the right cases.',
    skills: ['C1', 'C2', 'C3', 'C4', 'C5', 'C6'],
  },
  builder: {
    name: 'Support Chatbot Builder',
    summary: 'Everything in the agent program, plus configuring, testing and safely handing over a support chatbot.',
    skills: ['C1', 'C2', 'C3', 'C4', 'C5', 'C6', 'B1', 'B2', 'B3'],
  },
};

export const skillNames: Record<string, string> = {
  C1: 'Fact checking and comprehension',
  C2: 'Communication',
  C3: 'Workflow and escalation',
  C4: 'Privacy and AI safety',
  C5: 'AI output verification',
  C6: 'Triage and documentation',
  B1: 'Chatbot knowledge setup',
  B2: 'Handoff and escalation',
  B3: 'Chatbot testing',
};

export const courseNames: Record<string, string> = {
  'M-ROLE': 'Job policies and scope of authority',
  'M-COMM': 'AI-assisted, empathetic replies',
  'M-SAFE': 'Data privacy and prompt injection',
  'M-VERIFY': 'Fact-checking AI-generated text',
  'M-TRIAGE': 'Triage, priority and internal notes',
  'M-BOT-KB': 'Build a bounded FAQ chatbot',
  'M-BOT-HANDOFF': 'Human handoff workflows',
  'M-BOT-TEST': 'Test and improve chatbot behaviour',
  'M-CHALLENGE': 'Final integrated challenge',
};
