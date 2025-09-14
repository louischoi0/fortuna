import pytest
from fastapi.testclient import TestClient
from main import app

# Create test client
client = TestClient(app)

def test_root_endpoint():
    """Test root endpoint"""
    response = client.get("/")
    assert response.status_code == 200
    assert "message" in response.json()
    assert response.json()["status"] == "healthy"

def test_health_endpoint():
    """Test health check endpoint"""
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"
    assert response.json()["service"] == "fortuna-backend"

def test_api_status_endpoint():
    """Test API status endpoint"""
    response = client.get("/api/v1/status")
    assert response.status_code == 200
    assert "status" in response.json()
    assert "version" in response.json()

def test_api_info_endpoint():
    """Test API info endpoint"""
    response = client.get("/api/v1/info")
    assert response.status_code == 200
    assert response.json()["name"] == "Fortuna Backend"
    assert response.json()["version"] == "1.0.0"
