"""
Unit tests for the Hopfield algorithm.
"""

import pytest
import numpy as np
from src.hopfield_solver import HopfieldAssignmentSolver, solve_assignment_problem


class TestHopfieldAssignmentSolver:
    """Tests for the HopfieldAssignmentSolver class."""
    
    def setup_method(self):
        """Setup before each test."""
        self.solver = HopfieldAssignmentSolver(max_iterations=100, convergence_threshold=1e-4)
    
    def test_solve_simple_2x2(self):
        """Test with a simple 2x2 matrix."""
        cost_matrix = [[1, 2], [3, 4]]
        assignments, total_cost, iterations = self.solver.solve(cost_matrix)
        
        # Verify that valid assignments were obtained
        assert len(assignments) == 2
        assert set(assignments) == {0, 1}  # Each worker must have a unique job
        assert total_cost > 0
        assert iterations > 0
        assert iterations <= 100
    
    def test_solve_3x3(self):
        """Test with a 3x3 matrix."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ]
        assignments, total_cost, iterations = self.solver.solve(cost_matrix)
        
        # Verify valid assignments
        assert len(assignments) == 3
        assert set(assignments) == {0, 1, 2}
        assert total_cost > 0
        assert iterations > 0
    
    def test_solve_4x4(self):
        """Test with a 4x4 matrix."""
        cost_matrix = [
            [9, 2, 7, 8],
            [6, 4, 3, 7],
            [5, 8, 1, 8],
            [7, 6, 9, 4]
        ]
        assignments, total_cost, iterations = self.solver.solve(cost_matrix)
        
        # Verify valid assignments
        assert len(assignments) == 4
        assert set(assignments) == {0, 1, 2, 3}
        assert total_cost > 0
        assert iterations > 0
    
    def test_invalid_matrix_not_square(self):
        """Test with non-square matrix."""
        cost_matrix = [[1, 2, 3], [4, 5, 6]]
        
        with pytest.raises(ValueError, match="The cost matrix must be square"):
            self.solver.solve(cost_matrix)
    
    def test_empty_matrix(self):
        """Test with empty matrix."""
        cost_matrix = []
        
        with pytest.raises(ValueError, match="The cost matrix must be square"):
            self.solver.solve(cost_matrix)
    
    def test_single_element_matrix(self):
        """Test with 1x1 matrix."""
        cost_matrix = [[5]]
        assignments, total_cost, iterations = self.solver.solve(cost_matrix)
        
        assert assignments == [0]
        assert total_cost == 5.0
        assert iterations > 0
    
    def test_discretize_assignments(self):
        """Test the discretization method."""
        # Create a simulated state matrix
        V = np.array([
            [0.1, 0.9, 0.2],
            [0.8, 0.1, 0.3],
            [0.2, 0.3, 0.7]
        ])
        
        assignments = self.solver._discretize_assignments(V)
        
        # Verify that assignments are valid
        assert len(assignments) == 3
        assert set(assignments) == {0, 1, 2}
    
    def test_calculate_total_cost(self):
        """Test total cost calculation."""
        assignments = [1, 0, 2]
        cost_matrix = np.array([
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ])
        
        total_cost = self.solver._calculate_total_cost(assignments, cost_matrix)
        
        # assignments[0]=1 -> cost_matrix[0,1] = 2
        # assignments[1]=0 -> cost_matrix[1,0] = 4
        # assignments[2]=2 -> cost_matrix[2,2] = 9
        expected_cost = 2 + 4 + 9
        assert total_cost == expected_cost


class TestSolveAssignmentProblem:
    """Tests for the solve_assignment_problem convenience function."""
    
    def test_solve_assignment_problem_simple(self):
        """Test the solve_assignment_problem function."""
        cost_matrix = [[1, 2], [3, 4]]
        result = solve_assignment_problem(cost_matrix)
        
        # Verify response structure
        assert 'assignments' in result
        assert 'total_cost' in result
        assert 'iterations' in result
        assert 'cost_matrix' in result
        
        # Verify types
        assert isinstance(result['assignments'], list)
        assert isinstance(result['total_cost'], float)
        assert isinstance(result['iterations'], int)
        assert isinstance(result['cost_matrix'], list)
        
        # Verify valid assignments
        assert len(result['assignments']) == 2
        assert set(result['assignments']) == {0, 1}
        assert result['total_cost'] > 0
        assert result['iterations'] > 0
    
    def test_solve_assignment_problem_large(self):
        """Test with larger matrix."""
        cost_matrix = [
            [1, 2, 3, 4],
            [5, 6, 7, 8],
            [9, 10, 11, 12],
            [13, 14, 15, 16]
        ]
        result = solve_assignment_problem(cost_matrix)
        
        assert len(result['assignments']) == 4
        assert set(result['assignments']) == {0, 1, 2, 3}
        assert result['total_cost'] > 0


class TestConvergence:
    """Tests to verify algorithm convergence."""
    
    def test_convergence_with_different_thresholds(self):
        """Test with different convergence thresholds."""
        cost_matrix = [[1, 2], [3, 4]]
        
        # Stricter threshold
        solver_strict = HopfieldAssignmentSolver(convergence_threshold=1e-8)
        assignments1, _, iterations1 = solver_strict.solve(cost_matrix)
        
        # More relaxed threshold
        solver_relaxed = HopfieldAssignmentSolver(convergence_threshold=1e-2)
        assignments2, _, iterations2 = solver_relaxed.solve(cost_matrix)
        
        # Both should produce valid assignments
        assert set(assignments1) == {0, 1}
        assert set(assignments2) == {0, 1}
        
        # Stricter threshold should take more iterations
        assert iterations1 >= iterations2
    
    def test_max_iterations_limit(self):
        """Test that verifies the iteration limit is respected."""
        solver = HopfieldAssignmentSolver(max_iterations=5, convergence_threshold=1e-10)
        cost_matrix = [[1, 2], [3, 4]]
        
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Must respect iteration limit
        assert iterations <= 5
        # Still must produce valid assignments
        assert set(assignments) == {0, 1}


class TestEdgeCases:
    """Tests for edge cases."""
    
    def test_identical_costs(self):
        """Test with identical costs."""
        cost_matrix = [[1, 1], [1, 1]]
        solver = HopfieldAssignmentSolver()
        
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        assert set(assignments) == {0, 1}
        assert total_cost == 2.0  # Any assignment gives the same cost
        assert iterations > 0
    
    def test_zero_costs(self):
        """Test with zero costs."""
        cost_matrix = [[0, 0], [0, 0]]
        solver = HopfieldAssignmentSolver()
        
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        assert set(assignments) == {0, 1}
        assert total_cost == 0.0
        assert iterations > 0
    
    def test_very_large_costs(self):
        """Test with very large costs."""
        cost_matrix = [[1e6, 2e6], [3e6, 4e6]]
        solver = HopfieldAssignmentSolver()
        
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        assert set(assignments) == {0, 1}
        assert total_cost > 1e6
        assert iterations > 0


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
