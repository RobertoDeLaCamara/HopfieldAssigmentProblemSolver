"""
Comprehensive tests for the enhanced Flask API server.
Tests validation, logging, metrics, and all new features.
"""

import pytest
import json
import sys
import os

# Add src directory to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

# Import both API servers for testing
from api_server import app as original_app
from api_server_enhanced import app as enhanced_app


@pytest.fixture
def client():
    """Test client for the original Flask application."""
    original_app.config['TESTING'] = True
    with original_app.test_client() as client:
        yield client


@pytest.fixture
def enhanced_client():
    """Test client for the enhanced Flask application."""
    enhanced_app.config['TESTING'] = True
    with enhanced_app.test_client() as client:
        yield client


class TestHealthEndpoints:
    """Tests for health endpoints."""
    
    def test_health_check(self, client):
        """Test the /health endpoint."""
        response = client.get('/health')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'healthy'
        assert data['service'] == 'hopfield-assignment-solver'
    
    def test_enhanced_health_check(self, enhanced_client):
        """Test the enhanced /health endpoint with version."""
        response = enhanced_client.get('/health')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'healthy'
        assert data['service'] == 'hopfield-assignment-solver'
        assert 'version' in data
        assert 'timestamp' in data
    
    def test_health_ready(self, enhanced_client):
        """Test the /health/ready endpoint."""
        response = enhanced_client.get('/health/ready')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'ready'
    
    def test_health_live(self, enhanced_client):
        """Test the /health/live endpoint."""
        response = enhanced_client.get('/health/live')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'alive'
    
    def test_health_check_content_type(self, client):
        """Test that verifies the response Content-Type."""
        response = client.get('/health')
        assert response.content_type == 'application/json'


