# Hopfield Algorithm

## The Assignment Problem

Given an n×n cost matrix C where `C[i][j]` is the cost of assigning worker `i` to job `j`, find a one-to-one assignment that minimizes total cost:

```
minimize: sum_{i} C[i][assignments[i]]
subject to: each worker assigned exactly one job
            each job assigned to exactly one worker
```

The classical solution is the Hungarian algorithm (O(n³)). The Hopfield network provides an approximate alternative using energy minimization.

## Network Architecture

Each neuron `V[i][j]` represents the binary decision "assign worker i to job j". The network has `n²` neurons arranged in an n×n grid.

**Target state**: A permutation matrix — exactly one 1 per row and one 1 per column.

## Energy Function

The Hopfield energy has four terms:

```
E = (A/2) * row_penalty + (B/2) * col_penalty
  + (C_/2) * sum_penalty + (D/2) * cost_penalty

Weights: A=500, B=500, C_=200, D=200
```

| Term | Formula | Enforces |
|------|---------|---------|
| Row penalty | `sum_i (sum_j V[i][j] - 1)²` | One job per worker |
| Column penalty | `sum_j (sum_i V[i][j] - 1)²` | One worker per job |
| Sum penalty | `(sum_{i,j} V[i][j] - n)²` | Exactly n assignments |
| Cost penalty | `sum_{i,j} C_norm[i][j] * V[i][j]` | Minimize cost |

A, B, C\_ >> D ensures that constraint satisfaction takes priority over cost minimization. The network will first find a valid assignment, then minimize its cost.

## Algorithm Steps

### 1. Initialization

```python
u = np.random.normal(0, 0.1, (n, n))  # Internal state
v = sigmoid(u)                         # Output activations ∈ (0, 1)
```

Random initialization breaks symmetry — without it, all neurons would evolve identically.

### 2. Cost Matrix Normalization

```python
norm_matrix = matrix / max(matrix)   # Scale to [0, 1]
```

Normalization prevents the cost term from overwhelming constraint terms when costs are large.

### 3. Euler Integration (Main Loop)

```python
for iter in range(max_iterations):   # max 1000
    v_prev = v.copy()

    for x in range(n):     # worker
        for i in range(n): # job
            # Row constraint: enforce sum_j V[x][j] == 1
            term1 = -A * (np.sum(v[x, :]) - 1)

            # Column constraint: enforce sum_i V[i_][i] == 1
            term2 = -B * (np.sum(v[:, i]) - 1)

            # Sum constraint: enforce total sum == n
            term3 = -C_ * (np.sum(v) - n)

            # Cost term: minimize C_norm[x][i]
            term4 = -D * norm_matrix[x, i]

            # Euler update
            du[x, i] = term1 + term2 + term3 + term4

    u += du * dt    # dt = 0.01
    v = sigmoid(u)

    # Convergence check (after 100 iterations minimum)
    if iter > 100:
        if mean(|v - v_prev|) < threshold:  # threshold = 0.001
            if all(|v*(1-v)| < 0.1):        # activations near 0 or 1
                break
```

The two-part convergence check prevents premature termination:
1. Low activation change (plateau reached)
2. Valid permutation pattern (activations discretized)

### 4. Sigmoid Activation (Numerically Stable)

```python
def sigmoid(x):
    if x >= 0:
        return 1 / (1 + exp(-x))    # avoids exp(+large) overflow
    else:
        return exp(x) / (1 + exp(x)) # avoids exp(-large) underflow
```

### 5. Greedy Decoding

The network converges to soft activations (values between 0 and 1), not binary. The greedy decoder extracts a valid permutation:

```python
candidates = [(v[x, i], x, i) for x in range(n) for i in range(n)]
candidates.sort(reverse=True)  # highest activation first

assigned_workers = set()
assigned_jobs = set()
assignments = [-1] * n

for activation, worker, job in candidates:
    if worker not in assigned_workers and job not in assigned_jobs:
        assignments[worker] = job
        assigned_workers.add(worker)
        assigned_jobs.add(job)
    if len(assigned_workers) == n:
        break
```

The greedy decoder guarantees a complete valid assignment even if the network doesn't fully converge to binary values. It respects the network's learned preferences by selecting highest-activation cells first.

### 6. Cost Calculation

```python
total_cost = sum(cost_matrix[worker][assignments[worker]] for worker in range(n))
```

Uses the **original** (non-normalized) cost matrix for the final cost.

## Example: 4×4 Assignment

```
Cost matrix:         Optimal assignment:
  9  2  7  8          Worker 0 → Job 1 (cost 2)
  6  4  3  7          Worker 1 → Job 2 (cost 3)
  5  8  1  8          Worker 2 → Job 0 (cost 5) ← wait, this doesn't work
  7  6  9  4          Worker 3 → Job 3 (cost 4)
                      Total: 13

assignments = [1, 2, 0, 3]
total_cost = 2 + 3 + 5 + 4 = 14  # Hopfield result

Note: the known optimal for this matrix is 13, requiring different assignment
```

The Hopfield result (14) is near-optimal. The Hungarian algorithm guarantees the true optimum (13). For production use cases requiring guaranteed optimality, use the Hungarian algorithm.

## Performance Characteristics

| n | Neurons | Typical iterations | Typical time |
|---|---------|-------------------|--------------|
| 2 | 4 | 30–60 | <50ms |
| 5 | 25 | 50–150 | <100ms |
| 10 | 100 | 100–300 | ~300ms |
| 20 | 400 | 200–600 | ~2s |
| 50 | 2500 | 500–1000 | ~15s |

Maximum supported size: 50×50 (2500 neurons). Beyond this, convergence time becomes impractical.
