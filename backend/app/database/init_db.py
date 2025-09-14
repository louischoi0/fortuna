"""
Database initialization script
"""
import logging
from sqlalchemy import create_engine
from app.database.connection import Base, engine
from app.models import database_models
from app.config import settings

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def init_database():
    """Initialize database tables"""
    try:
        logger.info("Creating database tables...")
        Base.metadata.create_all(bind=engine)
        logger.info("Database tables created successfully!")
    except Exception as e:
        logger.error(f"Error creating database tables: {e}")
        raise

def check_database_connection():
    """Check if database connection is working"""
    try:
        with engine.connect() as connection:
            connection.execute("SELECT 1")
        logger.info("Database connection successful!")
        return True
    except Exception as e:
        logger.error(f"Database connection failed: {e}")
        return False

if __name__ == "__main__":
    logger.info("Initializing database...")
    logger.info(f"Database URL: {settings.database_url}")
    
    if check_database_connection():
        init_database()
    else:
        logger.error("Cannot initialize database - connection failed")
        exit(1)
