import pytest
import numpy as np
from hopfield.src.hopfield_solver import HopfieldAssignmentSolver, solve_assignment_problem

class TestHopfieldAssignmentSolver:
    def test_initialization(self):
        """Test that the solver initializes correctly."""
        solver = HopfieldAssignmentSolver(max_iterations=100, threshold=0.01)
        assert solver.max_iterations == 100
        assert solver.threshold == 0.01
        assert solver.A == 500.0
        assert solver.B == 500.0
        assert solver.C == 200.0
        assert solver.D == 200.0

    def test_activation_function(self):
        """Test the activation function behavior."""
        solver = HopfieldAssignmentSolver()
        
        # Test values
        assert solver._activation(0) == 0.5
        assert solver._activation(10) > 0.99
        assert solver._activation(-10) < 0.01

    def test_solve_simple_2x2(self):
        """Test solving a simple 2x2 assignment problem."""
        cost_matrix = [
            [1, 2],
            [3, 4]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 2
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_3x3(self):
        """Test solving a 3x3 assignment problem."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 3
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_rectangular_matrix(self):
        """Test handling of rectangular matrices."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should handle gracefully and return valid assignments
        assert len(assignments) == 2  # Should use min of rows/cols
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_zero_matrix(self):
        """Test solving with zero cost matrix."""
        cost_matrix = [
            [0, 0],
            [0, 0]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 2
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_large_values(self):
        """Test solving with large cost values."""
        cost_matrix = [
            [1000, 2000],
            [3000, 4000]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 2
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_negative_values(self):
        """Test solving with negative cost values."""
        cost_matrix = [
            [-1, -2],
            [-3, -4]
        ]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 2
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_single_element(self):
        """Test solving with single element matrix."""
        cost_matrix = [[5]]
        
        solver = HopfieldAssignmentSolver()
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should return valid assignments
        assert len(assignments) == 1
        assert isinstance(assignments[0], int)
        assert assignments[0] >= 0
        assert total_cost >= 0
        assert iterations > 0

    def test_solve_with_invalid_input(self):
        """Test handling of invalid input."""
        solver = HopfieldAssignmentSolver()
        
        # Test empty matrix
        with pytest.raises(ValueError):
            solver.solve([])
        
        # Test matrix with empty rows
        with pytest.raises(ValueError):
            solver.solve([[]])

    def test_convergence_properties(self):
        """Test that the solver converges to reasonable solutions."""
        cost_matrix = [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ]
        
        solver = HopfieldAssignmentSolver(max_iterations=500)
        assignments, total_cost, iterations = solver.solve(cost_matrix)
        
        # Should converge within reasonable iterations
        assert iterations <= 500
        assert len(assignments) == 3
        assert all(isinstance(a, int) and a >= 0 for a in assignments)
        assert total_cost >= 0

def test_solve_assignment_problem():
    """Test the convenience function."""
    cost_matrix = [
        [1, 2],
        [3, 4]
    ]
    
    result = solve_assignment_problem(cost_matrix)
    
    assert "assignments" in result
    assert "total_cost" in result
    assert "iterations" in result
    assert "cost_matrix" in result
    
    assert len(result["assignments"]) == 2
    assert isinstance(result["total_cost"], float)
    assert isinstance(result["iterations"], int)
    assert result["cost_matrix"] == cost_matrix

def test_solve_assignment_problem_with_rectangular():
    """Test the convenience function with rectangular matrix."""
    cost_matrix = [
        [1, 2, 3],
        [4, 5, 6]
    ]
    
    result = solve_assignment_problem(cost_matrix)
    
    assert "assignments" in result
    assert "total_cost" in result
    assert "iterations" in result
    assert "cost_matrix" in result
    
    # Should handle rectangular matrix gracefully
    assert len(result["assignments"]) == 2
    assert isinstance(result["total_cost"], float)
    assert isinstance(result["iterations"], int)
