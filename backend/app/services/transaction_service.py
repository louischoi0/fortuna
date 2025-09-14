from sqlalchemy.orm import Session
from typing import Optional, List
from app.models.database_models import Transaction
from app.models.schemas import TransactionCreate, TransactionUpdate
from app.services.base import BaseService
import uuid
import logging

logger = logging.getLogger(__name__)

class TransactionService(BaseService[Transaction]):
    """Transaction service for database operations"""
    
    def __init__(self, db: Session):
        super().__init__(Transaction, db)
    
    def create_transaction(self, transaction: TransactionCreate, user_id: int) -> Transaction:
        """Create a new transaction"""
        transaction_id = str(uuid.uuid4()).replace('-', '')
        transaction_data = {
            "transaction_id": transaction_id,
            "space_id": transaction.space_id,
            "user_id": user_id,
            "operation_type": transaction.operation_type,
            "data": transaction.data,
            "status": "pending"
        }
        return self.create(transaction_data)
    
    def get_by_transaction_id(self, transaction_id: str) -> Optional[Transaction]:
        """Get transaction by transaction_id"""
        try:
            return self.db.query(Transaction).filter(
                Transaction.transaction_id == transaction_id
            ).first()
        except Exception as e:
            logger.error(f"Error getting transaction by transaction_id {transaction_id}: {e}")
            raise
    
    def get_by_space(self, space_id: int, skip: int = 0, limit: int = 100) -> List[Transaction]:
        """Get transactions by space"""
        try:
            return self.db.query(Transaction).filter(
                Transaction.space_id == space_id
            ).offset(skip).limit(limit).all()
        except Exception as e:
            logger.error(f"Error getting transactions by space {space_id}: {e}")
            raise
    
    def get_by_user(self, user_id: int, skip: int = 0, limit: int = 100) -> List[Transaction]:
        """Get transactions by user"""
        try:
            return self.db.query(Transaction).filter(
                Transaction.user_id == user_id
            ).offset(skip).limit(limit).all()
        except Exception as e:
            logger.error(f"Error getting transactions by user {user_id}: {e}")
            raise
    
    def get_by_status(self, status: str, skip: int = 0, limit: int = 100) -> List[Transaction]:
        """Get transactions by status"""
        try:
            return self.db.query(Transaction).filter(
                Transaction.status == status
            ).offset(skip).limit(limit).all()
        except Exception as e:
            logger.error(f"Error getting transactions by status {status}: {e}")
            raise
    
    def update_transaction(self, transaction_id: int, transaction_update: TransactionUpdate) -> Optional[Transaction]:
        """Update transaction information"""
        transaction = self.get(transaction_id)
        if not transaction:
            return None
        
        update_data = {}
        if transaction_update.status is not None:
            update_data["status"] = transaction_update.status
        if transaction_update.data is not None:
            update_data["data"] = transaction_update.data
        
        if update_data:
            return self.update(transaction, update_data)
        return transaction
    
    def complete_transaction(self, transaction_id: int) -> Optional[Transaction]:
        """Mark transaction as completed"""
        from datetime import datetime
        transaction = self.get(transaction_id)
        if not transaction:
            return None
        
        update_data = {
            "status": "completed",
            "completed_at": datetime.utcnow()
        }
        return self.update(transaction, update_data)
    
    def fail_transaction(self, transaction_id: int) -> Optional[Transaction]:
        """Mark transaction as failed"""
        transaction = self.get(transaction_id)
        if not transaction:
            return None
        
        return self.update(transaction, {"status": "failed"})
