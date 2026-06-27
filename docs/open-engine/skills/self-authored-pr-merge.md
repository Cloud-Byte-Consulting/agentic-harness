# Self-Authored PR Merge

A clean workflow for reviewing and merging pull requests you authored yourself — the daily reality of solo developers and agent-heavy workflows, which GitHub's approval model doesn't really accommodate (you can't approve your own PR). The skill runs a genuine self-review pass (diff inspection with fresh eyes, not a rubber stamp), checks CI status and mergeability, handles the merge with the right strategy, and finishes with branch and worktree-safe cleanup.

## Why Build It
Solo shipping needs more review discipline, not less — nobody else is going to catch the bug. Encoding the review-check-merge-cleanup sequence as a skill means it happens the same way every time, including the steps that get skipped when you're moving fast (actually reading the diff; actually deleting the branch). It also handles the practical GitHub friction honestly instead of pretending the approval model works differently than it does.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "self-pr-merge", stored wherever my
harness loads skills from.

The skill's job: review and merge pull requests I authored myself, with real review
discipline despite GitHub not allowing self-approval.

Before writing it, confirm the gh CLI is authenticated and ask me for my merge
strategy preference (squash, merge, rebase) and my branch cleanup preference.

The skill must include: (1) trigger conditions — when I ask to merge my own PR or
review-and-merge something I wrote; (2) a genuine review pass FIRST: read the full
diff with fresh eyes, list anything questionable (bugs, debug leftovers, missing
tests, scope creep) and show me findings before merging — finding nothing must be a
conclusion, never a default; (3) pre-merge checks: CI status, mergeability, conflicts,
and an honest note about the self-approval limitation rather than working around it;
(4) the merge with my preferred strategy; (5) cleanup: delete the remote branch per my
preference, and if local worktrees are involved, use worktree-safe removal — never
plain branch deletion under a worktree; (6) a stop rule: any failing check or
unresolved review finding halts the merge and comes back to me.

After writing it, test it on my next real PR.
  </task>
</prompt>
```
