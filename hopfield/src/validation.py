"""
Enhanced validation utilities for the Hopfield Assignment Solver API.
"""

from typing import List, Optional, Tuple

import numpy as np

# Configuration constants
MAX_MATRIX_SIZE = 50
MIN_MATRIX_SIZE = 2
MIN_COST_VALUE = 0
MAX_COST_VALUE = 1e9
MAX_REQUEST_SIZE_MB = 10


class ValidationError(Exception):
    """Custom exception for validation errors."""

    pass


def validate_cost_matrix(cost_matrix: List[List[float]]) -> Tuple[bool, Optional[str]]:
    """
    Validate a cost matrix for the assignment problem.

    Args:
        cost_matrix: The cost matrix to validate

    Returns:
        Tuple of (is_valid, error_message)

    Raises:
        ValidationError: If validation fails with detailed error message
    """
    # Check if matrix is empty
    if not cost_matrix:
        raise ValidationError("Cost matrix cannot be empty")

    # Check if matrix is a list
    if not isinstance(cost_matrix, list):
        raise ValidationError("Cost matrix must be a list")

    n = len(cost_matrix)

    # Check matrix size constraints
    if n < MIN_MATRIX_SIZE:
        raise ValidationError(
            f"Matrix size {n}x{n} is too small. "
            f"Minimum size is {MIN_MATRIX_SIZE}x{MIN_MATRIX_SIZE}"
        )

    if n > MAX_MATRIX_SIZE:
        raise ValidationError(
            f"Matrix size {n}x{n} exceeds maximum allowed size of "
            f"{MAX_MATRIX_SIZE}x{MAX_MATRIX_SIZE}. "
            f"Please reduce the matrix size or contact support for larger matrices."
        )

    # Validate each row
    for i, row in enumerate(cost_matrix):
        # Check if row is a list
        if not isinstance(row, list):
            raise ValidationError(f"Row {i} must be a list, got {type(row).__name__}")

        # Check if matrix is square
        if len(row) != n:
            raise ValidationError(
                f"Matrix must be square. Row {i} has {len(row)} elements, "
                f"expected {n}. This is not a valid assignment problem."
            )

        # Validate each cost value
        for j, cost in enumerate(row):
            # Check if cost is a number
            if not isinstance(cost, (int, float)):
                raise ValidationError(
                    f"Cost at position [{i}][{j}] must be a number, "
                    f"got {type(cost).__name__}: {cost}"
                )

            # Check for NaN or Inf
            if not np.isfinite(cost):
                if np.isnan(cost):
                    raise ValidationError(
                        f"Cost at position [{i}][{j}] is NaN. "
                        f"All costs must be valid numbers."
                    )
                else:
                    raise ValidationError(
                        f"Cost at position [{i}][{j}] is infinite. "
                        f"All costs must be finite numbers."
                    )

            # Check cost value range
            if cost < MIN_COST_VALUE:
                raise ValidationError(
                    f"Cost at position [{i}][{j}] is {cost}, which is negative. "
                    f"All costs must be non-negative (>= {MIN_COST_VALUE})."
                )

            if cost > MAX_COST_VALUE:
                raise ValidationError(
                    f"Cost at position [{i}][{j}] is {cost}, which exceeds "
                    f"maximum allowed value of {MAX_COST_VALUE}. "
                    f"Please scale your costs down."
                )

    return True, None


def validate_batch_request(problems: List[dict]) -> Tuple[bool, Optional[str]]:
    """
    Validate a batch request.

    Args:
        problems: List of problem dictionaries

    Returns:
        Tuple of (is_valid, error_message)
    """
    if not isinstance(problems, list):
        raise ValidationError("Problems must be a list")

    if len(problems) == 0:
        raise ValidationError("Problems list cannot be empty")

    if len(problems) > 100:
        raise ValidationError(
            f"Batch size of {len(problems)} exceeds maximum of 100 problems. "
            f"Please split into smaller batches."
        )

    for i, problem in enumerate(problems):
        if not isinstance(problem, dict):
            raise ValidationError(
                f"Problem {i} must be a dictionary, got {type(problem).__name__}"
            )

        if "id" not in problem:
            raise ValidationError(f"Problem {i} is missing required field 'id'")

        if "cost_matrix" not in problem:
            raise ValidationError(
                f"Problem {i} (id: {problem.get('id', 'unknown')}) "
                f"is missing required field 'cost_matrix'"
            )

    return True, None


def get_validation_summary() -> dict:
    """
    Get a summary of validation constraints.

    Returns:
        Dictionary with validation constraints
    """
    return {
        "matrix_size": {"min": MIN_MATRIX_SIZE, "max": MAX_MATRIX_SIZE},
        "cost_value": {"min": MIN_COST_VALUE, "max": MAX_COST_VALUE},
        "batch_size": {"max": 100},
        "request_size_mb": {"max": MAX_REQUEST_SIZE_MB},
    }
