from pydantic import BaseModel, EmailStr
from typing import Optional, Dict, Any, List
from datetime import datetime

# API Response Models
class HealthResponse(BaseModel):
    """Health check response model"""
    status: str
    service: str
    timestamp: Optional[datetime] = None

class APIStatusResponse(BaseModel):
    """API status response model"""
    status: str
    service: str
    environment: str
    version: str

class APIInfoResponse(BaseModel):
    """API info response model"""
    name: str
    version: str
    description: str
    endpoints: Dict[str, str]

class ErrorResponse(BaseModel):
    """Error response model"""
    error: str
    message: str
    status_code: int

# Database Models (Pydantic schemas)
class UserBase(BaseModel):
    """Base user schema"""
    username: str
    email: EmailStr
    is_active: bool = True

class UserCreate(UserBase):
    """User creation schema"""
    password: str

class UserUpdate(BaseModel):
    """User update schema"""
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    is_active: Optional[bool] = None

class User(UserBase):
    """User response schema"""
    id: int
    is_superuser: bool
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class SpaceBase(BaseModel):
    """Base space schema"""
    name: str
    description: Optional[str] = None
    config: Optional[Dict[str, Any]] = None

class SpaceCreate(SpaceBase):
    """Space creation schema"""
    pass

class SpaceUpdate(BaseModel):
    """Space update schema"""
    name: Optional[str] = None
    description: Optional[str] = None
    config: Optional[Dict[str, Any]] = None
    is_active: Optional[bool] = None

class Space(SpaceBase):
    """Space response schema"""
    id: int
    space_id: str
    owner_id: int
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class TransactionBase(BaseModel):
    """Base transaction schema"""
    operation_type: str
    data: Optional[Dict[str, Any]] = None

class TransactionCreate(TransactionBase):
    """Transaction creation schema"""
    space_id: int

class TransactionUpdate(BaseModel):
    """Transaction update schema"""
    status: Optional[str] = None
    data: Optional[Dict[str, Any]] = None

class Transaction(TransactionBase):
    """Transaction response schema"""
    id: int
    transaction_id: str
    space_id: int
    user_id: int
    status: str
    created_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class EventBase(BaseModel):
    """Base event schema"""
    event_type: str
    event_data: Optional[Dict[str, Any]] = None
    sequence_number: int

class EventCreate(EventBase):
    """Event creation schema"""
    space_id: int
    transaction_id: Optional[int] = None

class Event(EventBase):
    """Event response schema"""
    id: int
    event_id: str
    space_id: int
    transaction_id: Optional[int] = None
    timestamp: datetime

    class Config:
        from_attributes = True

class ReplicaBase(BaseModel):
    """Base replica schema"""
    ip_address: str
    port: int
    metadata: Optional[Dict[str, Any]] = None

class ReplicaCreate(ReplicaBase):
    """Replica creation schema"""
    space_id: int

class ReplicaUpdate(BaseModel):
    """Replica update schema"""
    is_active: Optional[bool] = None
    metadata: Optional[Dict[str, Any]] = None

class Replica(ReplicaBase):
    """Replica response schema"""
    id: int
    replica_id: str
    space_id: int
    is_active: bool
    last_heartbeat: datetime
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class OracleNodeBase(BaseModel):
    """Base oracle node schema"""
    ip_address: str
    port: int
    config: Optional[Dict[str, Any]] = None

class OracleNodeCreate(OracleNodeBase):
    """Oracle node creation schema"""
    pass

class OracleNodeUpdate(BaseModel):
    """Oracle node update schema"""
    is_active: Optional[bool] = None
    config: Optional[Dict[str, Any]] = None

class OracleNode(OracleNodeBase):
    """Oracle node response schema"""
    id: int
    node_id: str
    is_active: bool
    last_heartbeat: datetime
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class OracleStatusResponse(BaseModel):
    """Oracle status response schema"""
    node_id: str
    ip_address: str
    port: int
    is_active: bool
    last_heartbeat: Optional[str] = None
    status: str  # online, offline, error
    sdk_response: Optional[str] = None
    error: Optional[str] = None

class OracleStatusListResponse(BaseModel):
    """Oracle status list response schema"""
    message: str
    total_oracles: int
    oracles: List[OracleStatusResponse]
