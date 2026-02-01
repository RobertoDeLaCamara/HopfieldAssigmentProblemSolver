"""
Flask server to expose the Hopfield algorithm as a REST API.
"""

from flask import Flask, request, jsonify
import logging
from hopfield_solver import solve_assignment_problem
import traceback
from werkzeug.exceptions import BadRequest

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@app.route('/health', methods=['GET'])
def health_check():
    """Service health endpoint."""
    return jsonify({"status": "healthy", "service": "hopfield-assignment-solver"})


@app.route('/solve', methods=['POST'])
def solve_assignment():
    """
    Solve an assignment problem.
    
    Request body:
    {
        "cost_matrix": [[float, ...], ...]  # nxn cost matrix
    }
    
    Response:
    {
        "success": bool,
        "result": {
            "assignments": [int, ...],  # Assignments: assignments[i] = job of worker i
            "total_cost": float,
            "iterations": int,
            "cost_matrix": [[float, ...], ...]
        },
        "error": string  # Only if success = false
    }
    """
    try:
        data = request.get_json()
        
        if data is None:
            return jsonify({
                "success": False,
                "error": "No JSON provided in request body"
            }), 400
            
        if 'cost_matrix' not in data:
            return jsonify({
                "success": False,
                "error": "Field 'cost_matrix' is required"
            }), 400
        
        cost_matrix = data['cost_matrix']
        
        # Validate the cost matrix
        if not isinstance(cost_matrix, list) or len(cost_matrix) == 0:
            return jsonify({
                "success": False,
                "error": "The cost matrix must be a non-empty list"
            }), 400
        
        n = len(cost_matrix)
        for i, row in enumerate(cost_matrix):
            if not isinstance(row, list) or len(row) != n:
                return jsonify({
                    "success": False,
                    "error": f"Row {i} must be a list of {n} elements"
                }), 400
            
            for j, cost in enumerate(row):
                if not isinstance(cost, (int, float)):
                    return jsonify({
                        "success": False,
                        "error": f"Cost at position [{i}][{j}] must be a number"
                    }), 400
        
        # Solve the problem
        logger.info(f"Solving {n}x{n} assignment problem")
        result = solve_assignment_problem(cost_matrix)
        
        return jsonify({
            "success": True,
            "result": result
        })
        
    except BadRequest as e:
        logger.error(f"Bad request: {str(e)}")
        return jsonify({
            "success": False,
            "error": "Invalid JSON format"
        }), 400
        
    except Exception as e:
        logger.error(f"Error solving assignment problem: {str(e)}")
        logger.error(traceback.format_exc())
        
        return jsonify({
            "success": False,
            "error": f"Internal server error: {str(e)}"
        }), 500


@app.route('/solve/batch', methods=['POST'])
def solve_batch():
    """
    Solve multiple assignment problems in batch.
    
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
                "error": "Field 'problems' is required"
            }), 400
        
        problems = data['problems']
        
        if not isinstance(problems, list):
            return jsonify({
                "success": False,
                "error": "Field 'problems' must be a list"
            }), 400
            
        if len(problems) == 0:
            return jsonify({
                "success": False,
                "error": "Problems list cannot be empty"
            }), 400
        
        results = []
        
        for problem in problems:
            if 'id' not in problem or 'cost_matrix' not in problem:
                results.append({
                    "id": problem.get('id', 'unknown'),
                    "success": False,
                    "error": "'id' and 'cost_matrix' are required"
                })
                continue
            
            try:
                result = solve_assignment_problem(problem['cost_matrix'])
                results.append({
                    "id": problem['id'],
                    "success": True,
                    "result": result
                })
            except Exception as e:
                results.append({
                    "id": problem['id'],
                    "success": False,
                    "error": str(e)
                })
        
        return jsonify({
            "success": True,
            "results": results
        })
        
    except BadRequest as e:
        logger.error(f"Bad request: {str(e)}")
        return jsonify({
            "success": False,
            "error": "Invalid JSON format"
        }), 400
        
    except Exception as e:
        logger.error(f"Error in batch processing: {str(e)}")
        return jsonify({
            "success": False,
            "error": f"Internal server error: {str(e)}"
        }), 500


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=False)
