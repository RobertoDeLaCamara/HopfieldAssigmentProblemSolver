"""
Integration tests for the Flask API server.
"""

import pytest
import json
from src.api_server import app


@pytest.fixture
def client():
    """Cliente de prueba para la aplicación Flask."""
    app.config['TESTING'] = True
    with app.test_client() as client:
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
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
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
    
    def test_batch_mixed_valid_invalid(self, client):
        """Test with batch containing valid and invalid problems."""
        problems = [
            {
                'id': 'valid_problem',
                'cost_matrix': [[1, 2], [3, 4]]
            },
            {
                'id': 'invalid_problem',
                'cost_matrix': [[1, 2, 3], [4, 5, 6]]  # No cuadrada
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


class TestErrorHandling:
    """Tests for error handling."""
    
    def test_internal_server_error_handling(self, client):
        """Test that verifies internal error handling."""
        # Simulate an internal error by temporarily modifying the solver
        # Este test es más conceptual ya que es difícil simular errores internos
        # sin modificar el código de producción
        
        # Test with data that could cause problems
        cost_matrix = [[1e10, 2e10], [3e10, 4e10]]
        data = {'cost_matrix': cost_matrix}
        
        response = client.post('/solve',
                             data=json.dumps(data),
                             content_type='application/json')
        
        # Should handle large numbers correctly
        assert response.status_code in [200, 500]  # Puede ser exitoso o fallar graciosamente


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
                data = {'cost_matrix': [[1, 2], [3, 4]]} if 'solve' in endpoint else {'problems': []}
                response = client.post(endpoint,
                                     data=json.dumps(data),
                                     content_type='application/json')
            
            assert response.content_type == 'application/json'


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
