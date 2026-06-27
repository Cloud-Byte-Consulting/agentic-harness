"""hillclimb.py - Bounded Hill-Climb Refinement Loop Orchestrator.

Enforces execution budgets, plateau detection, and state bounds for agentic workflows.
"""

import json
import math
import os
import time
from typing import Any, Dict, List, Optional, Tuple, Union

# Define the directive constants
CONTINUE_REFINE = "CONTINUE_REFINE"
STOP_BUDGET = "STOP_BUDGET"
STOP_PLATEAU = "STOP_PLATEAU"
STOP_REVIEW = "STOP_REVIEW"
STOP_ACCEPT = "STOP_ACCEPT"

DEFAULT_STATE_FILE = "state.json"


def is_finite_confidence(val: Any) -> bool:
    """Return True if val is a finite number (int or float)."""
    if val is None:
        return False
    if not isinstance(val, (int, float)):
        return False
    return math.isfinite(val)


def prune_runs(
    state_dict: Dict[str, Any],
    max_runs: int = 100,
    ttl_seconds: float = 21600.0,  # 6 hours default
    now: Optional[float] = None,
) -> Dict[str, Any]:
    """Prunes run states by TTL (expiration) and LRU (least recently updated) cap.

    Args:
        state_dict: Dict mapping runId -> run_state.
        max_runs: Maximum number of run states to retain.
        ttl_seconds: Expiration period in seconds (6 hours = 21600).
        now: Current timestamp (defaults to time.time()).

    Returns:
        The pruned state dictionary.
    """
    if now is None:
        now = time.time()

    # 1. TTL Pruning
    active_runs = {}
    for run_id, run_state in state_dict.items():
        last_updated = run_state.get("lastUpdated", 0.0)
        if now - last_updated <= ttl_seconds:
            active_runs[run_id] = run_state

    # 2. LRU Capping
    if len(active_runs) > max_runs:
        # Sort runs by lastUpdated ascending (oldest first)
        sorted_runs = sorted(
            active_runs.items(),
            key=lambda item: item[1].get("lastUpdated", 0.0)
        )
        # Identify how many we need to remove
        excess_count = len(active_runs) - max_runs
        for i in range(excess_count):
            run_id_to_remove = sorted_runs[i][0]
            active_runs.pop(run_id_to_remove, None)

    return active_runs


def compute_directive_pure(
    run_state: Dict[str, Any],
    verdict: str,
    gaps: List[str],
    confidence: Any,
    max_rounds: int = 3,
    max_flat_rounds: int = 2,
    min_confidence_delta: float = 0.05,
    now: Optional[float] = None,
) -> Tuple[str, Dict[str, Any]]:
    """Pure logic to compute the next directive based on state and new verification metrics.

    Args:
        run_state: The existing state dictionary for the specific runId.
        verdict: The current verification verdict ('rework', 'review', or 'accept').
        gaps: List of outstanding gaps/issues found in this round.
        confidence: The confidence score for this round.
        max_rounds: Maximum allowed rework rounds before budget stop.
        max_flat_rounds: Maximum consecutive flat (non-improving) rounds before plateau stop.
        min_confidence_delta: Minimum increase in confidence to count as progress.
        now: Timestamp to record.

    Returns:
        Tuple of (computed_directive, updated_run_state).
    """
    if now is None:
        now = time.time()

    # If the run has already halted in a sticky terminal state, preserve it.
    existing_directive = run_state.get("directive")
    if existing_directive in (STOP_BUDGET, STOP_PLATEAU, STOP_REVIEW, STOP_ACCEPT):
        return existing_directive, run_state

    # Add the current round to the execution history
    rounds = run_state.setdefault("rounds", [])
    flat_rounds = run_state.get("flatRounds", 0)

    # Determine if there was concrete improvement in this round compared to the previous
    has_improved = True
    if len(rounds) > 0 and verdict == "rework":
        prev_round = rounds[-1]
        prev_gaps_count = len(prev_round.get("gaps", []))
        curr_gaps_count = len(gaps)

        prev_conf = prev_round.get("confidence")
        curr_conf = confidence

        # Check for gap reduction
        gap_shrank = curr_gaps_count < prev_gaps_count

        # Check for meaningful confidence improvement
        conf_improved = False
        if is_finite_confidence(prev_conf) and is_finite_confidence(curr_conf):
            conf_improved = curr_conf >= (prev_conf + min_confidence_delta)

        has_improved = gap_shrank or conf_improved

    # Update flat rounds tracking
    if verdict == "rework":
        if len(rounds) == 0:
            # Baseline round: starts at 0 flat rounds
            flat_rounds = 0
        else:
            if has_improved:
                flat_rounds = 0
            else:
                flat_rounds += 1

    # Record the round metrics
    rounds.append({
        "roundIndex": len(rounds),
        "verdict": verdict,
        "gaps": list(gaps),
        "confidence": confidence,
        "timestamp": now
    })

    # Determine the directive based on the verdict and limits
    directive = CONTINUE_REFINE

    if verdict == "accept":
        directive = STOP_ACCEPT
    elif verdict == "review":
        directive = STOP_REVIEW
    elif verdict == "rework":
        # Count the number of rework rounds in this run
        rework_rounds_count = sum(1 for r in rounds if r["verdict"] == "rework")

        # 1. Plateau condition: consecutive non-improving rounds
        if flat_rounds >= max_flat_rounds:
            directive = STOP_PLATEAU
        # 2. Budget condition: max rework rounds reached
        elif rework_rounds_count >= max_rounds:
            directive = STOP_BUDGET
        else:
            directive = CONTINUE_REFINE

    # Build the updated run state
    updated_run_state = {
        "runId": run_state.get("runId"),
        "rounds": rounds,
        "flatRounds": flat_rounds,
        "directive": directive,
        "lastUpdated": now
    }

    return directive, updated_run_state


