---
name: setup
description: Set up the local dev environment for this project — dependencies, .env, git hooks, migrations, and test verification.
---

Step-by-step local dev environment setup for this project.

Usage: /setup

## Steps

1. **Check PHP version:**
   Run: `php -v`
   Required: PHP 8.5+. If lower → warn and stop.

2. **Install dependencies:**
   Run: `composer install --prefer-dist --optimize-autoloader`

3. **Environment file:**
   - Check if `.env` exists
   - If not → `cp .env.example .env` then `php artisan key:generate`
   - If yes → skip (don't overwrite)

4. **Activate git hooks:**
   Run: `git config core.hooksPath .githooks`
   Verify: `git config core.hooksPath` → should output `.githooks`
   This enables the pre-commit Pint linter hook.

5. **Run database migrations:**
   Run: `php artisan migrate`

6. **Verify queue config:**
   - Check `QUEUE_CONNECTION` in `.env`
   - For local dev: `QUEUE_CONNECTION=sync` is fine (no RabbitMQ needed)
   - For full testing: `QUEUE_CONNECTION=database` with migrations applied

7. **Run tests to verify setup:**
   Run: `./vendor/bin/phpunit`
   Expected: all pass (tests use `array` cache + `sync` queue, no external deps)

8. **Output summary:**
   ```
   ✅ PHP 8.5.x
   ✅ Dependencies installed
   ✅ .env configured
   ✅ Git hooks active (.githooks/pre-commit)
   ✅ Migrations applied
   ✅ Tests: N passed

   ⚠️ Manual steps still needed:
   - Configure HSM credentials in .env (FIUU_HSM_API_HOST, etc.)
   - Configure Apple mTLS certs (see docs/apple_top_architecture.md)
   - Configure CP API socket host/port
   ```
