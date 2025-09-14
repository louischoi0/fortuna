from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from app.models.database_models import OracleNode
from app.services.base import BaseService
from app.fortuna_sdk import execute_sdk
import logging

logger = logging.getLogger(__name__)

class OracleService(BaseService[OracleNode]):
    """Oracle service for database operations and SDK integration"""
    
    def __init__(self, db: Session):
        super().__init__(OracleNode, db)
    
    def get_all_oracles(self) -> List[OracleNode]:
        """Get all oracle nodes from database"""
        try:
            return self.db.query(OracleNode).filter(OracleNode.is_active == True).all()
        except Exception as e:
            logger.error(f"Error getting all oracles: {e}")
            raise
    
    def get_oracle_statuses(self) -> List[Dict[str, Any]]:
        """
        Get oracle status by querying each oracle node via SDK
        Returns list of oracle status information
        """
        oracles = self.get_all_oracles()
        oracle_statuses = []
        
        for oracle in oracles:
            try:
                logger.info(f"Checking status for oracle {oracle.node_id} at {oracle.ip_address}:{oracle.port}")
                
                # Execute SDK command to get oracle status
                result = execute_sdk("oracle", "status", oracle.ip_address, str(oracle.port))
                oracle_status = json.loads(result)
                
                oracle_statuses.append(oracle_status)
                logger.info(f"Oracle {oracle.node_id} status: {oracle_status['status']}")
                
            except Exception as e:
                logger.error(f"Error checking status for oracle {oracle.node_id}: {e}")
                
                oracle_status = {
                    "node_id": oracle.node_id,
                    "ip_address": oracle.ip_address,
                    "port": oracle.port,
                    "is_active": oracle.is_active,
                    "last_heartbeat": oracle.last_heartbeat.isoformat() if oracle.last_heartbeat else None,
                    "status": "error",
                    "sdk_response": None,
                    "error": str(e)
                }
                
                oracle_statuses.append(oracle_status)
        
        return oracle_statuses
    
    def get_oracle_by_node_id(self, node_id: str) -> Optional[OracleNode]:
        """Get oracle by node_id"""
        try:
            return self.db.query(OracleNode).filter(OracleNode.node_id == node_id).first()
        except Exception as e:
            logger.error(f"Error getting oracle by node_id {node_id}: {e}")
            raise
    
    def get_oracle_by_address(self, ip_address: str, port: int) -> Optional[OracleNode]:
        """Get oracle by IP address and port"""
        try:
            return self.db.query(OracleNode).filter(
                OracleNode.ip_address == ip_address,
                OracleNode.port == port
            ).first()
        except Exception as e:
            logger.error(f"Error getting oracle by address {ip_address}:{port}: {e}")
            raise
    
    def update_heartbeat(self, node_id: str) -> Optional[OracleNode]:
        """Update oracle heartbeat timestamp"""
        try:
            oracle = self.get_oracle_by_node_id(node_id)
            if oracle:
                from datetime import datetime
                oracle.last_heartbeat = datetime.utcnow()
                self.db.commit()
                self.db.refresh(oracle)
                return oracle
            return None
        except Exception as e:
            logger.error(f"Error updating heartbeat for oracle {node_id}: {e}")
            raise