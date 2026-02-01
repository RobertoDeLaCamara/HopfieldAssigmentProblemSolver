"""
Enhanced Flask server to expose the Hopfield algorithm as a REST API.
Includes validation, logging, metrics, and error handling.
"""

from flask import Flask, request, jsonify, g
import logging
import time
from hopfield_solver import solve_assignment_problem
import traceback
from werkzeug.exceptions import BadRequest
import os

# Import our new modules
from validation import validate_cost_matrix, validate_batch_request, ValidationError, get_validation_summary
from logging_config import setup_logging, generate_request_id
from metrics import metrics, track_request

# Initialize Flask app
app = Flask(__name__)

# Setup structured logging
log_level = os.getenv('LOG_LEVEL', 'INFO')
use_json_logging = os.getenv('JSON_LOGGING', 'true').lower() == 'true'
setup_logging(level=log_level, use_json=use_json_logging)

logger = logging.getLogger(__name__)


@app.before_request
def before_request():
    """Add request ID and start timer before each request."""
    g.request_id = request.headers.get('X-Request-ID', generate_request_id())
    g.start_time = time.time()


@app.after_request
def after_request(response):
    """Add headers and log request after completion."""
    # Add request ID to response
    response.headers['X-Request-ID'] = g.request_id
    
    # Add CORS headers
    response.headers['Access-Control-Allow-Origin'] = '*'
    response.headers['Access-Control-Allow-Headers'] = 'Content-Type, X-API-Key, Authorization'
    
    # Calculate request duration
    if hasattr(g, 'start_time'):
        duration = time.time() - g.start_time
        response.headers['X-Response-Time'] = f"{duration * 1000:.2f}ms"
        
        # Record metrics
        matrix_size = getattr(g, 'matrix_size', None)
        metrics.record_request(duration, response.status_code, matrix_size)
    
    return response


@app.route('/health', methods=['GET'])
def health_check():
    """Service health endpoint with detailed status."""
    return jsonify({
        "status": "healthy",
        "service": "hopfield-assignment-solver",
        "version": "1.0.0",
        "timestamp": time.time()
    })


@app.route('/health/ready', methods=['GET'])
def readiness_check():
    """Readiness check - service is ready to handle requests."""
    try:
        # Test that we can import the solver
        from hopfield_solver import HopfieldAssignmentSolver
        return jsonify({"status": "ready"}), 200
    except Exception as e:
        logger.error(f"Readiness check failed: {e}")
        return jsonify({"status": "not ready", "error": str(e)}), 503


@app.route('/health/live', methods=['GET'])
def liveness_check():
    """Liveness check - service is alive."""
    return jsonify({"status": "alive"}), 200


@app.route('/metrics', methods=['GET'])
def get_metrics():
    """Get service metrics."""
    return jsonify(metrics.get_metrics())


@app.route('/validation/info', methods=['GET'])
def validation_info():
    """Get validation constraints information."""
    return jsonify(get_validation_summary())


