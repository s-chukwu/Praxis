// The one rule for errors in this app: only messages written for learners reach the screen.
// Anything else (network, server, library text) becomes a plain sentence, and the raw error
// goes to the console in development only, never in a production build.

export class FriendlyError extends Error {}

// A backend refusal with its machine code, so screens can react to specific cases. Its text never reaches the screen.
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    readonly details?: Record<string, unknown>,
  ) {
    super(`API ${status} ${code}`);
  }
}

export const GENERIC = 'Something went wrong on our side. Try again in a minute.';

export function report(error: unknown) {
  if (process.env.NODE_ENV !== 'production') console.error(error);
}

export function friendly(error: unknown, fallback = GENERIC) {
  report(error);
  if (error instanceof FriendlyError) return error.message;
  // fetch() rejects with a TypeError when the network is down, which on a phone is the common case.
  if (error instanceof TypeError) return 'You seem to be offline. Check your connection and try again.';
  return fallback;
}
