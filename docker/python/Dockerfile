FROM python:3.10-slim

# Install system dependencies (optional but recommended)
RUN apt-get update && apt-get install -y curl && apt-get clean

# Install Poetry
RUN curl -sSL https://install.python-poetry.org | python3 -

# Add Poetry to PATH
ENV PATH="/root/.local/bin:$PATH"

# Create app directory
WORKDIR /app

# Copy only pyproject first (for caching)
#COPY pyproject.toml poetry.lockx* /app/

# Copy app source
COPY . /app

# Install dependencies (no venv so Docker image stays slim)
RUN poetry config virtualenvs.create false \
    && poetry install

# Expose FastAPI port
EXPOSE 8000

# Start FastAPI
CMD ["poetry", "run", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--reload"]