"""Solver side of the validation contract shared with the Go gateway (contract/)."""

import json
import os
import sys

import pytest
import yaml

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

from validation import (  # noqa: E402
    ValidationError,
    get_validation_summary,
    validate_batch_request,
    validate_cost_matrix,
)

CONTRACT_DIR = os.path.join(os.path.dirname(__file__), "..", "..", "contract")
with open(os.path.join(CONTRACT_DIR, "validation_cases.json")) as f:
    CONTRACT = json.load(f)


def matrix_of(case):
    if "generate" in case:
        n, fill = case["generate"]["size"], case["generate"]["fill"]
        return [[fill] * n for _ in range(n)]
    return case["matrix"]


@pytest.mark.parametrize("case", CONTRACT["matrix_cases"], ids=lambda c: c["name"])
def test_matrix_cases(case):
    if case["valid"]:
        assert validate_cost_matrix(matrix_of(case)) == (True, None)
    else:
        with pytest.raises(ValidationError):
            validate_cost_matrix(matrix_of(case))


@pytest.mark.parametrize("case", CONTRACT["batch_size_cases"], ids=lambda c: c["name"])
def test_batch_size_cases(case):
    problems = [{"id": str(i), "cost_matrix": [[1, 2], [3, 4]]} for i in range(case["size"])]
    if case["valid"]:
        assert validate_batch_request(problems) == (True, None)
    else:
        with pytest.raises(ValidationError):
            validate_batch_request(problems)


def test_solver_limits_equal_the_contract():
    limits = CONTRACT["limits"]
    summary = get_validation_summary()
    assert summary["matrix_size"] == limits["matrix_size"]
    assert summary["cost_value"] == limits["cost_value"]
    assert summary["batch_size"] == limits["batch_size"]


def test_openapi_documents_the_same_limits():
    """contract/openapi.yaml is documentation; this keeps it from drifting from the enforced limits."""
    with open(os.path.join(CONTRACT_DIR, "openapi.yaml")) as f:
        spec = yaml.safe_load(f)
    limits = CONTRACT["limits"]
    schemas = spec["components"]["schemas"]

    matrix = schemas["CostMatrix"]
    assert (matrix["minItems"], matrix["maxItems"]) == (limits["matrix_size"]["min"], limits["matrix_size"]["max"])
    row = matrix["items"]
    assert (row["minItems"], row["maxItems"]) == (limits["matrix_size"]["min"], limits["matrix_size"]["max"])
    cost = row["items"]
    assert (cost["minimum"], cost["maximum"]) == (limits["cost_value"]["min"], limits["cost_value"]["max"])
    assert schemas["BatchRequest"]["properties"]["problems"]["maxItems"] == limits["batch_size"]["max"]
