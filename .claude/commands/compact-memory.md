Compress and deduplicate all memory files in /memory/. Run when CLAUDE.md exceeds 200 lines or memory feels stale.

## When to run
- CLAUDE.md > 200 lines
- `/done-task` triggers it automatically when over budget
- Manually when memory feels bloated or outdated

## Steps

1. Run: `wc -l CLAUDE.md memory/core/*.md` — note current sizes

2. Read all files in /memory/core/ and /memory/branches/

3. For each core/ file apply compression:
   - Merge duplicate or near-duplicate entries into one line
   - Convert prose to single-line facts
   - Use tables where 3+ related items exist
   - Remove filler: "Note that", "Keep in mind", "It's important to"
   - Remove strikethrough entries that are no longer relevant

4. Prune what does NOT belong in core/:
   - Sprint-specific or branch-specific notes → move to branches/ or delete
   - File-level implementation details → delete (put in skills instead)
   - Facts that are only true for one feature → delete or move to branch file
   - Anything already enforced by a skill (e.g. /pci-review covers PCI rules) → delete from core/

5. Cross-file deduplication:
   - Same fact in multiple files → keep in most relevant file only
   - Branch file entry that is reusable cross-cutting knowledge → move to /memory/core/

6. Show before/after estimate:
```
memory/core/conventions.md   <N> lines → <N> lines (-<N>%)
memory/core/gotchas.md       <N> lines → <N> lines (-<N>%)
memory/core/domain.md        <N> lines → <N> lines (-<N>%)
CLAUDE.md (projected)        <N> lines → <N> lines

Target: CLAUDE.md < 200 lines total
```
Wait for: confirm / skip

7. On confirmation:
   - Write all updated memory/core/ files
   - Regenerate CLAUDE.md (preserve top section, replace ## Codebase Memory downward)
   - Regenerate .cursorrules (imperative rules from core/ only, grouped by category)
   - Run: git add memory/ CLAUDE.md .cursorrules
   - Run: git commit -m "chore(memory): compact and deduplicate"
