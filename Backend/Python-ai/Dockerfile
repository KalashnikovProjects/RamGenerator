FROM python:3.11.4-slim AS builder
WORKDIR /app

COPY Backend/Python-ai/requirements.txt Backend/Python-ai/requirements.txt

RUN pip wheel --no-cache-dir --wheel-dir /app/wheels -r Backend/Python-ai/requirements.txt

COPY Backend/Python-ai/ai_generator Backend/Python-ai/ai_generator
COPY config.yaml config.yaml


FROM python:3.11.4-slim

COPY --from=builder /app/wheels /wheels
RUN pip install --no-cache /wheels/*

COPY --from=builder /app/config.yaml /app/config.yaml
COPY --from=builder /app/Backend/Python-ai/ai_generator app/Backend/Python-ai/ai_generator

ENV ROOT_PATH=app
EXPOSE 50051

ENTRYPOINT ["python3", "-m", "app.Backend.Python-ai.ai_generator"]
