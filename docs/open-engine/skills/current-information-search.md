# Current-Information Search

Routes the agent's web research through a search API built for discovering new information (Perplexity's API is the canonical choice) instead of the harness's built-in search. The skill defines when to use it — recent releases, pricing, anything that may contradict the model's training data — and carries the exact API call shape, default model choice, and key location. Optionally it can be wired in as a hook so all web searches redirect automatically.

## Why Build It
Agents are confidently out of date. The single most common failure mode in AI-assisted research is the model "confirming" stale training data instead of discovering what changed last week. A dedicated search skill turns "search for current info" from a hope into a procedure — and it makes every other research-flavored skill in this library more trustworthy.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "current-info-search", stored
wherever my harness loads skills from.

The skill's job: when I ask about anything that changes quickly — AI model releases,
pricing, software versions, news, APIs — call the Perplexity API directly instead of
relying on training data or default web search.

Before writing it, interview me for: where to store my Perplexity API key (env file)
and which Perplexity model to default to (suggest one).

The skill must include: (1) trigger conditions — any question about recent or
fast-moving information, and any time my claim or the agent's knowledge might be
stale; (2) a working curl example for the Perplexity chat completions endpoint that
reads the key from the env file; (3) a rule to cite dates and primary sources in
answers built from search results; (4) a rule that when search results contradict the
model's training data, the search results win.

After writing it, test it by asking yourself one question about something released in
the last month, run the search, and show me the answer with sources.
  </task>
</prompt>
```
