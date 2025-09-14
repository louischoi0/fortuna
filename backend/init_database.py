#!/usr/bin/env python3
"""
Database initialization script for Fortuna Backend
"""
import sys
import os

# Add the current directory to Python path
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from app.database.init_db import init_database, check_database_connection
from app.config import settings

def main():
    """Main function to initialize database"""
    print("🚀 Initializing Fortuna Backend Database...")
    print(f"📊 Database URL: {settings.database_url}")
    
    # Check database connection
    if not check_database_connection():
        print("❌ Database connection failed!")
        print("Please ensure PostgreSQL is running and the database exists.")
        print("You can create the database with:")
        print(f"  createdb -U {settings.database_user} {settings.database_name}")
        sys.exit(1)
    
    # Initialize database tables
    try:
        init_database()
        print("✅ Database initialized successfully!")
        print("🎉 You can now start the FastAPI server with: python main.py")
    except Exception as e:
        print(f"❌ Error initializing database: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
