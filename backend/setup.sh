#!/bin/bash

# Fortuna Backend Setup Script

echo "🚀 Setting up Fortuna Backend..."

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 is not installed. Please install Python 3.8 or higher."
    exit 1
fi

# Check Python version
python_version=$(python3 -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
echo "✅ Python version: $python_version"

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv venv
    echo "✅ Virtual environment created"
else
    echo "✅ Virtual environment already exists"
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Upgrade pip
echo "⬆️ Upgrading pip..."
pip install --upgrade pip

# Install dependencies
echo "📚 Installing dependencies..."
pip install -r requirements.txt

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "⚙️ Creating .env file..."
    cp env.example .env
    echo "✅ .env file created from env.example"
else
    echo "✅ .env file already exists"
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📊 Database Setup:"
echo "  1. Make sure PostgreSQL is running"
echo "  2. Create database: createdb -U fortuna fortuna_db"
echo "  3. Initialize tables: python init_database.py"
echo ""
echo "🚀 To start the development server:"
echo "  source venv/bin/activate"
echo "  python main.py"
echo ""
echo "Or use the run script:"
echo "  python run.py"
echo ""
echo "📚 API documentation will be available at:"
echo "  http://localhost:8000/docs"
echo ""
echo "🔧 Database management:"
echo "  - Initialize: python init_database.py"
echo "  - Check connection: python -c \"from app.database.connection import check_db_connection; print('Connected' if check_db_connection() else 'Failed')\""
