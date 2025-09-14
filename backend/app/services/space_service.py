from sqlalchemy.orm import Session
from typing import Optional, List
from app.models.database_models import Space
from app.models.schemas import SpaceCreate, SpaceUpdate
from app.services.base import BaseService
from app.fortuna_sdk import execute_sdk
import uuid
import logging

logger = logging.getLogger(__name__)

class SpaceService(BaseService[Space]):
    """Space service for database operations"""
    
    def __init__(self, db: Session):
        super().__init__(Space, db)
    
    def create_space(self, space: SpaceCreate, owner_id: int) -> Space:
        """Create a new space"""
        space_id = str(uuid.uuid4()).replace('-', '')
        space_data = {
            "space_id": space_id,
            "name": space.name,
            "description": space.description,
            "owner_id": owner_id,
            "config": space.config
        }
        return self.create(space_data)
    
    def get_by_space_id(self, space_id: str) -> Optional[Space]:
        """Get space by space_id"""
        try:
            return self.db.query(Space).filter(Space.space_id == space_id).first()
        except Exception as e:
            logger.error(f"Error getting space by space_id {space_id}: {e}")
            raise
    
    def get_by_owner(self, owner_id: int, skip: int = 0, limit: int = 100) -> List[Space]:
        """Get spaces by owner"""
        try:
            return self.db.query(Space).filter(
                Space.owner_id == owner_id
            ).offset(skip).limit(limit).all()
        except Exception as e:
            logger.error(f"Error getting spaces by owner {owner_id}: {e}")
            raise
    
    def get_active_spaces(self, skip: int = 0, limit: int = 100) -> List[Space]:
        """Get active spaces"""
        try:
            return self.db.query(Space).filter(
                Space.is_active == True
            ).offset(skip).limit(limit).all()
        except Exception as e:
            logger.error(f"Error getting active spaces: {e}")
            raise
    
    def update_space(self, space_id: int, space_update: SpaceUpdate) -> Optional[Space]:
        """Update space information"""
        space = self.get(space_id)
        if not space:
            return None
        
        update_data = {}
        if space_update.name is not None:
            update_data["name"] = space_update.name
        if space_update.description is not None:
            update_data["description"] = space_update.description
        if space_update.config is not None:
            update_data["config"] = space_update.config
        if space_update.is_active is not None:
            update_data["is_active"] = space_update.is_active
        
        if update_data:
            return self.update(space, update_data)
        return space
    
    def deactivate_space(self, space_id: int) -> Optional[Space]:
        """Deactivate a space"""
        return self.update_space(space_id, SpaceUpdate(is_active=False))
    
    def activate_space(self, space_id: int) -> Optional[Space]:
        """Activate a space"""
        return self.update_space(space_id, SpaceUpdate(is_active=True))
