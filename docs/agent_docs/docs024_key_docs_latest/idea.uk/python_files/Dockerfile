# idea.uk service — engine + billing + delivery
FROM python:3.12-slim

WORKDIR /app

# Engine + service deps. stripe/openai optional but included so the image is
# complete; the service runs without them (FakeProvider / Anthropic-only critique).
RUN pip install --no-cache-dir \
    fastapi uvicorn[standard] python-multipart \
    anthropic "stripe>=9" openai

COPY idea_method_runner.py idea_service.py ./

# Persist the sqlite db on a mounted volume in production:
#   docker run -v idea_data:/data -e IDEA_DB_PATH=/data/idea_orders.db ...
ENV IDEA_DB_PATH=/data/idea_orders.db
VOLUME ["/data"]

EXPOSE 8080
CMD ["uvicorn", "idea_service:app", "--host", "0.0.0.0", "--port", "8080"]
