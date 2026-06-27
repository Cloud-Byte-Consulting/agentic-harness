"""test_hillclimb.py - Unit tests for the hillclimb refinement loop.
"""

import json
import math
import os
import sys
import unittest
from unittest.mock import patch

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "examples"))

# Import the hillclimb module under test
import hillclimb
from hillclimb import (
    CONTINUE_REFINE,
    STOP_BUDGET,
    STOP_PLATEAU,
    STOP_REVIEW,
    STOP_ACCEPT,
    computeDirective,
    is_finite_confidence,
    prune_runs,
)

STATE_TEST_FILE = "state_test.json"


class TestHillClimbRefinement(unittest.TestCase):

    def setUp(self):
        # Remove state_test.json if it exists to start fresh in each test
        if os.path.exists(STATE_TEST_FILE):
            os.remove(STATE_TEST_FILE)

    def tearDown(self):
        # Clean up testing state file
        if os.path.exists(STATE_TEST_FILE):
            os.remove(STATE_TEST_FILE)

    def test_finite_confidence_checking(self):
        """Verify is_finite_confidence utility matches requirements."""
        self.assertTrue(is_finite_confidence(0.95))
        self.assertTrue(is_finite_confidence(0))
        self.assertTrue(is_finite_confidence(-0.5))
        self.assertFalse(is_finite_confidence(None))
        self.assertFalse(is_finite_confidence("0.95"))
        self.assertFalse(is_finite_confidence(float("nan")))
        self.assertFalse(is_finite_confidence(float("inf")))
        self.assertFalse(is_finite_confidence(float("-inf")))

    def test_basic_accept_and_review(self):
        """Verify standard terminal verdicts (accept and review) are handled instantly."""
        # Accept verdict
        dir1 = computeDirective(
            runId="run-1",
            verdict="accept",
            gaps=[],
            confidence=0.9,
            state_path=STATE_TEST_FILE
        )
        self.assertEqual(dir1, STOP_ACCEPT)

        # Review verdict
        dir2 = computeDirective(
            runId="run-2",
            verdict="review",
            gaps=["mild-typo"],
            confidence=0.8,
            state_path=STATE_TEST_FILE
        )
        self.assertEqual(dir2, STOP_REVIEW)

    def test_continue_refine_on_first_rework(self):
        """Verify a run starts in CONTINUE_REFINE on the initial rework round."""
        dir1 = computeDirective(
            runId="run-rework-start",
            verdict="rework",
            gaps=["gap1", "gap2"],
            confidence=0.6,
            state_path=STATE_TEST_FILE
        )
        self.assertEqual(dir1, CONTINUE_REFINE)

    def test_budget_exhaustion(self):
        """Verify maxRounds (3) rework cycles halts the loop with STOP_BUDGET."""
        run_id = "run-budget"
        
        # Round 0 (rework cycle 1)
        d0 = computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE)
        self.assertEqual(d0, CONTINUE_REFINE)
        
        # Round 1 (rework cycle 2) - we improve so plateau doesn't trigger
        d1 = computeDirective(run_id, "rework", [], 0.8, STATE_TEST_FILE)
        self.assertEqual(d1, CONTINUE_REFINE)
        
        # Round 2 (rework cycle 3) - budget reached (3 rework cycles total)
        d2 = computeDirective(run_id, "rework", ["new-gap"], 0.7, STATE_TEST_FILE)
        self.assertEqual(d2, STOP_BUDGET)

    def test_plateau_detection_by_gaps_not_shrinking(self):
        """Verify loop stops after 2 flat rounds where gaps do not shrink and confidence drifts."""
        run_id = "run-plateau-gaps"
        
        # Round 0 (baseline)
        d0 = computeDirective(run_id, "rework", ["gap1", "gap2"], 0.5, STATE_TEST_FILE)
        self.assertEqual(d0, CONTINUE_REFINE)

        # Round 1 (Flat round 1: gaps size is same, confidence is same)
        d1 = computeDirective(run_id, "rework", ["gap1", "gap3"], 0.5, STATE_TEST_FILE)
        self.assertEqual(d1, CONTINUE_REFINE)

        # Round 2 (Flat round 2: gaps size is same, confidence is same) -> STOP_PLATEAU
        d2 = computeDirective(run_id, "rework", ["gap1", "gap4"], 0.5, STATE_TEST_FILE)
        self.assertEqual(d2, STOP_PLATEAU)

    def test_plateau_not_triggered_if_interrupted_by_improvement(self):
        """Verify one noisy flat round does not abort if subsequent round improves."""
        run_id = "run-plateau-recovered"
        
        # Round 0 (baseline)
        d0 = computeDirective(run_id, "rework", ["gap1", "gap2"], 0.5, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d0, CONTINUE_REFINE)

        # Round 1 (Flat round 1)
        d1 = computeDirective(run_id, "rework", ["gap1", "gap2"], 0.5, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d1, CONTINUE_REFINE)

        # Round 2 (Improvement: gaps shrink) -> resets flat count
        d2 = computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d2, CONTINUE_REFINE)

        # Round 3 (Flat round 1)
        d3 = computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d3, CONTINUE_REFINE)

        # Round 4 (Flat round 2) -> STOP_PLATEAU
        d4 = computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d4, STOP_PLATEAU)

    def test_improvement_by_confidence_delta(self):
        """Verify improvement triggers on confidence delta >= 0.05 even if gaps remain constant."""
        run_id = "run-conf-delta"
        
        # Round 0
        computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE)
        
        # Round 1 (Improvement: confidence rose by 0.05)
        d1 = computeDirective(run_id, "rework", ["gap1"], 0.55, STATE_TEST_FILE)
        self.assertEqual(d1, CONTINUE_REFINE)

        # Round 2 (Rework budget limits it, but flatRounds was reset to 0)
        # Let's run with higher max_rounds to verify flat count resetting:
        d2 = computeDirective(run_id, "rework", ["gap1"], 0.55, STATE_TEST_FILE, max_rounds=5)
        # Round 2 was flat (no change in gaps or confidence). flatRounds = 1
        self.assertEqual(d2, CONTINUE_REFINE)

        # Round 3 (Improvement: confidence rose by 0.06) -> resets flatRounds to 0
        d3 = computeDirective(run_id, "rework", ["gap1"], 0.61, STATE_TEST_FILE, max_rounds=5)
        self.assertEqual(d3, CONTINUE_REFINE)

    def test_no_improvement_on_small_confidence_delta(self):
        """Verify confidence increase < 0.05 is not counted as improvement (flat round)."""
        run_id = "run-small-conf-delta"
        
        # Round 0
        computeDirective(run_id, "rework", ["gap1"], 0.50, STATE_TEST_FILE)
        
        # Round 1 (Flat round 1: confidence rose by only 0.04)
        d1 = computeDirective(run_id, "rework", ["gap1"], 0.54, STATE_TEST_FILE)
        self.assertEqual(d1, CONTINUE_REFINE)

        # Round 2 (Flat round 2: confidence rose by 0.04 again) -> plateau
        d2 = computeDirective(run_id, "rework", ["gap1"], 0.58, STATE_TEST_FILE)
        self.assertEqual(d2, STOP_PLATEAU)

    def test_non_finite_confidence_handling(self):
        """Verify NaN, Inf, and None confidence do not trigger spurious improvement."""
        run_id = "run-non-finite"
        
        # Round 0
        computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE)
        
        # Round 1 (Flat round 1: confidence is None)
        d1 = computeDirective(run_id, "rework", ["gap1"], None, STATE_TEST_FILE)
        self.assertEqual(d1, CONTINUE_REFINE)

        # Round 2 (Flat round 2: confidence is NaN) -> STOP_PLATEAU
        d2 = computeDirective(run_id, "rework", ["gap1"], float("nan"), STATE_TEST_FILE)
        self.assertEqual(d2, STOP_PLATEAU)

    def test_sticky_stop(self):
        """Verify once stopped (budget or plateau), subsequent checks on the same runId remain stopped."""
        run_id = "run-sticky"
        
        # Round 0, 1, 2 force a STOP_BUDGET
        computeDirective(run_id, "rework", ["gap1"], 0.5, STATE_TEST_FILE)
        computeDirective(run_id, "rework", ["gap1"], 0.6, STATE_TEST_FILE) # improved
        d2 = computeDirective(run_id, "rework", ["gap1"], 0.6, STATE_TEST_FILE) # budget reached
        self.assertEqual(d2, STOP_BUDGET)

        # Send a perfect accepted run - should still return STOP_BUDGET (sticky check)
        d3 = computeDirective(run_id, "accept", [], 1.0, STATE_TEST_FILE)
        self.assertEqual(d3, STOP_BUDGET)

        # A different runId starts fresh
        d_new = computeDirective("run-fresh", "accept", [], 1.0, STATE_TEST_FILE)
        self.assertEqual(d_new, STOP_ACCEPT)

    def test_state_pruning_ttl_and_lru(self):
        """Verify state registry is pruned correctly by age and capacity limits."""
        states = {
            "run-old": {"runId": "run-old", "lastUpdated": 1000.0, "rounds": []},
            "run-med": {"runId": "run-med", "lastUpdated": 5000.0, "rounds": []},
            "run-new": {"runId": "run-new", "lastUpdated": 9000.0, "rounds": []},
        }

        # 1. Test TTL pruning (keep only things newer than 9000 - 3000 = 6000 seconds)
        pruned_ttl = prune_runs(states, max_runs=10, ttl_seconds=3000.0, now=9000.0)
        self.assertNotIn("run-old", pruned_ttl)
        self.assertNotIn("run-med", pruned_ttl)
        self.assertIn("run-new", pruned_ttl)

        # 2. Test LRU capacity capping (max_runs=2)
        pruned_lru = prune_runs(states, max_runs=2, ttl_seconds=10000.0, now=9000.0)
        self.assertEqual(len(pruned_lru), 2)
        # Oldest one ("run-old") should be evicted
        self.assertNotIn("run-old", pruned_lru)
        self.assertIn("run-med", pruned_lru)
        self.assertIn("run-new", pruned_lru)


if __name__ == "__main__":
    unittest.main()
