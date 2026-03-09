---
name: explain
description: Explain what a file or module does using memory and code context. Use when asking "what does this do", "how does this work", or exploring an unfamiliar file.
---

Explain what a file or module does using memory and code context.

Target: $ARGUMENTS (file path or directory)

## Steps
1. Read the target file(s)
2. Cross-reference with /memory/core/services.md and /memory/core/domain.md

3. Output:
```
## <file/module name>

### Purpose
<what this is responsible for, in one sentence>

### Responsibilities
- <responsibility 1>
- <responsibility 2>

### Key Methods/Classes
| Name | What it does |
|------|-------------|
| <method> | <description> |

### Dependencies
- Calls: <other services/files it depends on>
- Called by: <what uses this>

### Gotchas
- <non-obvious thing about this file from memory>
```
