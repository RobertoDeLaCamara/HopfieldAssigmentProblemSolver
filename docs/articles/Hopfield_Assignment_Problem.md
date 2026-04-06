# Solving the Assignment Problem with Hopfield Neural Networks: Energy Minimization, Greedy Decoding, and Why the Hungarian Algorithm Still Wins

The assignment problem is a classic combinatorial optimization problem: given n workers and n jobs with an n×n cost matrix, find a one-to-one assignment that minimizes total cost. The Hungarian algorithm solves it exactly in O(n³). So why implement a Hopfield neural network instead?

The answer isn't optimality — the Hopfield approach doesn't guarantee the optimal solution. The answer is that it's a useful demonstration of how constraint satisfaction and optimization can be encoded as energy minimization in a recurrent neural network, and the approximate solutions it produces are often good enough for practical applications.

---

## Encoding Constraints as Energy

The core insight of the Hopfield approach is that the constraints of the assignment problem can be expressed as terms in an energy function that the network is designed to minimize.

For an n×n assignment with neuron values V[i][j] representing "assign worker i to job j":

**Row constraint** — each worker gets exactly one job:
```
E_row = (A/2) * sum_i (sum_j V[i][j] - 1)²
```

**Column constraint** — each job gets exactly one worker:
```
E_col = (B/2) * sum_j (sum_i V[i][j] - 1)²
```

**Sum constraint** — total assignments equal n:
```
E_sum = (C/2) * (sum_{i,j} V[i][j] - n)²
```

**Cost term** — minimize total assignment cost:
```
E_cost = (D/2) * sum_{i,j} C_norm[i][j] * V[i][j]
```

The total energy is E = E_row + E_col + E_sum + E_cost with weights A=500, B=500, C=200, D=200.

The weight hierarchy is intentional. A and B are highest because the row and column constraints are hard — violating them produces an invalid assignment regardless of cost. D is lowest because cost optimization only matters once we have a valid assignment.

This hierarchy prevents the most common failure mode: the network finding a "solution" that minimizes cost by only activating low-cost neurons without satisfying the permutation constraints.

---

## The Network Dynamics

Each neuron has an internal state `u[i][j]` and an output `v[i][j] = sigmoid(u[i][j])`. The output stays in (0, 1), approaching binary values as the dynamics evolve.

The update rule comes from taking the gradient of the energy with respect to u:

```python
for x in range(n):       # worker
    for i in range(n):   # job
        term1 = -A * (sum_v_row_x - 1)    # row constraint
        term2 = -B * (sum_v_col_i - 1)    # column constraint
        term3 = -C * (sum_v_all - n)      # sum constraint
        term4 = -D * normalized_cost[x][i] # cost

        du[x, i] = term1 + term2 + term3 + term4

u += du * dt    # Euler integration, dt=0.01
v = sigmoid(u)
```

Negative signs appear because we're doing gradient *descent* on the energy, which means moving in the direction of steepest energy decrease.

**Temperature annealing via sigmoid**: The sigmoid function's effective temperature can be controlled by scaling its input. As the network evolves, the activations naturally sharpen toward binary values because the gradient updates push neurons away from the linear region of sigmoid. This is an implicit form of annealing.

---

## Convergence and Its Limits

The network is declared converged when two conditions are met simultaneously:
1. The mean change in activation between iterations is below 0.001
2. All activations satisfy `|v*(1-v)| < 0.1` — each neuron is near 0 or near 1

Both conditions are checked only after 100 iterations. This prevents false early termination when the network is in a transient state that happens to be momentarily stable.

The problem: the network can converge to a local minimum of the energy that corresponds to an invalid assignment. A worker might end up assigned to two jobs (both V[i][0] and V[i][1] high), or no jobs (all V[i][j] low). The energy function penalizes this but doesn't prevent it.

This is the fundamental limitation of Hopfield networks for constraint satisfaction: the energy landscape has local minima that aren't feasible solutions.

---

## Greedy Decoding: Guaranteeing a Valid Permutation

The greedy decoder converts the soft network state into a hard assignment regardless of whether the network converged to a valid permutation:

