from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import uvicorn
from dotenv import load_dotenv
import os
from app.config import settings
from app.routers import api
from app.database.connection import check_db_connection

# Load environment variables
load_dotenv()

# Create FastAPI instance
app = FastAPI(
    title=settings.api_title,
    description=settings.api_description,
    version=settings.api_version,
    docs_url="/docs",
    redoc_url="/redoc"
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(api.router)

# Startup event
@app.on_event("startup")
async def startup_event():
    """Initialize database on startup"""
    if not check_db_connection():
        raise Exception("Database connection failed")

# Health check endpoint
@app.get("/")
async def root():
    return {"message": "Fortuna Backend API is running", "status": "healthy"}

@app.get("/health")
async def health_check():
    db_status = "connected" if check_db_connection() else "disconnected"
    return {
        "status": "healthy", 
        "service": "fortuna-backend",
        "database": db_status
    }

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        log_level="info"
    )