@app.route('/solve', methods=['POST'])
@track_request
def solve_assignment():
    """
    Solve an assignment problem with enhanced validation and error handling.
    
    Request body:
    {
        "cost_matrix": [[float, ...], ...]  # nxn cost matrix
    }
    
    Response:
    {
        "success": bool,
        "result": {
            "assignments": [int, ...],
            "total_cost": float,
            "iterations": int,
            "cost_matrix": [[float, ...], ...]
        },
        "request_id": string,
        "error": string  # Only if success = false
    }
    """
    try:
        data = request.get_json()
        
        if data is None:
            return jsonify({
                "success": False,
                "error": "No JSON provided in request body",
                "request_id": g.request_id
            }), 400
            
        if 'cost_matrix' not in data:
            return jsonify({
                "success": False,
                "error": "Field 'cost_matrix' is required",
                "request_id": g.request_id
            }), 400
        
        cost_matrix = data['cost_matrix']
        
        # Enhanced validation with detailed error messages
        try:
            validate_cost_matrix(cost_matrix)
        except ValidationError as e:
            logger.warning(f"Validation error: {str(e)}", extra={
                "request_id": g.request_id,
                "error": str(e)
            })
            return jsonify({
                "success": False,
                "error": str(e),
                "request_id": g.request_id
            }), 400
        
        # Store matrix size for metrics
        n = len(cost_matrix)
        g.matrix_size = n
        
        # Solve the problem
        logger.info(f"Solving {n}x{n} assignment problem", extra={
            "request_id": g.request_id,
            "matrix_size": n
        })
        
        start_time = time.time()
        result = solve_assignment_problem(cost_matrix)
        solve_duration = time.time() - start_time
        
        # Record solve metrics
        iterations = result.get('iterations', 0)
        metrics.record_solve(iterations, n, solve_duration)
        
        logger.info(f"Problem solved successfully", extra={
            "request_id": g.request_id,
            "matrix_size": n,
            "iterations": iterations,
            "solve_duration_ms": solve_duration * 1000,
            "total_cost": result.get('total_cost', 0)
        })
        
        return jsonify({
            "success": True,
            "result": result,
            "request_id": g.request_id
        })
        
    except BadRequest as e:
        logger.error(f"Bad request: {str(e)}", extra={"request_id": g.request_id})
        return jsonify({
            "success": False,
            "error": "Invalid JSON format",
            "request_id": g.request_id
        }), 400
        
    except Exception as e:
        logger.error(f"Error solving assignment problem: {str(e)}", extra={
            "request_id": g.request_id,
            "traceback": traceback.format_exc()
        })
        
        return jsonify({
            "success": False,
            "error": f"Internal server error: {str(e)}",
            "request_id": g.request_id
        }), 500


@app.route('/solve/batch', methods=['POST'])
@track_request
def solve_batch():
    """
    Solve multiple assignment problems in batch with validation.
    
    Request body:
    {
        "problems": [
            {
                "id": string,
                "cost_matrix": [[float, ...], ...]
            },
            ...
        ]
    }
    """
    try:
        data = request.get_json()
        
        if not data or 'problems' not in data:
            return jsonify({
                "success": False,
                "error": "Field 'problems' is required",
                "request_id": g.request_id
            }), 400
        
        problems = data['problems']
        
        # Validate batch request
        try:
            validate_batch_request(problems)
        except ValidationError as e:
            return jsonify({
                "success": False,
                "error": str(e),
                "request_id": g.request_id
            }), 400
        
        # Record batch size
        metrics.record_batch(len(problems))
        
        logger.info(f"Processing batch of {len(problems)} problems", extra={
            "request_id": g.request_id,
            "batch_size": len(problems)
        })
        
        results = []
        
        for problem in problems:
            problem_id = problem['id']
            cost_matrix = problem['cost_matrix']
            
            try:
                # Validate individual problem
                validate_cost_matrix(cost_matrix)
                
                # Solve the problem
                result = solve_assignment_problem(cost_matrix)
                results.append({
                    "id": problem_id,
                    "success": True,
                    "result": result
                })
                
            except ValidationError as e:
                results.append({
                    "id": problem_id,
                    "success": False,
                    "error": str(e)
                })
            except Exception as e:
                logger.error(f"Error solving problem {problem_id}: {str(e)}", extra={
                    "request_id": g.request_id,
                    "problem_id": problem_id
                })
                results.append({
                    "id": problem_id,
                    "success": False,
                    "error": str(e)
                })
        
        successful_count = sum(1 for r in results if r['success'])
        logger.info(f"Batch processing complete: {successful_count}/{len(problems)} successful", extra={
            "request_id": g.request_id,
            "batch_size": len(problems),
            "successful": successful_count
        })
        
        return jsonify({
            "success": True,
            "results": results,
            "request_id": g.request_id,
            "summary": {
                "total": len(problems),
                "successful": successful_count,
                "failed": len(problems) - successful_count
            }
        })
        
    except BadRequest as e:
        logger.error(f"Bad request: {str(e)}", extra={"request_id": g.request_id})
        return jsonify({
            "success": False,
            "error": "Invalid JSON format",
            "request_id": g.request_id
        }), 400
        
    except Exception as e:
        logger.error(f"Error in batch processing: {str(e)}", extra={
            "request_id": g.request_id,
            "traceback": traceback.format_exc()
        })
        return jsonify({
            "success": False,
            "error": f"Internal server error: {str(e)}",
            "request_id": g.request_id
        }), 500


if __name__ == '__main__':
    # Development mode
    logger.info("Starting Hopfield Assignment Solver API in development mode")
    app.run(host='0.0.0.0', port=5000, debug=False)