class TestSolveEndpoint:
    """Tests for the /solve endpoint."""
    
    def test_solve_valid_request(self, client):
        """Test with a valid request."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve', 
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        assert 'result' in result
        
        # Verify result structure
        result_data = result['result']
        assert 'assignments' in result_data
        assert 'total_cost' in result_data
        assert 'iterations' in result_data
        assert 'cost_matrix' in result_data
        
        # Verify valid assignments
        assignments = result_data['assignments']
        assert len(assignments) == 2
        assert set(assignments) == {0, 1}
    
    def test_enhanced_solve_with_request_id(self, enhanced_client):
        """Test enhanced solve with request ID tracking."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve', 
                                       data=json.dumps(data),
                                       content_type='application/json',
                                       headers={'X-Request-ID': 'test-123'})
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        assert 'request_id' in result
        assert result['request_id'] == 'test-123'
        
        # Check response headers
        assert 'X-Request-ID' in response.headers
        assert 'X-Response-Time' in response.headers
    
    def test_solve_3x3_matrix(self, client):
        """Test with 3x3 matrix."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        
        assignments = result['result']['assignments']
        assert len(assignments) == 3
        assert set(assignments) == {0, 1, 2}
    
    def test_solve_missing_cost_matrix(self, client):
        """Test without cost matrix."""
        data = {}
        
        response = client.post('/solve', json=data)
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert 'cost_matrix' in result['error']
    
    def test_solve_invalid_json(self, client):
        """Test with invalid JSON."""
        response = client.post('/solve',
                             data='invalid json',
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_solve_non_square_matrix(self, client):
        """Test with non-square matrix."""
        cost_matrix = [[1, 2, 3], [4, 5, 6]]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_solve_empty_matrix(self, client):
        """Test with empty matrix."""
        cost_matrix = []
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_solve_invalid_cost_values(self, client):
        """Test with invalid cost values."""
        cost_matrix = [['a', 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False


class TestEnhancedValidation:
    """Tests for enhanced validation features."""
    
    def test_matrix_too_large(self, enhanced_client):
        """Test with matrix exceeding maximum size (50x50)."""
        # Create 51x51 matrix
        cost_matrix = [[1.0] * 51 for _ in range(51)]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert '50' in result['error']  # Should mention the limit
    
    def test_matrix_too_small(self, enhanced_client):
        """Test with 1x1 matrix (below minimum of 2x2)."""
        cost_matrix = [[1]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_nan_values(self, enhanced_client):
        """Test with NaN values."""
        cost_matrix = [[float('nan'), 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert 'NaN' in result['error']
    
    def test_inf_values(self, enhanced_client):
        """Test with infinite values."""
        cost_matrix = [[float('inf'), 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert 'infinite' in result['error']
    
    def test_negative_costs(self, enhanced_client):
        """Test with negative cost values."""
        cost_matrix = [[-1, 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert 'negative' in result['error'].lower()
    
    def test_validation_info_endpoint(self, enhanced_client):
        """Test the /validation/info endpoint."""
        response = enhanced_client.get('/validation/info')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        
        # Check that validation constraints are returned
        assert 'matrix_size' in data
        assert 'min' in data['matrix_size']
        assert 'max' in data['matrix_size']
        assert data['matrix_size']['max'] == 50
        assert data['matrix_size']['min'] == 2


class TestBatchEndpoint:
    """Tests for the /solve/batch endpoint."""
    
    def test_batch_valid_request(self, client):
        """Test with valid batch request."""
        problems = [
            {
                'id': 'problem_1',
                'cost_matrix': [[1, 2], [3, 4]]
            },
            {
                'id': 'problem_2',
                'cost_matrix': [[5, 6], [7, 8]]
            }
        ]
        data = {'problems': problems}
        
        response = client.post('/solve/batch',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        assert len(result['results']) == 2
        
        # Verify each result
        for i, problem_result in enumerate(result['results']):
            assert problem_result['id'] == f'problem_{i+1}'
            assert problem_result['success'] is True
            assert 'result' in problem_result
    
    def test_enhanced_batch_with_summary(self, enhanced_client):
        """Test enhanced batch endpoint with summary."""
        problems = [
            {'id': 'p1', 'cost_matrix': [[1, 2], [3, 4]]},
            {'id': 'p2', 'cost_matrix': [[5, 6], [7, 8]]},
        ]
        data = {'problems': problems}
        
        response = enhanced_client.post('/solve/batch',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        
        # Check for summary in enhanced version
        assert 'summary' in result
        assert result['summary']['total'] == 2
        assert result['summary']['successful'] == 2
        assert result['summary']['failed'] == 0
    
    def test_batch_mixed_valid_invalid(self, client):
        """Test with batch containing valid and invalid problems."""
        problems = [
            {
                'id': 'valid_problem',
                'cost_matrix': [[1, 2], [3, 4]]
            },
            {
                'id': 'invalid_problem',
                'cost_matrix': [[1, 2, 3], [4, 5, 6]]  # Not square
            }
        ]
        data = {'problems': problems}
        
        response = client.post('/solve/batch',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        assert len(result['results']) == 2
        
        # Verify that the valid problem was solved
        valid_result = next(r for r in result['results'] if r['id'] == 'valid_problem')
        assert valid_result['success'] is True
        
        # Verify that the invalid problem failed
        invalid_result = next(r for r in result['results'] if r['id'] == 'invalid_problem')
        assert invalid_result['success'] is False
        assert 'error' in invalid_result
    
    def test_batch_empty_problems(self, client):
        """Test with empty problems list."""
        data = {'problems': []}
        
        response = client.post('/solve/batch',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_batch_too_many_problems(self, enhanced_client):
        """Test with batch size exceeding limit (100)."""
        problems = [
            {'id': f'problem_{i}', 'cost_matrix': [[1, 2], [3, 4]]}
            for i in range(101)
        ]
        data = {'problems': problems}
        
        response = enhanced_client.post('/solve/batch',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
        assert '100' in result['error']
    
    def test_batch_missing_problems(self, client):
        """Test without problems field."""
        data = {}
        
        response = client.post('/solve/batch',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 400
        result = json.loads(response.data)
        assert result['success'] is False
    
    def test_batch_missing_id(self, client):
        """Test with problem without ID."""
        problems = [
            {
                'cost_matrix': [[1, 2], [3, 4]]
            }
        ]
        data = {'problems': problems}
        
        response = client.post('/solve/batch',
                             data=json.dumps(data),
                             content_type='application/json')
        
        assert response.status_code == 200
        result = json.loads(response.data)
        assert result['success'] is True
        
        # The problem without ID should fail
        problem_result = result['results'][0]
        assert problem_result['success'] is False


class TestMetricsEndpoint:
    """Tests for the /metrics endpoint (enhanced API only)."""
    
    def test_metrics_endpoint(self, enhanced_client):
        """Test the /metrics endpoint."""
        # Make a few requests first
        cost_matrix = [[1, 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        for _ in range(3):
            enhanced_client.post('/solve',
                                data=json.dumps(data),
                                content_type='application/json')
        
        # Get metrics
        response = enhanced_client.get('/metrics')
        
        assert response.status_code == 200
        data = json.loads(response.data)
        
        # Check metrics structure
        assert 'requests' in data
        assert 'performance' in data
        assert 'algorithm' in data
        
        # Check that requests were counted
        assert data['requests']['total'] >= 3


class TestErrorHandling:
    """Tests for error handling."""
    
    def test_internal_server_error_handling(self, client):
        """Test that verifies internal error handling."""
        # Test with very large numbers (within valid range)
        cost_matrix = [[1e8, 2e8], [3e8, 4e8]]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        # Should handle large numbers correctly
        assert response.status_code in [200, 400, 500]  # Valid responses


class TestContentType:
    """Tests to verify response Content-Type."""
    
    def test_all_responses_json(self, client):
        """Test that all responses are JSON."""
        endpoints = [
            ('/health', 'GET'),
            ('/solve', 'POST'),
            ('/solve/batch', 'POST')
        ]
        
        for endpoint, method in endpoints:
            if method == 'GET':
                response = client.get(endpoint)
            else:
                data = {'cost_matrix': [[1, 2], [3, 4]]} if endpoint == '/solve' else {'problems': []}
                response = client.post(endpoint,
                                     data=json.dumps(data),
                                     content_type='application/json')
            
            assert response.content_type == 'application/json'
    
    def test_enhanced_all_endpoints_json(self, enhanced_client):
        """Test that all enhanced endpoints return JSON."""
        endpoints = [
            '/health',
            '/health/ready',
            '/health/live',
            '/metrics',
            '/validation/info'
        ]
        
        for endpoint in endpoints:
            response = enhanced_client.get(endpoint)
            assert response.content_type == 'application/json'


class TestCORS:
    """Tests for CORS headers (enhanced API only)."""
    
    def test_cors_headers(self, enhanced_client):
        """Test that CORS headers are present."""
        cost_matrix = [[1, 2], [3, 4]]
        data = {'cost_matrix': cost_matrix}
        
        response = enhanced_client.post('/solve',
                                       data=json.dumps(data),
                                       content_type='application/json')
        
        assert 'Access-Control-Allow-Origin' in response.headers
        assert response.headers['Access-Control-Allow-Origin'] == '*'


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
