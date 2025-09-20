"""
Implementation of the Hopfield algorithm to solve the assignment problem.
The assignment problem consists of assigning n jobs to n workers in such a way
that the total cost is minimized, given the costs of assigning each worker to each job.
"""

import numpy as np
from typing import List, Tuple, Optional
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class HopfieldAssignmentSolver:
    """
    Assignment problem solver using Hopfield neural networks.
    """
    
    def __init__(self, max_iterations: int = 1000, convergence_threshold: float = 1e-6):
        """
        Initialize the solver.
        
        Args:
            max_iterations: Maximum number of iterations
            convergence_threshold: Convergence threshold
        """
        self.max_iterations = max_iterations
        self.convergence_threshold = convergence_threshold
    
    def solve(self, cost_matrix: List[List[float]]) -> Tuple[List[int], float, int]:
        """
        Solve the assignment problem using the Hopfield algorithm.
        
        Args:
            cost_matrix: nxn cost matrix where cost_matrix[i][j] is the cost
                        of assigning worker i to job j.
        
        Returns:
            Tuple with (assignments, total_cost, iterations_used)
            - assignments: List where assignments[i] is the job assigned to worker i
            - total_cost: Total cost of the optimal assignment
            - iterations_used: Number of iterations used
        """
        cost_matrix = np.array(cost_matrix, dtype=np.float64)
        n = cost_matrix.shape[0]
        
        if cost_matrix.shape[0] != cost_matrix.shape[1]:
            raise ValueError("The cost matrix must be square")
        
        # Initialize the neuron state matrix
        # V[i][j] = 1 if worker i is assigned to job j
        V = np.random.random((n, n))
        
        # Algorithm parameters
        A = 1.0  # Weight for one job per worker constraint
        B = 1.0  # Weight for one worker per job constraint
        C = 1.0  # Weight for complete assignment constraint
        D = 1.0  # Weight for cost minimization
        
        # Hopfield network weight matrix
        W = np.zeros((n*n, n*n))
        
        # Build the weight matrix
        for i in range(n):
            for j in range(n):
                for k in range(n):
                    for l in range(n):
                        idx1 = i * n + j
                        idx2 = k * n + l
                        
                        # Constraint: each worker must have exactly one job
                        if i == k and j != l:
                            W[idx1, idx2] -= A
                        
                        # Constraint: each job must have exactly one worker
                        if j == l and i != k:
                            W[idx1, idx2] -= B
                        
                        # Constraint: complete assignment (n assignments)
                        W[idx1, idx2] -= C
                        
                        # Cost minimization
                        if i == k and j == l:
                            W[idx1, idx2] -= D * cost_matrix[i, j]
        
        # Thresholds
        theta = np.zeros(n*n)
        for i in range(n):
            for j in range(n):
                idx = i * n + j
                theta[idx] = C * n
        
        # Update algorithm
        for iteration in range(self.max_iterations):
            V_old = V.copy()
            
            # Update each neuron
            for i in range(n):
                for j in range(n):
                    idx = i * n + j
                    
                    # Calculate total input
                    total_input = np.sum(W[idx, :] * V.flatten()) - theta[idx]
                    
                    # Sigmoid activation function
                    V[i, j] = 1.0 / (1.0 + np.exp(-total_input))
            
            # Check convergence
            if np.max(np.abs(V - V_old)) < self.convergence_threshold:
                logger.info(f"Convergence reached in {iteration + 1} iterations")
                break
        
        # Convert continuous output to discrete assignments
        assignments = self._discretize_assignments(V)
        total_cost = self._calculate_total_cost(assignments, cost_matrix)
        
        return assignments, total_cost, iteration + 1
    
    def _discretize_assignments(self, V: np.ndarray) -> List[int]:
        """
        Convert continuous neuron output to discrete assignments.
        Uses simplified Hungarian algorithm to guarantee a valid assignment.
        """
        n = V.shape[0]
        assignments = [-1] * n
        
        # Create a copy of V to modify
        V_copy = V.copy()
        
        for _ in range(n):
            # Find the maximum remaining value
            max_val = -np.inf
            max_i, max_j = -1, -1
            
            for i in range(n):
                for j in range(n):
                    if assignments[i] == -1 and V_copy[i, j] > max_val:
                        max_val = V_copy[i, j]
                        max_i, max_j = i, j
            
            # Assign
            assignments[max_i] = max_j
            
            # Mark row and column as used
            V_copy[max_i, :] = -np.inf
            V_copy[:, max_j] = -np.inf
        
        return assignments
    
    def _calculate_total_cost(self, assignments: List[int], cost_matrix: np.ndarray) -> float:
        """Calculate the total cost of the assignment."""
        total_cost = 0.0
        for worker, job in enumerate(assignments):
            total_cost += cost_matrix[worker, job]
        return total_cost


def solve_assignment_problem(cost_matrix: List[List[float]]) -> dict:
    """
    Convenience function to solve the assignment problem.
    
    Args:
        cost_matrix: nxn cost matrix
    
    Returns:
        Dictionary with the results
    """
    solver = HopfieldAssignmentSolver()
    assignments, total_cost, iterations = solver.solve(cost_matrix)
    
    return {
        "assignments": assignments,
        "total_cost": float(total_cost),
        "iterations": iterations,
        "cost_matrix": cost_matrix
    }


if __name__ == "__main__":
    # Usage example
    cost_matrix = [
        [9, 2, 7, 8],
        [6, 4, 3, 7],
        [5, 8, 1, 8],
        [7, 6, 9, 4]
    ]
    
    result = solve_assignment_problem(cost_matrix)
    print("Cost matrix:")
    for row in cost_matrix:
        print(row)
    
    print(f"\nAssignments: {result['assignments']}")
    print(f"Total cost: {result['total_cost']}")
    print(f"Iterations: {result['iterations']}")
