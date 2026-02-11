"""
Integration tests for the complete system.
"""

import os

import pytest
import requests
import time
import json
from typing import Dict, Any


class TestSystemIntegration:
    """Integration tests for the complete system."""

    @pytest.fixture(autouse=True)
    def setup(self):
        """Setup before each test."""
        self.api_base_url = os.environ.get("API_BASE_URL", "http://localhost:8080")
        self.hopfield_base_url = os.environ.get("HOPFIELD_BASE_URL", "http://localhost:5000")
        self.timeout = 30
        
        # Wait for services to be ready
        self._wait_for_services()
    
    def _wait_for_services(self):
        """Wait for services to be ready."""
        max_attempts = 30
        attempt = 0
        
        while attempt < max_attempts:
            try:
                # Check API Gateway
                response = requests.get(f"{self.api_base_url}/health", timeout=5)
                if response.status_code == 200:
                    # Check Hopfield service
                    response = requests.get(f"{self.hopfield_base_url}/health", timeout=5)
                    if response.status_code == 200:
                        return
            except requests.exceptions.RequestException:
                pass
            
            attempt += 1
            time.sleep(1)
        
        pytest.skip("Services are not available")
    
    def test_api_gateway_health(self):
        """Test API Gateway health."""
        response = requests.get(f"{self.api_base_url}/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "hopfield-assignment-api"
    
    def test_hopfield_service_health(self):
        """Test Hopfield service health."""
        response = requests.get(f"{self.hopfield_base_url}/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "hopfield-assignment-solver"
    
    def test_end_to_end_solve_simple(self):
        """End-to-end test with simple problem."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {"cost_matrix": cost_matrix}
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        # Verify result structure
        assert "result" in result
        result_data = result["result"]
        assert "assignments" in result_data
        assert "total_cost" in result_data
        assert "iterations" in result_data
        assert "cost_matrix" in result_data
        
        # Verify valid assignments
        assignments = result_data["assignments"]
        assert len(assignments) == 2
        assert set(assignments) == {0, 1}
        assert result_data["total_cost"] > 0
        assert result_data["iterations"] > 0
    
    def test_end_to_end_solve_3x3(self):
        """End-to-end test with 3x3 matrix."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ]
        data = {"cost_matrix": cost_matrix}
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        assignments = result["result"]["assignments"]
        assert len(assignments) == 3
        assert set(assignments) == {0, 1, 2}
    
    def test_end_to_end_solve_4x4(self):
        """End-to-end test with 4x4 matrix."""
        cost_matrix = [
            [9, 2, 7, 8],
            [6, 4, 3, 7],
            [5, 8, 1, 8],
            [7, 6, 9, 4]
        ]
        data = {"cost_matrix": cost_matrix}
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        assignments = result["result"]["assignments"]
        assert len(assignments) == 4
        assert set(assignments) == {0, 1, 2, 3}
    
    def test_batch_processing(self):
        """Test batch processing."""
        problems = [
            {
                "id": "problem_1",
                "cost_matrix": [[1, 2], [3, 4]]
            },
            {
                "id": "problem_2",
                "cost_matrix": [[5, 6], [7, 8]]
            },
            {
                "id": "problem_3",
                "cost_matrix": [[9, 10, 11], [12, 13, 14], [15, 16, 17]]
            }
        ]
        data = {"problems": problems}
        
        response = requests.post(
            f"{self.api_base_url}/solve/batch",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        assert len(result["results"]) == 3
        
        # Verify each result
        for i, problem_result in enumerate(result["results"]):
            assert problem_result["id"] == f"problem_{i+1}"
            assert problem_result["success"] is True
            assert "result" in problem_result
            
            # Verify valid assignments
            assignments = problem_result["result"]["assignments"]
            expected_size = 2 if i < 2 else 3
            assert len(assignments) == expected_size
            assert set(assignments) == set(range(expected_size))
    
    def test_error_handling_invalid_matrix(self):
        """Test error handling with invalid matrix."""
        data = {"cost_matrix": [[1, 2, 3], [4, 5, 6]]}  # Not square
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 400
        result = response.json()
        assert result["success"] is False
        assert "error" in result
    
    def test_error_handling_missing_field(self):
        """Test error handling with missing field."""
        data = {"invalid_field": "test"}
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 400
        result = response.json()
        assert result["success"] is False
        assert "error" in result
    
    def test_direct_hopfield_service_access(self):
        """Test direct access to Hopfield service."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {"cost_matrix": cost_matrix}
        
        response = requests.post(
            f"{self.hopfield_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        assignments = result["result"]["assignments"]
        assert len(assignments) == 2
        assert set(assignments) == {0, 1}
    
    def test_performance_small_matrices(self):
        """Test performance with small matrices."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {"cost_matrix": cost_matrix}
        
        start_time = time.time()
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        end_time = time.time()
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        # Verify that the response was fast (less than 5 seconds)
        response_time = end_time - start_time
        assert response_time < 5.0
        
        # Verify that the algorithm converged in a reasonable number of iterations
        iterations = result["result"]["iterations"]
        assert iterations <= 1000  # Should converge within max iterations
    
    def test_concurrent_requests(self):
        """Test concurrent requests."""
        import concurrent.futures
        import threading
        
        def make_request():
            cost_matrix = [[1, 2], [3, 4]]
            data = {"cost_matrix": cost_matrix}
            response = requests.post(
                f"{self.api_base_url}/solve",
                json=data,
                timeout=self.timeout
            )
            return response.status_code == 200
        
        # Make 5 concurrent requests
        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(make_request) for _ in range(5)]
            results = [future.result() for future in concurrent.futures.as_completed(futures)]
        
        # All requests must be successful
        assert all(results)
    
    def test_large_matrix_handling(self):
        """Test with larger matrix."""
        # Create 5x5 matrix
        cost_matrix = []
        for i in range(5):
            row = []
            for j in range(5):
                row.append(i * 5 + j + 1)
            cost_matrix.append(row)
        
        data = {"cost_matrix": cost_matrix}
        
        response = requests.post(
            f"{self.api_base_url}/solve",
            json=data,
            timeout=self.timeout
        )
        
        assert response.status_code == 200
        result = response.json()
        assert result["success"] is True
        
        assignments = result["result"]["assignments"]
        assert len(assignments) == 5
        assert set(assignments) == {0, 1, 2, 3, 4}
    
    def test_nginx_proxy(self):
        """Test Nginx proxy (if available)."""
        try:
            # Try to access through Nginx proxy
            response = requests.get("http://localhost/health", timeout=5)
            if response.status_code == 200:
                data = response.json()
                assert data["status"] == "healthy"
        except requests.exceptions.RequestException:
            # Nginx is not available, this is normal in development
            pytest.skip("Nginx is not available")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
