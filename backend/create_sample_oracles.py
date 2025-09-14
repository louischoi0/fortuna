#!/usr/bin/env python3
"""
Create sample oracle nodes for testing
"""
import sys
import os

# Add the current directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from sqlalchemy.orm import Session
from app.database.connection import SessionLocal, init_db
from app.models.database_models import OracleNode
from datetime import datetime
import uuid

def create_sample_oracles():
    """Create sample oracle nodes in the database"""
    
    # Initialize database
    init_db()
    
    # Create database session
    db = SessionLocal()
    
    try:
        # Check if oracles already exist
        existing_oracles = db.query(OracleNode).count()
        if existing_oracles > 0:
            print(f"✅ {existing_oracles} oracle(s) already exist in database")
            return
        
        # Sample oracle nodes
        sample_oracles = [
            {
                "node_id": str(uuid.uuid4()).replace('-', ''),
                "ip_address": "127.0.0.1",
                "port": 8080,
                "is_active": True,
                "config": {"region": "local", "priority": 1}
            },
            {
                "node_id": str(uuid.uuid4()).replace('-', ''),
                "ip_address": "127.0.0.1", 
                "port": 8081,
                "is_active": True,
                "config": {"region": "local", "priority": 2}
            },
            {
                "node_id": str(uuid.uuid4()).replace('-', ''),
                "ip_address": "192.168.1.100",
                "port": 8080,
                "is_active": True,
                "config": {"region": "remote", "priority": 3}
            }
        ]
        
        # Create oracle nodes
        for oracle_data in sample_oracles:
            oracle = OracleNode(**oracle_data)
            db.add(oracle)
        
        # Commit to database
        db.commit()
        
        print("✅ Sample oracle nodes created successfully!")
        print(f"📊 Created {len(sample_oracles)} oracle nodes:")
        
        for oracle in sample_oracles:
            print(f"  - {oracle['node_id']}: {oracle['ip_address']}:{oracle['port']}")
        
        print("\n🚀 You can now test the oracle status API:")
        print("  GET http://localhost:8000/api/v1/oracles/status")
        
    except Exception as e:
        print(f"❌ Error creating sample oracles: {e}")
        db.rollback()
        raise
    finally:
        db.close()

if __name__ == "__main__":
    print("🔧 Creating sample oracle nodes...")
    create_sample_oracles()
