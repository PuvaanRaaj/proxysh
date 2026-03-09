---
name: review
description: Pre-commit review for this PCI-DSS payment codebase. Checks code quality, project gotchas, AND PCI-DSS compliance in one pass. Use before every commit or when reviewing any file that touches payment logic, card data, PIN, HSM, or logging.
---

Pre-commit review covering code quality + PCI-DSS compliance in one pass.

Usage: /review [file-path]
  - /review              → reviews git diff --staged
  - /review path/to/File.php → reviews a specific file

## Step 1 — Get the code
- If $ARGUMENTS is a file path → Read that file
- If blank → Run: `git diff --staged`
- If no staged changes → Run: `git diff HEAD~1`

## Step 2 — Load project context
- Read /memory/core/gotchas.md
- Read /memory/branches/<current-branch-name>.md

## Step 3 — Code Quality Checks
- Known gotchas from memory that apply to this change
- Missing tests for changed logic
- Hardcoded values that should be `config('services.*')` or `env()`
- Error handling gaps (missing `catch` on socket/HSM/external calls)
- Anything that looks like a previously abandoned approach

## Step 4 — PCI-DSS Checks
Run through each. Only flag real issues — skip checks that don't apply.

### Cardholder Data (CRITICAL)
- PAN, CVV, PIN, track data stored anywhere after authorization → ❌ CRITICAL
- Raw card data returned in JSON responses or error messages → ❌ CRITICAL
- Card numbers not masked (must be first 6 + `****` + last 4) → ❌ HIGH
- `dd()`, `var_dump()`, `dump()` near card/PIN variables → ❌ HIGH

### Logging
Grep all `Log::`, `logger()`, `error_log()` calls.
- Log contains PAN, CVV, PIN, expiry, or full authorization response → ❌ HIGH
- `returnValue` from HSM logged without `maskSensitiveData()` → ❌ HIGH
- `dd()` / `var_dump()` left in code → ❌ HIGH
- Sensitive keys missing from `maskSensitiveData()` call → ⚠️ MEDIUM

### Secrets & TLS
- Hard-coded API keys, certificates, or passwords → ❌ CRITICAL
- Credentials not from `config('services.*')` or `env()` → ❌ HIGH
- TLS cert verification disabled (`verify => false`, `CURLOPT_SSL_VERIFYPEER`) → ❌ CRITICAL
- Hard-coded HTTP (not HTTPS) for payment URLs → ❌ HIGH

### Input Validation
- Payment controllers not using `FormRequest` → ❌ HIGH
- Raw `$request->all()` passed to services without validation → ⚠️ MEDIUM

### Idempotency & Double-Charge
- Purchase/capture endpoints missing transaction ID cache check → ❌ HIGH
- Auto-reversal jobs that could fire multiple times on same transaction → ⚠️ MEDIUM

### Authorization
- Payment endpoints missing `CheckGeoLocationMiddleware` → ❌ CRITICAL

### Apple TTP (if applicable)
- `pinToken` treated as raw PIN → ❌ CRITICAL (must go through HSM M20/21)
- `AppleProximityService` response fields logged without masking → ⚠️ MEDIUM

## Output

```
## Review: <file or "staged changes">

### ❌ Must fix before commit
- [Type | Severity] file:line — what's wrong → how to fix

### ⚠️ Fix soon
- [Type | Severity] file:line — issue

### ✅ Passed
- (only list checks that clearly passed)

### 💡 Suggestions (optional)
- (PHP 8.4 / Laravel 12 improvements, non-blocking)
```

If any ❌ Critical found: stop and ask "Fix now or abort commit?"
If only ⚠️ or suggestions: ask "Fix now, commit anyway, or abort?"
