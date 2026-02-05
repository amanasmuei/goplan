import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="GoPlan Embedding Service",
    description="Generates semantic embeddings for task descriptions",
    version="1.0.0"
)

# Load model on startup
MODEL_NAME = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
model = None


class EmbedRequest(BaseModel):
    text: str


class EmbedResponse(BaseModel):
    embedding: list[float]
    dimensions: int


class BatchEmbedRequest(BaseModel):
    texts: list[str]


class BatchEmbedResponse(BaseModel):
    embeddings: list[list[float]]
    dimensions: int


@app.on_event("startup")
async def load_model():
    global model
    logger.info(f"Loading model: {MODEL_NAME}")
    model = SentenceTransformer(MODEL_NAME)
    logger.info(f"Model loaded successfully. Embedding dimensions: {model.get_sentence_embedding_dimension()}")


@app.get("/health")
async def health():
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    return {"status": "healthy", "model": MODEL_NAME}


@app.post("/embed", response_model=EmbedResponse)
async def embed(request: EmbedRequest):
    """Generate embedding for a single text."""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    if not request.text or len(request.text.strip()) == 0:
        raise HTTPException(status_code=400, detail="Text cannot be empty")

    # Preprocess text
    text = preprocess_text(request.text)

    # Generate embedding
    embedding = model.encode(text, normalize_embeddings=True)

    return EmbedResponse(
        embedding=embedding.tolist(),
        dimensions=len(embedding)
    )


@app.post("/embed/batch", response_model=BatchEmbedResponse)
async def embed_batch(request: BatchEmbedRequest):
    """Generate embeddings for multiple texts."""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    if not request.texts or len(request.texts) == 0:
        raise HTTPException(status_code=400, detail="Texts list cannot be empty")

    if len(request.texts) > 100:
        raise HTTPException(status_code=400, detail="Maximum batch size is 100")

    # Preprocess texts
    texts = [preprocess_text(t) for t in request.texts]

    # Generate embeddings
    embeddings = model.encode(texts, normalize_embeddings=True)

    return BatchEmbedResponse(
        embeddings=[e.tolist() for e in embeddings],
        dimensions=embeddings.shape[1] if len(embeddings) > 0 else 0
    )


def preprocess_text(text: str) -> str:
    """Preprocess text for embedding generation."""
    # Convert to lowercase
    text = text.lower()

    # Remove extra whitespace
    text = " ".join(text.split())

    # Truncate if too long (model has max sequence length)
    max_length = 512
    if len(text) > max_length:
        text = text[:max_length]

    return text


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
