type Health = {
  status: string;
  service: string;
  version: string;
};

async function getBackendHealth(): Promise<Health | null> {
  const base = process.env.API_INTERNAL_URL ?? "http://localhost:8080";
  try {
    const res = await fetch(`${base}/health`, { cache: "no-store" });
    if (!res.ok) return null;
    return (await res.json()) as Health;
  } catch {
    return null;
  }
}

export default async function Home() {
  const health = await getBackendHealth();

  return (
    <main>
      <h1>Praxis</h1>
      <p>AI workforce transition prototype.</p>
      <h2>Service status</h2>
      {health ? (
        <ul>
          <li>Backend: {health.status}</li>
          <li>Service: {health.service}</li>
          <li>Version: {health.version}</li>
        </ul>
      ) : (
        <p>Backend unreachable. Is the API running?</p>
      )}
    </main>
  );
}
