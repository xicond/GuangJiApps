# Stage 1: Build dependencies and compile/install pymssql
FROM python:3.11-alpine AS builder

RUN apk add --no-cache gcc musl-dev freetds-dev
RUN pip install --no-cache-dir pymssql

# Stage 2: Minimal production image
FROM python:3.11-alpine

# Install only the lightweight runtime dependency required by pymssql (FreeTDS libsybdb)
RUN apk add --no-cache freetds

# Copy global site-packages directly without virtual environment
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages

WORKDIR /backup