```python
# Rank all (worker, job) pairs by activation value
candidates = [(v[x, i], x, i) for x in range(n) for i in range(n)]
candidates.sort(reverse=True)

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

This is a standard greedy assignment: always take the highest-activation pair that doesn't conflict with already-made assignments. It's equivalent to the "highest confidence first" heuristic used in many sequence labeling tasks.

The decoder's output is always a valid permutation — n assignments, each worker and job appearing exactly once. But the permutation may not be the one the network was trying to express.

---

## The Go/Flask Architecture Choice

The system separates concerns across two services: a Go API gateway and a Python Flask solver.

**Why Go for the gateway?**

The gateway's job is request validation, timeout management, routing, and middleware. These are I/O-bound tasks with no numerical computation. Go's concurrency model (goroutines + channels) handles many concurrent requests efficiently, and its static typing catches request schema errors at compile time.

The Go handler adds three layers of timeout protection:
```go
// Context-level timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// HTTP client timeout (also 30s, separate mechanism)
client := &http.Client{Timeout: 30 * time.Second}

// Batch: each problem gets its own 30s context
for _, problem := range batch.Problems {
    ctx, cancel := context.WithTimeout(...)
    // solve problem with this context
}
```

**Why Python for the solver?**

NumPy's vectorized operations make the Euler integration loop significantly faster than equivalent Go code. The entire network update — computing row/column sums, applying the sigmoid, updating all n² neurons — runs in a few numpy operations rather than explicit nested loops.

The separation also allows independent scaling: you could run multiple Python solver replicas behind the Go gateway without changing the gateway code.

**The batch error model** is worth noting: a batch request returns HTTP 200 with per-problem success flags. This means clients process successful problems immediately rather than waiting to handle the error case. The alternative (returning 400 if any problem fails) would require clients to implement retry logic for the valid problems in a failed batch.

---

## Nginx Rate Limiting

Two separate rate limit zones protect different resources:

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;       # general API
limit_req_zone $binary_remote_addr zone=hopfield:10m rate=5r/s;   # direct solver
```

The `/hopfield/` debug route (direct to Flask) has a stricter limit (5 req/s) than the main `/api/` route (10 req/s). This reflects the relative cost: the Hopfield solver is CPU-intensive, and direct access to it from outside the Go gateway bypasses the validation and timeout protection layers.

The `burst=20 nodelay` configuration allows short bursts above the rate limit without queueing. Requests that exceed the burst are dropped immediately with 503 rather than queued, which prevents the solver from being overwhelmed by a backlog.

---

## When to Use Hopfield vs. Hungarian

The Hungarian algorithm (implemented in `scipy.optimize.linear_sum_assignment`) solves the assignment problem exactly in O(n³). For most applications, this is the right choice:

```python
from scipy.optimize import linear_sum_assignment
row_ind, col_ind = linear_sum_assignment(cost_matrix)
total_cost = cost_matrix[row_ind, col_ind].sum()
```

**Use Hopfield when**:
- You want to study or demonstrate neural network optimization
- The cost matrix is noisy and approximate solutions are acceptable
- You're integrating assignment as one component of a larger neural computation graph
- The constraint structure is more complex than standard assignment (e.g., soft constraints, partial assignments)

**Use Hungarian when**:
- You need guaranteed optimal solutions
- Determinism matters (Hopfield results vary with random initialization)
- Performance is critical (Hungarian is faster for small n)
- You're building a production system

The Hopfield approach in this project produces near-optimal solutions (typically within 5–15% of optimal) but not guaranteed optimal. For the 4×4 example matrix, the known optimal cost is 13; the Hopfield network often finds 14 (within ~8% of optimal).

---

## Conclusion

The Hopfield assignment solver demonstrates that combinatorial constraints can be encoded as energy terms and minimized through neural network dynamics. The architecture — Go gateway + Python solver + nginx — separates routing, computation, and traffic management into independent components that can evolve independently.

The greedy decoder is the practical key: it converts a soft, potentially inconsistent network state into a hard, always-valid assignment. Without it, network convergence to a local minimum (which may not satisfy all constraints) would produce unusable results.

Full implementation: `HopfieldAssigmentProblemSolver/`. Start with `make setup && make up`.
