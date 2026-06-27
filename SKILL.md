---
name: link-agent-instructions
description: Maintain a single source-of-truth agent-instructions file and fan it out to every coding harness via symlinks, so CLAUDE.md, AGENTS.md, GEMINI.md, .github/copilot-instructions.md, Cursor, Windsurf, Cline, and Kiro all stay in sync from one file. Use this whenever the user wants one shared instructions/memory file across multiple AI coding tools, asks to symlink or sync CLAUDE.md / AGENTS.md / GEMINI.md / copilot-instructions, wants to avoid duplicating agent rules per tool, or sets up a repo to work with more than one coding agent. Trigger even if the user only says they want "one rules file for all my agents" without naming the tools.
---

# Link agent instructions across harnesses

Different coding agents read instructions from different files: Claude Code uses `CLAUDE.md`, Codex/opencode/Amp use `AGENTS.md`, Gemini CLI uses `GEMINI.md`, GitHub Copilot uses `.github/copilot-instructions.md`, and Cursor/Windsurf/Cline/Kiro use files inside their own rules directories. Maintaining the same rules in all of them by hand guarantees drift.

This skill keeps **one canonical file** and points every harness path at it, so editing the canonical file updates all harnesses at once. The `air agents` command (in the cross-platform Go CLI) does this deterministically and safely — it picks **symlinks on macOS/Linux and copies on Windows automatically** (`--mode auto`).

## Canonical file choice

