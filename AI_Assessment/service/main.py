import logging

from praxis_assessment.api import create_app

logging.basicConfig(level=logging.INFO, format="%(message)s")

app = create_app()
