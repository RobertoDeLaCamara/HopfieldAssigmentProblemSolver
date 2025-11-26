import pytest
import numpy as np
from src.hopfield_solver import HopfieldAssignmentSolver, solve_assignment_problem

class TestHopfieldAssignmentSolver:
    def test_initialization(self):
        solver = HopfieldAssignmentSolver(max_iterations=100, threshold=0.01)
        assert solver.max_iterations == 100
        assert solver.threshold == 0.01

    def test_solve_simple_2x2(self):
        # Cost matrix:
        # [1, 2]
        # [3, 4]
        # Optimal assignment: (0, 0), (1, 1) -> Cost 1 + 4 = 5? No.
        # Wait.
        # (0, 1) -> 2
        # (1, 0) -> 3
        # Total cost 2 + 3 = 5.
        # (0, 0) -> 1
        # (1, 1) -> 4
        # Total cost 1 + 4 = 5.
        # Both are equal. Let's try a distinct one.
        
        # [1, 10]
        # [10, 1]
        # Optimal: (0, 0), (1, 1) -> Cost 2
        cost_matrix = [[1, 10], [10, 1]]
        solver = HopfieldAssignmentSolver()
        assignments, cost, iterations = solver.solve(cost_matrix)
        
        assert len(assignments) == 2
        assert assignments[0] == 0
        assert assignments[1] == 1
        assert cost == 2.0

    def test_solve_3x3(self):
        # [1, 2, 3]
        # [4, 5, 6]
        # [7, 8, 9]
        # Optimal: (0, 2), (1, 1), (2, 0) -> 3 + 5 + 7 = 15
        # Or (0, 0), (1, 2), (2, 1) -> 1 + 6 + 8 = 15
        # Or (0, 0), (1, 1), (2, 2) -> 1 + 5 + 9 = 15
        # All diagonals have same sum?
        # Let's use a clearer one.
        
        # [10, 2, 10]
        # [2, 10, 10]
        # [10, 10, 2]
        # Optimal: (0, 1), (1, 0), (2, 2) -> 2 + 2 + 2 = 6
        cost_matrix = [
            [10, 2, 10],
            [2, 10, 10],
            [10, 10, 2]
        ]
        solver = HopfieldAssignmentSolver()
        assignments, cost, iterations = solver.solve(cost_matrix)
        
        assert len(assignments) == 3
        # We expect assignments to pick the 2s.
        # Row 0 -> Col 1
        # Row 1 -> Col 0
        # Row 2 -> Col 2
        assert assignments[0] == 1
        assert assignments[1] == 0
        assert assignments[2] == 2
        assert cost == 6.0

    def test_solve_function_wrapper(self):
        cost_matrix = [[1, 10], [10, 1]]
        result = solve_assignment_problem(cost_matrix)
        
        assert result["total_cost"] == 2.0
        assert result["assignments"] == [0, 1]
        assert "iterations" in result
        assert "cost_matrix" in result

    def test_calculate_total_cost(self):
        solver = HopfieldAssignmentSolver()
        cost_matrix = np.array([[1, 2], [3, 4]])
        assignments = [1, 0] # (0, 1) -> 2, (1, 0) -> 3. Total 5.
        
        cost = solver._calculate_total_cost(assignments, cost_matrix)
        assert cost == 5.0
