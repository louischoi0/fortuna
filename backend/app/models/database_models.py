from sqlalchemy import Column, Integer, String, DateTime, Boolean, Text, JSON, ForeignKey, Index
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database.connection import Base

class User(Base):
    """User model"""
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), unique=True, index=True, nullable=False)
    email = Column(String(100), unique=True, index=True, nullable=False)
    hashed_password = Column(String(255), nullable=False)
    is_active = Column(Boolean, default=True)
    is_superuser = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    # Relationships
    spaces = relationship("Space", back_populates="owner")
    transactions = relationship("Transaction", back_populates="user")

class Space(Base):
    """Space model - represents a Fortuna space"""
    __tablename__ = "spaces"
    
    id = Column(Integer, primary_key=True, index=True)
    space_id = Column(String(64), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    description = Column(Text)
    owner_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    is_active = Column(Boolean, default=True)
    config = Column(JSON)  # Space configuration as JSON
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    # Relationships
    owner = relationship("User", back_populates="spaces")
    transactions = relationship("Transaction", back_populates="space")
    events = relationship("Event", back_populates="space")
    
    # Indexes
    __table_args__ = (
        Index('idx_space_owner', 'owner_id'),
        Index('idx_space_active', 'is_active'),
    )

class Transaction(Base):
    """Transaction model"""
    __tablename__ = "transactions"
    
    id = Column(Integer, primary_key=True, index=True)
    transaction_id = Column(String(64), unique=True, index=True, nullable=False)
    space_id = Column(Integer, ForeignKey("spaces.id"), nullable=False)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    operation_type = Column(String(50), nullable=False)  # create, update, delete, etc.
    status = Column(String(20), default="pending")  # pending, completed, failed
    data = Column(JSON)  # Transaction data as JSON
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    completed_at = Column(DateTime(timezone=True))
    
    # Relationships
    space = relationship("Space", back_populates="transactions")
    user = relationship("User", back_populates="transactions")
    events = relationship("Event", back_populates="transaction")
    
    # Indexes
    __table_args__ = (
        Index('idx_transaction_space', 'space_id'),
        Index('idx_transaction_user', 'user_id'),
        Index('idx_transaction_status', 'status'),
        Index('idx_transaction_created', 'created_at'),
    )

class Event(Base):
    """Event model - represents Fortuna events"""
    __tablename__ = "events"
    
    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(String(64), unique=True, index=True, nullable=False)
    space_id = Column(Integer, ForeignKey("spaces.id"), nullable=False)
    transaction_id = Column(Integer, ForeignKey("transactions.id"), nullable=True)
    event_type = Column(String(50), nullable=False)
    event_data = Column(JSON)  # Event data as JSON
    sequence_number = Column(Integer, nullable=False)
    timestamp = Column(DateTime(timezone=True), server_default=func.now())
    
    # Relationships
    space = relationship("Space", back_populates="events")
    transaction = relationship("Transaction", back_populates="events")
    
    # Indexes
    __table_args__ = (
        Index('idx_event_space', 'space_id'),
        Index('idx_event_transaction', 'transaction_id'),
        Index('idx_event_type', 'event_type'),
        Index('idx_event_sequence', 'sequence_number'),
        Index('idx_event_timestamp', 'timestamp'),
    )

class Replica(Base):
    """Replica model - represents Fortuna replicas"""
    __tablename__ = "replicas"
    
    id = Column(Integer, primary_key=True, index=True)
    replica_id = Column(String(64), unique=True, index=True, nullable=False)
    space_id = Column(Integer, ForeignKey("spaces.id"), nullable=False)
    ip_address = Column(String(45), nullable=False)  # IPv4 or IPv6
    port = Column(Integer, nullable=False)
    is_active = Column(Boolean, default=True)
    last_heartbeat = Column(DateTime(timezone=True), server_default=func.now())
    metadata = Column(JSON)  # Replica metadata as JSON
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    # Relationships
    space = relationship("Space")
    
    # Indexes
    __table_args__ = (
        Index('idx_replica_space', 'space_id'),
        Index('idx_replica_active', 'is_active'),
        Index('idx_replica_ip_port', 'ip_address', 'port'),
    )

class OracleNode(Base):
    """Oracle node model"""
    __tablename__ = "oracle_nodes"
    
    id = Column(Integer, primary_key=True, index=True)
    node_id = Column(String(64), unique=True, index=True, nullable=False)
    ip_address = Column(String(45), nullable=False)
    port = Column(Integer, nullable=False)
    is_active = Column(Boolean, default=True)
    last_heartbeat = Column(DateTime(timezone=True), server_default=func.now())
    config = Column(JSON)  # Node configuration as JSON
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    # Indexes
    __table_args__ = (
        Index('idx_oracle_active', 'is_active'),
        Index('idx_oracle_ip_port', 'ip_address', 'port'),
    )
