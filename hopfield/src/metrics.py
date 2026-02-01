"""
Prometheus metrics collection for the Hopfield Assignment Solver.
"""

from typing import Dict
import time
from functools import wraps


class MetricsCollector:
    """Simple in-memory metrics collector."""
    
    def __init__(self):
        self.request_count = 0
        self.error_count = 0
        self.request_durations = []
        self.solve_iterations = []
        self.matrix_sizes = []
        self.batch_sizes = []
        
    def record_request(self, duration: float, status_code: int, matrix_size: int = None):
        """Record a request."""
        self.request_count += 1
        self.request_durations.append(duration)
        
        if status_code >= 400:
            self.error_count += 1
        
        if matrix_size:
            self.matrix_sizes.append(matrix_size)
    
    def record_solve(self, iterations: int, matrix_size: int, duration: float):
        """Record a solve operation."""
        self.solve_iterations.append(iterations)
        self.matrix_sizes.append(matrix_size)
        
    def record_batch(self, batch_size: int):
        """Record a batch operation."""
        self.batch_sizes.append(batch_size)
    
    def get_metrics(self) -> Dict:
        """Get current metrics."""
        return {
            "requests": {
                "total": self.request_count,
                "errors": self.error_count,
                "success_rate": (
                    (self.request_count - self.error_count) / self.request_count * 100
                    if self.request_count > 0 else 0
                )
            },
            "performance": {
                "avg_duration_ms": (
                    sum(self.request_durations) / len(self.request_durations) * 1000
                    if self.request_durations else 0
                ),
                "min_duration_ms": (
                    min(self.request_durations) * 1000
                    if self.request_durations else 0
                ),
                "max_duration_ms": (
                    max(self.request_durations) * 1000
                    if self.request_durations else 0
                )
            },
            "algorithm": {
                "avg_iterations": (
                    sum(self.solve_iterations) / len(self.solve_iterations)
                    if self.solve_iterations else 0
                ),
                "avg_matrix_size": (
                    sum(self.matrix_sizes) / len(self.matrix_sizes)
                    if self.matrix_sizes else 0
                )
            },
            "batch": {
                "avg_batch_size": (
                    sum(self.batch_sizes) / len(self.batch_sizes)
                    if self.batch_sizes else 0
                ),
                "total_batches": len(self.batch_sizes)
            }
        }
    
    def reset(self):
        """Reset all metrics."""
        self.__init__()


# Global metrics collector
metrics = MetricsCollector()


def track_request(func):
    """Decorator to track request metrics."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        status_code = 200
        
        try:
            result = func(*args, **kwargs)
            if isinstance(result, tuple):
                status_code = result[1] if len(result) > 1 else 200
            return result
        except Exception as e:
            status_code = 500
            raise
        finally:
            duration = time.time() - start_time
            metrics.record_request(duration, status_code)
    
    return wrapper
