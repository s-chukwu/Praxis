from fastapi import FastAPI

app = FastAPI(
    title="Praxis AI / Assessment Service",
    version="0.1.0",
    description="Internal service. Not exposed to the browser in production.",
)


@app.get("/health")
def health() -> dict:
    return {"status": "ok", "service": "praxis-ai-assessment", "version": "0.1.0"}


@app.get("/catalog")
def catalog() -> dict:
    return {
        "version": "0.1.0",
        "job_family": "customer_support",
        "competencies": [
            {"id": "C1", "name": "Fact checking and comprehension"},
            {"id": "C2", "name": "Communication"},
            {"id": "C3", "name": "Workflow and escalation"},
            {"id": "C4", "name": "Privacy and AI safety"},
            {"id": "C5", "name": "AI output verification"},
            {"id": "C6", "name": "Triage and documentation"},
        ],
    }
