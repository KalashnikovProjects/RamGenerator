FROM python:3.11.4 AS builder

COPY ../requirements.txt .

RUN pip install --user -r requirements.txt

FROM python:3.11.4-slim
WORKDIR /app

COPY --from=builder /root/.local /root/.local
COPY ../Backend/Python/ai_generator .

EXPOSE 50051

CMD python3 __main__.py