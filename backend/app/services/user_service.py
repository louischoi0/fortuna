from sqlalchemy.orm import Session
from typing import Optional, List
from app.models.database_models import User
from app.models.schemas import UserCreate, UserUpdate
from app.services.base import BaseService
from passlib.context import CryptContext
import logging

logger = logging.getLogger(__name__)

# Password hashing context
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

class UserService(BaseService[User]):
    """User service for database operations"""
    
    def __init__(self, db: Session):
        super().__init__(User, db)
    
    def create_user(self, user: UserCreate) -> User:
        """Create a new user with hashed password"""
        hashed_password = pwd_context.hash(user.password)
        user_data = {
            "username": user.username,
            "email": user.email,
            "hashed_password": hashed_password,
            "is_active": user.is_active
        }
        return self.create(user_data)
    
    def get_by_username(self, username: str) -> Optional[User]:
        """Get user by username"""
        try:
            return self.db.query(User).filter(User.username == username).first()
        except Exception as e:
            logger.error(f"Error getting user by username {username}: {e}")
            raise
    
    def get_by_email(self, email: str) -> Optional[User]:
        """Get user by email"""
        try:
            return self.db.query(User).filter(User.email == email).first()
        except Exception as e:
            logger.error(f"Error getting user by email {email}: {e}")
            raise
    
    def authenticate_user(self, username: str, password: str) -> Optional[User]:
        """Authenticate user with username and password"""
        user = self.get_by_username(username)
        if not user:
            return None
        if not pwd_context.verify(password, user.hashed_password):
            return None
        return user
    
    def update_user(self, user_id: int, user_update: UserUpdate) -> Optional[User]:
        """Update user information"""
        user = self.get(user_id)
        if not user:
            return None
        
        update_data = {}
        if user_update.username is not None:
            update_data["username"] = user_update.username
        if user_update.email is not None:
            update_data["email"] = user_update.email
        if user_update.is_active is not None:
            update_data["is_active"] = user_update.is_active
        
        if update_data:
            return self.update(user, update_data)
        return user
    
    def is_active(self, user: User) -> bool:
        """Check if user is active"""
        return user.is_active
    
    def is_superuser(self, user: User) -> bool:
        """Check if user is superuser"""
        return user.is_superuser
