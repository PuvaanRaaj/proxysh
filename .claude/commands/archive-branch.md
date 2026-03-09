Archive a branch's memory after merging. Run this as part of the merge checklist.

Branch: $ARGUMENTS (if empty, use current branch)

## Steps
1. Determine branch name from $ARGUMENTS or run: git branch --show-current
2. Check memory/branches/<branch-name>.md exists — abort if not found

3. Before archiving, extract any reusable facts:
   - Scan branch memory for entries that are cross-cutting and timeless (not branch-specific)
   - Only promote facts that would help any dev on any future branch
   - Show proposed promotions:
```
📤 Promote to memory/core/<relevant-file>.md:
+ <entry>

📦 Archive (branch-specific, not reusable):
- <entry>
```
Wait for: confirm / edit / skip

4. On confirmation:
   - Write promoted entries to memory/core/<relevant-file>.md
   - Move branch file: memory/branches/<branch-name>.md → memory/archive/<branch-name>.md
   - Add archived-on date as first line: `> Archived: <today's date>`

5. Line budget check — run: `wc -l CLAUDE.md`
   - If CLAUDE.md > 200 lines → run /compact-memory before regenerating

6. Regenerate CLAUDE.md:
   - Preserve everything ABOVE `## Codebase Memory` exactly as-is
   - Replace `## Codebase Memory` section downward with fresh concat of memory/core/*.md

7. Regenerate .cursorrules from memory/core/ only (imperative rules, grouped by category)

8. Run: git add memory/
9. Run: git commit -m "chore(memory): archive <branch-name>"
