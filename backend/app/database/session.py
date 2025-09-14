from sqlalchemy.orm import Session
from typing import Generator
from app.database.connection import SessionLocal

def get_db_session() -> Generator[Session, None, None]:
    """
    Get database session with proper cleanup
    """
    db = SessionLocal()
    try:
        yield db
    except Exception as e:
        db.rollback()
        raise e
    finally:
        db.close()

def get_db() -> Session:
    """
    Get database session (for direct use)
    """
    return SessionLocal()