def computeDirective(
    runId: str,
    verdict: str,
    gaps: List[str],
    confidence: Any,
    state_path: str = DEFAULT_STATE_FILE,
    max_rounds: int = 3,
    max_flat_rounds: int = 2,
    min_confidence_delta: float = 0.05,
    max_lru_runs: int = 100,
    ttl_hours: float = 6.0,
    now: Optional[float] = None,
) -> str:
    """Computes the next refinement directive, loading/persisting state to state_path.

    This function coordinates state loading, computing the next directive, pruning state,
    and writing the updated state back to disk.

    Args:
        runId: A stable identifier for the active run.
        verdict: Verdict returned by validation ('rework', 'review', 'accept').
        gaps: Outstanding issues identified in this round.
        confidence: Confidence score of this validation round.
        state_path: Path to the JSON file tracking states.
        max_rounds: Maximum rework cycles allowed.
        max_flat_rounds: Max consecutive non-improving cycles before halting.
        min_confidence_delta: Minimum delta to count as progress.
        max_lru_runs: LRU cache cap for the state file.
        ttl_hours: Time-to-live for run state in hours.
        now: Custom timestamp (for testing).

    Returns:
        The decision directive string (e.g. CONTINUE_REFINE, STOP_BUDGET, etc.)
    """
    if now is None:
        now = time.time()

    # Load all states from disk
    all_states: Dict[str, Any] = {}
    if os.path.exists(state_path):
        try:
            with open(state_path, "r", encoding="utf-8") as f:
                all_states = json.load(f)
        except (json.JSONDecodeError, OSError):
            all_states = {}

    # Fetch or initialize the specific run's state
    run_state = all_states.get(runId, {
        "runId": runId,
        "rounds": [],
        "flatRounds": 0,
        "directive": None,
        "lastUpdated": now
    })

    # Compute next directive
    directive, updated_run_state = compute_directive_pure(
        run_state=run_state,
        verdict=verdict,
        gaps=gaps,
        confidence=confidence,
        max_rounds=max_rounds,
        max_flat_rounds=max_flat_rounds,
        min_confidence_delta=min_confidence_delta,
        now=now
    )

    # Save to global state registry
    all_states[runId] = updated_run_state

    # Prune outdated runs
    ttl_seconds = ttl_hours * 3600.0
    pruned_states = prune_runs(
        state_dict=all_states,
        max_runs=max_lru_runs,
        ttl_seconds=ttl_seconds,
        now=now
    )

    # Write back to disk
    try:
        # Ensure directories exist
        dir_name = os.path.dirname(state_path)
        if dir_name:
            os.makedirs(dir_name, exist_ok=True)
        with open(state_path, "w", encoding="utf-8") as f:
            json.dump(pruned_states, f, indent=2)
    except OSError:
        pass

    return directive
