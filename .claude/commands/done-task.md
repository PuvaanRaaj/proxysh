Extract learnings from this session and persist them to shared codebase memory.

## Step 1 — Summarize what happened
Review the conversation and changes. Identify:
- What changed and WHY (not just what)
- Non-obvious decisions and reasoning
- Gotchas, bugs, edge cases found
- Anything that took more than 1 attempt
- Domain knowledge that surfaced (business rules, external system behaviors)
- Patterns introduced or reinforced

## Step 2 — Read existing memory
Read all files in the repo-level `memory/` folder (NOT ~/.claude/...) to avoid duplicating existing entries.
Use Glob with pattern `memory/**/*.md` to find all memory files, then Read each one.
Determine which files need updates:
- `memory/core/gotchas.md` — cross-cutting bugs, constraints, system limits
- `memory/core/domain.md` — business rules, external system behaviors
- `memory/core/conventions.md` — naming, structure, testing patterns
- `memory/core/services.md` — service boundaries, contracts, config
- `memory/branches/<current-branch-name>.md` — branch-specific context

## Step 3 — Show proposed updates
Show exactly what you're adding using this format:

```
📝 memory/core/<relevant-file>.md (+<N>)
+ <learning one>
+ <learning two>

📝 memory/branches/<current-branch-name>.md (<new|+N>)
+ <learning one>
```

Wait for: confirm / edit / skip

## Step 4 — Write and commit
After confirmation:
1. Write updated files into the repo-level `memory/` folder using absolute paths (e.g. `<repo-root>/memory/core/gotchas.md`). Do NOT write to `~/.claude/projects/...`.

2. Line budget check — run: `wc -l CLAUDE.md`
   - If CLAUDE.md is currently > 200 lines → run /compact-memory on memory/core/ first before regenerating
   - Goal: keep CLAUDE.md under 200 lines total

3. Regenerate CLAUDE.md:
   - Preserve everything ABOVE `## Codebase Memory` exactly as-is
   - Replace `## Codebase Memory` section downward with:
     ```
     ## Codebase Memory
     > Auto-generated — edit /memory/*.md, not this file.
     > Last synced: <today's date>

     <concat all /memory/core/*.md content, one section per file>
     ```

4. Regenerate .cursorrules from /memory/core/ only:
   - Imperative rules only ("Always", "Never", "Use", "Avoid")
   - Grouped by category, no prose, no explanations
   - Branch memory excluded

5. Run: git add memory/ CLAUDE.md .cursorrules
6. Run: git commit -m "chore(memory): session learnings"

## Memory Format Rules
- Facts only — no prose, no "we decided", no "it was found that"
- One line per fact
- Inline code for all paths, methods, variables, line numbers
- Tables over paragraphs for 3+ related file/config listings
- Strikethrough abandoned approaches: ~~old way~~ → reason
- No filler: "Note that", "Keep in mind", "It's important to" → delete
- Branch-specific → memory/branches/<current-branch-name>.md
- Reusable cross-cutting facts → memory/core/

## What belongs in core/ (timeless only)
✅ Keep in core/: architectural decisions, domain rules, gotchas that bite any dev at any time
❌ Not in core/: sprint-specific notes, file-level implementation details, anything true for only one branch
→ If unsure: put it in branches/ — it can be promoted to core/ via /archive-branch after merging
