---
name: trace
description: Trace the full request flow for a payment scheme (MYDEBIT, VISA, AMEX, etc.) through the codebase. Use when asking how a purchase works, debugging a flow, or onboarding to a scheme.
---

Trace the full request flow for a payment scheme through the codebase.

Usage: /trace [scheme] [operation]
  - /trace MYDEBIT purchase
  - /trace VISA reversal
  - /trace AMEX
  - /trace (shows all schemes)

Schemes: MYDEBIT, VISA, MASTERCARD, UPI, DCI, AMEX
Operations: purchase, reversal-purchase, cancellation, reversal-cancellation-purchase

## Steps

1. Identify the scheme and operation from $ARGUMENTS (default: MYDEBIT + purchase).

2. Read these files in order:
   - `routes/api.v1.php` — find the route
   - The matching controller (e.g. `app/Http/Controllers/V1/AP/PurchaseApiController.php`)
   - The matching FormRequest
   - The scheme service (from the `$services` map in the controller)

3. Trace each stage and output as a numbered flow:

```
## Flow: <SCHEME> <OPERATION>

Route:    POST /v1/ap/<operation>
Request:  <FormRequest class>

1. Middleware
   → SanitizeJsonMiddleware
   → CheckGeoLocationMiddleware (validates merID, country, API key)

2. Controller: <ControllerClass>::<method>()
   → Checksum validation (HandlesChecksumValidation trait)
   → [Apple only] AppleProximityService::translate() — mTLS to Apple
   → [Apple only] Extract pinToken / transactionId if present

3. HSM Phase 1 (parallel)
   → GuzzleService M17 — decrypt card session key (cardholderDataKey)
   → GuzzleService M18 — decrypt PIN session key (pinKey)  [if PIN present]

4. HSM Phase 2
   → GuzzleService M19 — decrypt card data → parse EMV TLV → extract PAN (tag 57)
   → GuzzleService M20/M21 — translate PIN block (M20=non-MyDebit, M21=MyDebit)

5. Scheme Processing
   → <SchemeService>::process() — build ISO8583 bit 55 + bitmap

6. Socket
   → SocketClientService::sendData() — TCP to CP API
   → Auto-reversal job dispatched on failure

7. Response
   → Iso8583Service::parseISO() — parse response
   → Cache transaction data (encrypted, 1 day)
   → PurchaseResource — format response
      [+ pinToken, transactionId if Apple TTP PIN flow]

Key files:
| Stage | File |
|-------|------|
| Route | routes/api.v1.php |
| Controller | <path> |
| Scheme service | <path> |
| Request | <path> |
```

4. Highlight any non-obvious gotchas for the requested scheme:
   - MYDEBIT → uses M21 (not M20) for PIN; slot = `BTT_`
   - Apple TTP → pinToken must go through HSM before PIN block is available
   - Multi-region → country code validated via `CountryCurrencyValidationService`

5. If $ARGUMENTS is empty, show a comparison table of all schemes side by side.
