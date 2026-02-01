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
    def __init__(self, max_iterations: int = 1000, threshold: float = 0.001):
        """
        Initialize the Hopfield solver.
        
        Args:
            max_iterations: Maximum number of iterations to run
            threshold: Convergence threshold
        """
        self.max_iterations = max_iterations
        self.threshold = threshold
        
        # Hyperparameters for the energy function
        # A, B, C enforce valid assignment constraints (one per row, one per col, total n)
        # D minimizes the cost
        self.A = 500.0
        self.B = 500.0
        self.C = 200.0
        self.D = 200.0

    def _kronecker_delta(self, i: int, j: int) -> int:
        return 1 if i == j else 0

    def _activation(self, x) -> float:
        """Sigmoid activation function."""
        # Using a numerically stable sigmoid for scalars and arrays
        if np.isscalar(x):
            if x >= 0:
                z = np.exp(-x)
                return 1 / (1 + z)
            else:
                z = np.exp(x)
                return z / (1 + z)
        else:
            # Handle arrays element-wise
            result = np.zeros_like(x, dtype=float)
            pos_mask = x >= 0
            neg_mask = ~pos_mask
            
            # Positive values
            z_pos = np.exp(-x[pos_mask])
            result[pos_mask] = 1 / (1 + z_pos)
            
            # Negative values
            z_neg = np.exp(x[neg_mask])
            result[neg_mask] = z_neg / (1 + z_neg)
            
            return result

    def solve(self, cost_matrix: List[List[float]]) -> Tuple[List[int], float, int]:
        """
        Solve the assignment problem using a Hopfield network.
        
        Args:
            cost_matrix: nxn cost matrix
            
        Returns:
            Tuple containing:
            - List of assignments (index of job for each worker)
            - Total cost
            - Number of iterations
        """
        # Validate input
        if not cost_matrix or not cost_matrix[0]:
            raise ValueError("Cost matrix cannot be empty")
            
        n_rows = len(cost_matrix)
        n_cols = len(cost_matrix[0])
        
        # Check if matrix is square
        if n_rows != n_cols:
            raise ValueError(f"Cost matrix must be square, got {n_rows}x{n_cols}")
            
        n = n_rows
        matrix = cost_matrix
            
        # Convert to numpy array
        matrix = np.array(matrix)
        
        # Normalize cost matrix to [0, 1] range for better convergence
        max_cost = np.max(matrix)
        if max_cost > 0:
            norm_matrix = matrix / max_cost
        else:
            norm_matrix = matrix

        # Initialize neurons with random values plus noise
        # u is the internal state, v is the output (activation)
        u = np.random.normal(0, 0.1, (n, n))
        v = self._activation(u)
        
        iterations = 0
        prev_v = np.copy(v)
        
        # Time step for Euler method
        dt = 0.01
        
        for it in range(self.max_iterations):
            iterations = it + 1
            
            # Compute equations of motion
            du = np.zeros((n, n))
            
            # Sum of activations in each row (minus self)
            row_sums = np.sum(v, axis=1)
            # Sum of activations in each col (minus self)
            col_sums = np.sum(v, axis=0)
            # Total sum of activations
            total_sum = np.sum(v)
            
            for x in range(n):
                for i in range(n):
                    # Constraint 1: One 1 per row
                    term1 = -self.A * (row_sums[x] - 1)
                    
                    # Constraint 2: One 1 per column
                    term2 = -self.B * (col_sums[i] - 1)
                    
                    # Constraint 3: Total n units active
                    term3 = -self.C * (total_sum - n)
                    
                    # Data term: Minimize cost
                    term4 = -self.D * norm_matrix[x, i]
                    
                    du[x, i] = term1 + term2 + term3 + term4
            
            # Update internal state
            u += du * dt
            
            # Update output
            v = self._activation(u)
            
            # Check convergence
            diff = np.mean(np.abs(v - prev_v))
            if diff < self.threshold and it > 100:
                # Also check if we have a valid permutation matrix (close to 0 or 1)
                if np.all(np.abs(v * (1 - v)) < 0.1):
                    break
            
            prev_v = np.copy(v)
            
        # Discretize result to get permutation matrix
        # We use a greedy approach on the final activations to ensure valid assignment
        assignments = [-1] * n
        final_assignments = [-1] * n
        final_v = np.copy(v)
        
        # Simple greedy decoding: pick max in row, mask that col, repeat
        # This ensures we return a valid assignment even if the network didn't fully converge
        rows_indices = list(range(n))
        cols_indices = list(range(n))
        
        temp_assignments = []
        
        # Create a list of (value, row, col) tuples
        candidates = []
        for r in range(n):
            for c in range(n):
                candidates.append((final_v[r, c], r, c))
        
        # Sort by activation value descending
        candidates.sort(key=lambda x: x[0], reverse=True)
        
        assigned_rows = set()
        assigned_cols = set()
        
        for val, r, c in candidates:
            if r not in assigned_rows and c not in assigned_cols:
                assignments[r] = c
                final_assignments[r] = int(c)
                assigned_rows.add(r)
                assigned_cols.add(c)
                
        # Fill in any remaining (shouldn't happen if n*n)
        # But just in case
        if len(assigned_rows) < n:
            # Fallback for failed convergence
            logger.warning("Hopfield network did not converge to a valid permutation.")
            # Fill missing
            missing_rows = list(set(range(n)) - assigned_rows)
            missing_cols = list(set(range(n)) - assigned_cols)
            for r, c in zip(missing_rows, missing_cols):
                final_assignments[r] = int(c)

        total_cost = self._calculate_total_cost(final_assignments, matrix)
        return final_assignments, total_cost, iterations

    def _calculate_total_cost(self, assignments: List[int], cost_matrix: np.ndarray) -> float:
        """
        Calculate the total cost of the assignment by summing up the costs of each worker's job.
        
        For example, if the cost matrix is [[1, 2], [3, 4]] and the assignments are [0, 1], then the total cost will be
        1 + 4 = 5.
        
        Args:
            assignments: List of job indices assigned to each worker
            cost_matrix: nxn cost matrix
            
        Returns:
            Total cost of the assignment
        """
        total = 0.0
        for worker, job in enumerate(assignments):
            if job != -1:
                total += cost_matrix[worker, job]
        return total


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
