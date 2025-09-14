from fastapi import APIRouter, HTTPException, Depends
from typing import Dict, Any, List
import os
from sqlalchemy.orm import Session
from app.database.connection import get_db
from app.services.oracle_service import OracleService
from app.models.schemas import OracleStatusListResponse

# Create API router
router = APIRouter(prefix="/api/v1", tags=["api"])

@router.get("/")
async def api_root():
    """API root endpoint"""
    return {
        "message": "Fortuna API v1",
        "status": "active",
        "version": "1.0.0"
    }

@router.get("/status")
async def get_api_status():
    """Get API status"""
    return {
        "status": "healthy",
        "service": "fortuna-api",
        "environment": os.getenv("ENVIRONMENT", "development")
    }

@router.get("/info")
async def get_api_info():
    """Get API information"""
    return {
        "name": "Fortuna API",
        "version": "1.0.0",
        "description": "Backend API for Fortuna project",
        "endpoints": {
            "health": "/api/v1/status",
            "oracles": "/api/v1/oracles/status",
            "docs": "/docs",
            "redoc": "/redoc"
        }
    }

@router.get("/oracles/status", response_model=OracleStatusListResponse)
async def get_oracles_status(db: Session = Depends(get_db)):
    """
    Get status of all oracle nodes
    Queries database for oracle nodes and checks their status via SDK
    """
    try:
        oracle_service = OracleService(db)
        oracle_statuses = oracle_service.get_oracle_statuses()
        
        return OracleStatusListResponse(
            message="Oracle status retrieved successfully",
            total_oracles=len(oracle_statuses),
            oracles=oracle_statuses
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Error retrieving oracle statuses: {str(e)}"
        )
