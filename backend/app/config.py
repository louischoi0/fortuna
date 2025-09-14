import os
from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    """Application settings"""
    
    # API Settings
    api_title: str = "Fortuna Backend API"
    api_version: str = "1.0.0"
    api_description: str = "Backend API for Fortuna project"
    
    # Server Settings
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = True
    
    # Environment
    environment: str = "development"
    
    # Database Settings
    database_url: str = "postgresql://fortuna:password@localhost:5432/fortuna_db"
    database_host: str = "localhost"
    database_port: int = 5432
    database_name: str = "fortuna_db"
    database_user: str = "fortuna"
    database_password: str = "password"
    database_echo: bool = False
    
    # CORS Settings
    cors_origins: list = ["*"]
    
    class Config:
        env_file = ".env"
        case_sensitive = False

# Global settings instance
settings = Settings()
