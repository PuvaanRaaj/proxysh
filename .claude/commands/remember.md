Save a specific learning to shared codebase memory right now.

Learning: $ARGUMENTS

## Steps
1. Read memory/core/ and memory/branches/<current-branch-name>.md
2. Determine which file this belongs in:
   - Timeless cross-cutting fact → memory/core/<relevant-file>.md
   - Branch-specific or sprint-specific → memory/branches/<current-branch-name>.md
3. Check for duplicates — skip if already captured

## Show proposed update
```
📝 <file> (+1)
+ <the learning>
```
Wait for: confirm / skip

## On confirmation
1. Write the entry

2. Line budget check — run: `wc -l CLAUDE.md`
   - If CLAUDE.md > 200 lines → run /compact-memory first before regenerating

3. Regenerate CLAUDE.md:
   - Preserve everything ABOVE `## Codebase Memory` exactly as-is
   - Replace `## Codebase Memory` section downward with fresh concat of memory/core/*.md

4. Regenerate .cursorrules from memory/core/ only (imperative rules, grouped by category)

5. Run: git add memory/ CLAUDE.md .cursorrules
6. Run: git commit -m "chore(memory): <one-line summary>"

## Memory Format Rules
- Facts only — no prose, no "we decided", no "it was found that"
- One line per fact
- Inline code for all paths, methods, variables, line numbers
- Strikethrough abandoned approaches: ~~old way~~ → reason
- No filler: "Note that", "Keep in mind", "It's important to" → delete