Default the canonical file to **`AGENTS.md`** at the repo root, because it is the open cross-tool standard (under the Linux Foundation's Agentic AI Foundation) and the file most tools already read natively. Every other harness path becomes a symlink to it. If the user prefers a neutrally-named source file, pass `--canonical <file>` and `AGENTS.md` itself will be linked to it like the rest.

## Workflow

1. **Confirm scope.** Ask which harnesses the user actually uses, or default to the common set (Claude, Gemini, Copilot, Cursor, Windsurf). `--all` covers every known harness including legacy `.cursorrules` / `.windsurfrules`. Run `air agents list` to show the full table.
2. **Seed the canonical file.** If `AGENTS.md` does not exist, `air agents link` creates it from `templates/AGENTS.template.md` (pass `--templates <dir>` if the templates live elsewhere). Edit that template (or the user's existing `AGENTS.md`) so it holds the real instructions before fanning out.
3. **Fan out.** Run the script (see Usage). It is idempotent, so it is safe to rerun any time the harness list changes.
4. **Verify.** Confirm each link resolves to the canonical content (`head -1 CLAUDE.md` should match `AGENTS.md`).
5. **Commit.** Commit the canonical file and the symlinks together.

## Usage

```
# default harnesses, AGENTS.md as canonical (seeds it if missing)
air agents link

# every known harness
air agents link --all

# pick specific harnesses
air agents link --harness claude,gemini,copilot

# use a neutrally-named source file
air agents link --canonical AGENT_GUIDE.md --all

# preview without changing anything
air agents link --all --dry-run

# force real copies instead of symlinks (auto on Windows)
air agents link --all --mode copy

# remove the managed links (restores any backed-up originals)
air agents unlink --all
```

## Harness paths

| Harness | File it reads |
| --- | --- |
| Codex / opencode / Amp / the standard | `AGENTS.md` (the canonical file) |
| Claude Code | `CLAUDE.md` |
| Gemini CLI | `GEMINI.md` |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Cursor | `.cursor/rules/agents.mdc` |
| Windsurf | `.windsurf/rules/agents.md` |
| Cline | `.clinerules/agents.md` |
| Kiro | `.kiro/steering/agents.md` |
| Cursor (legacy) | `.cursorrules` |
| Windsurf (legacy) | `.windsurfrules` |

Most of these tools also read `AGENTS.md` natively now, so the symlinks are belt-and-suspenders: they guarantee the rules apply even for a tool's own preferred filename and for older versions.

## Important caveats

**Symlinks vs. Windows / CI.** Git tracks symlinks on Unix, but Windows checkouts (without developer mode) and some CI or archive steps do not preserve them. `air agents link` already chooses **copy** automatically on Windows (`--mode auto`); force it anywhere with `--mode copy`. Rerun after editing the canonical file to re-sync copies — the trade-off is that copies are not auto-updated the way symlinks are.

**Rule-directory harnesses need frontmatter for "always apply."** Cursor (`.mdc`) and Windsurf rule files support frontmatter that controls whether a rule is always on, manually invoked, or model-decided. A raw symlink gives a basic, always-available rule but not guaranteed "always apply" semantics. If the user needs that, create a small wrapper file with the harness's frontmatter that references or imports `AGENTS.md`, instead of a raw symlink. `air agents` prints a note when it links these.

**Do not symlink referenced sidecar files.** The canonical content may point to other files (for example `opinions.md` and `voice.md`, read conditionally). Those are referenced on demand, not fanned out - leave them as ordinary files next to the canonical file. Run `air agents link --stubs` to create them from the bundled templates so the references resolve, then populate them (see the next section).

**Never lose existing content.** If a real file already exists at a target path, `air agents` moves it to `<path>.bak` before linking and `air agents unlink` restores it. It never overwrites a real file in place.

## Bootstrapping voice.md and opinions.md from your own writing

When the canonical instructions reference a voice profile (`voice.md`) and viewpoints (`opinions.md`), this skill can generate a first draft of those files from the user's own writing, on request. Only do this when the user explicitly asks; never harvest someone's communications unprompted.

First create the files with `air agents link --stubs` so the skeletons exist, then fill them in using the source chain below.

### Source chain (use the first that is available)

1. **Microsoft Work IQ (`workiq`).** If the user has Work IQ - Microsoft's CLI/MCP layer over Microsoft 365 (Outlook email, Teams messages, meetings, documents) - use it to pull samples of the user's *own authored* messages. In CLI mode the command is `workiq ask -q "..."`; in MCP mode call its query tools. Ask specifically for content the user wrote, e.g. emails they sent and Teams messages they posted, not messages they received, since voice must reflect how the user writes. Pull a spread across audiences (a teammate, a wider group, an external recipient) so the profile captures register shifts.
2. **Any other connected mail/chat source.** If Work IQ is not present but another connector is (Gmail, an Outlook MCP, Slack), use it the same way: gather a sample of the user's *sent* messages.
3. **Interview / direct input.** If no connector is available, ask the user directly. Have them paste 3-6 representative messages or emails they have written, or answer a short set of questions (how formal, how direct, sign-offs, humor, hard rules, positions they hold). Prompts are a perfectly good source; do not block on connectors.

### Synthesis

Produce two distinct artifacts, following the bundled templates:

- **`voice.md` - how they write.** Characterize register, directness, sentence rhythm, greetings and sign-offs, warmth and humor, signature vocabulary, hard rules (e.g. no em dash), and what to avoid. Include only a few short, abstracted before/after examples.
- **`opinions.md` - what they consistently think.** Extract recurring positions and strong defaults (engineering principles, product and process views, things they push back on, tradeoff leanings). Capture patterns that show up repeatedly, not one-off remarks.

### Privacy and safety (important)

- Operate only on the user's own data, by their explicit request.
- **Abstract, do not paste.** These files get symlinked across the repo and likely committed. Describe style and summarize positions; do not copy large verbatim private messages, and never include other people's private content, confidential specifics, names, or sensitive business details.
- Always present the generated `voice.md` and `opinions.md` as drafts for the user to review and edit before saving. Never auto-commit them.
- Treat the result as a starting point that the user refines over time.



This skill targets project-level files at the repo root. For machine-wide defaults, the same idea applies to global paths (for example `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.gemini/GEMINI.md`), but those are per-tool home directories rather than a single repo, so link them individually to one global canonical file if desired.
