"""
Tests for the validation module.
"""

import pytest
import sys
import os

# Add src directory to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from validation import (
    validate_cost_matrix,
    validate_batch_request,
    ValidationError,
    get_validation_summary
)


class TestValidateCostMatrix:
    """Tests for cost matrix validation."""
    
    def test_valid_2x2_matrix(self):
        """Test with valid 2x2 matrix."""
        cost_matrix = [[1, 2], [3, 4]]
        is_valid, error = validate_cost_matrix(cost_matrix)
        assert is_valid is True
        assert error is None
    
    def test_valid_5x5_matrix(self):
        """Test with valid 5x5 matrix."""
        cost_matrix = [[i*5+j for j in range(5)] for i in range(5)]
        is_valid, error = validate_cost_matrix(cost_matrix)
        assert is_valid is True
    
    def test_empty_matrix(self):
        """Test with empty matrix."""
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix([])
        assert 'empty' in str(exc_info.value).lower()
    
    def test_matrix_too_large(self):
        """Test with matrix exceeding maximum size."""
        cost_matrix = [[1] * 51 for _ in range(51)]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert '50' in str(exc_info.value)
    
    def test_matrix_too_small(self):
        """Test with 1x1 matrix."""
        cost_matrix = [[1]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'too small' in str(exc_info.value).lower()
    
    def test_non_square_matrix(self):
        """Test with non-square matrix."""
        cost_matrix = [[1, 2, 3], [4, 5, 6]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'square' in str(exc_info.value).lower()
    
    def test_nan_value(self):
        """Test with NaN value."""
        cost_matrix = [[float('nan'), 2], [3, 4]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'NaN' in str(exc_info.value)
    
    def test_inf_value(self):
        """Test with infinite value."""
        cost_matrix = [[float('inf'), 2], [3, 4]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'infinite' in str(exc_info.value).lower()
    
    def test_negative_cost(self):
        """Test with negative cost."""
        cost_matrix = [[-1, 2], [3, 4]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'negative' in str(exc_info.value).lower()
    
    def test_cost_too_large(self):
        """Test with cost exceeding maximum value."""
        cost_matrix = [[1e10, 2], [3, 4]]  # Exceeds MAX_COST_VALUE (1e9)
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'exceeds' in str(exc_info.value).lower()
    
    def test_invalid_row_type(self):
        """Test with invalid row type."""
        cost_matrix = [[1, 2], "invalid"]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'must be a list' in str(exc_info.value).lower()
    
    def test_invalid_cost_type(self):
        """Test with invalid cost value type."""
        cost_matrix = [['a', 2], [3, 4]]
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix(cost_matrix)
        assert 'must be a number' in str(exc_info.value).lower()
    
    def test_non_list_matrix(self):
        """Test with non-list matrix."""
        with pytest.raises(ValidationError) as exc_info:
            validate_cost_matrix("not a list")
        assert 'must be a list' in str(exc_info.value).lower()


class TestValidateBatchRequest:
    """Tests for batch request validation."""
    
    def test_valid_batch(self):
        """Test with valid batch request."""
        problems = [
            {'id': 'p1', 'cost_matrix': [[1, 2], [3, 4]]},
            {'id': 'p2', 'cost_matrix': [[5, 6], [7, 8]]}
        ]
        is_valid, error = validate_batch_request(problems)
        assert is_valid is True
    
    def test_empty_batch(self):
        """Test with empty problems list."""
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request([])
        assert 'empty' in str(exc_info.value).lower()
    
    def test_batch_too_large(self):
        """Test with batch exceeding maximum size."""
        problems = [
            {'id': f'p{i}', 'cost_matrix': [[1, 2], [3, 4]]}
            for i in range(101)
        ]
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request(problems)
        assert '100' in str(exc_info.value)
    
    def test_non_list_batch(self):
        """Test with non-list batch."""
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request("not a list")
        assert 'must be a list' in str(exc_info.value).lower()
    
    def test_missing_id(self):
        """Test with problem missing ID."""
        problems = [
            {'cost_matrix': [[1, 2], [3, 4]]}
        ]
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request(problems)
        assert 'id' in str(exc_info.value).lower()
    
    def test_missing_cost_matrix(self):
        """Test with problem missing cost_matrix."""
        problems = [
            {'id': 'p1'}
        ]
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request(problems)
        assert 'cost_matrix' in str(exc_info.value).lower()
    
    def test_invalid_problem_type(self):
        """Test with non-dict problem."""
        problems = ['not a dict']
        with pytest.raises(ValidationError) as exc_info:
            validate_batch_request(problems)
        assert 'dictionary' in str(exc_info.value).lower()


class TestGetValidationSummary:
    """Tests for get_validation_summary function."""
    
    def test_validation_summary_structure(self):
        """Test that validation summary has correct structure."""
        summary = get_validation_summary()
        
        assert 'matrix_size' in summary
        assert 'cost_value' in summary
        assert 'batch_size' in summary
        assert 'request_size_mb' in summary
    
    def test_matrix_size_constraints(self):
        """Test matrix size constraints in summary."""
        summary = get_validation_summary()
        
        assert summary['matrix_size']['min'] == 2
        assert summary['matrix_size']['max'] == 50
    
    def test_cost_value_constraints(self):
        """Test cost value constraints in summary."""
        summary = get_validation_summary()
        
        assert summary['cost_value']['min'] == 0
        assert summary['cost_value']['max'] == 1e9
    
    def test_batch_size_constraint(self):
        """Test batch size constraint in summary."""
        summary = get_validation_summary()
        
        assert summary['batch_size']['max'] == 100


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
