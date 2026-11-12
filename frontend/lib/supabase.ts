import { createClient } from '@supabase/supabase-js';

const url = process.env.NEXT_PUBLIC_SUPABASE_URL;
const anonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;

// The anon key is public by design; row-level security and our Go backend guard the data.
export const supabase = url && anonKey ? createClient(url, anonKey) : null;
