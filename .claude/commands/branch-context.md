Load current branch memory into context at session start.

## Steps
1. Run: git branch --show-current
2. Check if memory/branches/<current-branch-name>.md exists
   - If yes: read it
   - If no: note there is no branch memory for this feature area, offer to create the file

3. Also read memory/core/ files relevant to this branch's domain
   (infer from branch name and any open files in context)

4. Output a concise session brief:
```
🌿 Branch: <current-branch-name>
📋 Branch memory: <N> entries loaded
🧠 Relevant core memory: <list of loaded files>

Key context for this session:
- <most important fact 1>
- <most important fact 2>
- <most important fact 3>
```
