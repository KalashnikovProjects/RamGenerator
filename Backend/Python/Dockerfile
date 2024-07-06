FROM python:3.11.4-slim AS builder
WORKDIR /app

COPY ./requirements.txt .

RUN pip install --user --no-cache-dir -r requirements.txt

COPY ./ai_generator ./ai_generator

FROM python:3.11.4-slim

COPY --from=builder /root/.local /root/.local
COPY --from=builder /app/ai_generator /app/ai_generator

EXPOSE 50051

ENTRYPOINT ["python3", "-m", "app/ai_generator"]
