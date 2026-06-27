# Memory, Sessions & Knowledge

How to take an agent from stateless (each request isolated) to stateful (remembers context).
Three layers: working memory (this conversation), long-term memory (across sessions), and
retrieved knowledge (RAG / live data). Add only what the task needs — memory has real overhead.

## Table of contents
- Stateless vs. stateful — when you need memory
- Working memory: manual message lists
- Working memory: Sessions
- Managing large conversations: sliding window & summarization
- Long-term memory: persistent message logs
- Long-term memory: structured recall (the scalable pattern)
- Knowledge: training vs. retrieved (RAG)

## Stateless vs. stateful — when you need memory

A stateless agent treats every request in isolation (like a plain API call). A stateful agent
carries information forward. Conversational agents (chatbots, assistants, anything with
follow-up questions like "how big is *it*?") **must** be stateful. One-shot, routine,
non-learning tasks are fine stateless — don't pay the memory overhead you don't need.

## Working memory: manual message lists

The most fundamental approach: pass a list of message items (each `{"role": ..., "content": ...}`)
instead of a bare string, and append to it each turn. Roles are `system`, `user`, `assistant`.

```python
from agents import Agent, Runner

agent = Agent(name="QA", instructions="Answer questions.")
messages = []
while True:
    q = input("You: ")
    messages.append({"role": "user", "content": q})
    result = Runner.run_sync(agent, messages)
    print("Agent:", result.final_output)
    messages = result.to_input_list()   # neat way to carry full history forward
```

`result.to_input_list()` returns the full conversation ready to pass back in. This mirrors how
chat UIs work and aligns with the OpenAI message spec. It's the foundation everything else
builds on, but you manage it by hand.

## Working memory: Sessions

The SDK's `Session` primitive automates history storage/recall — no manual `to_input_list()`.
Pass a `session=` to `Runner`; it tracks the whole conversation behind the scenes.

```python
from agents import Agent, Runner, SQLiteSession

agent = Agent(name="QA", instructions="Answer questions.")
session = SQLiteSession("user_42_convo_1")     # unique id per conversation thread
while True:
    q = input("You: ")
    result = Runner.run_sync(agent, q, session=session)
    print("Agent:", result.final_output)
```

Use a distinct `session_id` per user/thread to keep memory contexts isolated (e.g.
`f"{username}:{conversation_id}"`). Plain `SQLiteSession(id)` is in-memory (lost on restart);
add a `db_path` to persist (next section).

## Managing large conversations: sliding window & summarization

LLMs have finite context windows. Appending forever eventually overflows the window (the agent
fails) and always inflates cost/latency. Two strategies, often combined:

- **Sliding window** — keep only the most recent N messages (FIFO). Simplest and cheapest.
  Risk: drops important early facts (the user's name/goal) once they age out.
  ```python
  from collections import deque
  messages = deque(maxlen=5)   # keep last 5 messages
  # append user + assistant turns; pass list(messages) to Runner
  ```
- **Summarization** — when history exceeds a threshold, summarize the oldest N messages with an
  LLM call and replace them with the compact summary. Retains key facts/decisions over long
  conversations. Cost: an extra LLM call per summarization. Bridges short- and long-term memory.

In practice combine them: sliding window for routine churn, summarization to preserve older
salient context in compressed form.

## Long-term memory: persistent message logs

The simplest persistence: give `SQLiteSession` a `db_path` and it saves/loads the conversation
to a local SQLite file automatically across process restarts.

```python
session = SQLiteSession("first_session", db_path="messages.db")
# run, exit the program, rerun — the agent still remembers earlier turns
```

Seamless (no custom schema), but it inherits the same problems at scale: storing/loading full
logs gets inefficient, and the log can bloat past the context window. Storing every message
verbatim is also a dumb form of memory — usually you only want the *key* facts.

## Long-term memory: structured recall (the scalable pattern)

Instead of persisting whole transcripts, give the agent **tools** to save and load distilled
facts. The conversation stays clean; the agent consults long-term memory only when relevant.
This mirrors how humans remember (you log "my friend likes sushi," not the whole transcript).

```python
import json, os
from agents import Agent, Runner, function_tool

FILE = "memory.json"   # could just as easily be a database

@function_tool
def save_memory(memory_type: str, memory: str) -> str:
    """Save an important fact.
    Args:
        memory_type: one of 'user_profile', 'order_preferences', 'other'.
        memory: the fact to store.
    """
    data = json.load(open(FILE)) if os.path.exists(FILE) else {}
    data.setdefault(memory_type, []).append(memory)
    json.dump(data, open(FILE, "w"), indent=2)
    return f"Saved: {memory}"

@function_tool
def load_memory(memory_type: str) -> str:
    """Load stored facts of a given type."""
    data = json.load(open(FILE)) if os.path.exists(FILE) else {}
    return " | ".join(data.get(memory_type, []))

agent = Agent(
    name="Assistant",
    instructions=("Save a memory when you learn an important fact. "
                  "Load memory when asked about the user."),
    tools=[save_memory, load_memory],
)
```

Scalable, semantically precise, and avoids context bloat. To scale further (fuzzy, semantic
recall over many facts), store each memory as a vector embedding in a vector store and
semantically search it — at which point you're doing RAG over your memory store (see the
**rag-and-knowledge-graphs** skill). You can combine persistent logs *and* structured recall.

## Knowledge: training vs. retrieved (RAG)

- **Training knowledge** — baked into model weights. Fast and broad, but frozen at the cutoff,
  not proprietary, and hard to cite. Changing it = fine-tuning, which is expensive (often
  $10k+), inflexible (re-do on every base-model update), and risks knowledge-mixing/contradiction.
  Prefer prompting or retrieval for most needs. For domain answers, often instruct the agent to
  *ignore* training knowledge and use retrieval.
- **Retrieved knowledge (RAG)** — pulled in at runtime via a tool, grounded and citable. The
  loop: user asks → agent retrieves relevant data (API, DB, web, or vector store) → data is
  added to the prompt (augment) → model answers (generate). For structured data a DB/API tool
  is enough; for unstructured docs use embeddings + semantic search + a vector store.

The OpenAI Agents SDK offers hosted RAG via `FileSearchTool` (upload to a vector store, pass
its id) which automates ingestion + retrieval — see `tools-and-function-calling.md`.

RAG pitfalls to design around: ambiguous queries (which "return policy"?), no relevant chunk
found (agent may hallucinate — instruct it to say "not found"), and conflicting sources (agent
may silently pick one). For chunking, embedding choice, retrievers, reranking, and vector-DB
tuning, see the **rag-and-knowledge-graphs** skill — this skill treats RAG only as a tool an
agent calls.
