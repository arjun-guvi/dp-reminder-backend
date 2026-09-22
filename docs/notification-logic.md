# Notification Logic Explained

## When Are Notifications Sent?

Notifications are sent automatically by the worker based on specific conditions.

---

## Trigger Conditions (ALL must be true)

### 1. Worker Schedule
The worker runs every **30 minutes** (configurable via `NOTIFICATION_INTERVAL_MINUTES`)

```
┌─────────────────────────────────────────────────────────────┐
│  Worker runs every 30 minutes (00:00, 00:30, 01:00, ...)    │
└─────────────────────────────────────────────────────────────┘
```

### 2. Payment Must Match These Conditions

A notification is sent for a payment if ALL of the following are true:

| Condition | Description |
|-----------|-------------|
| `status = "pending"` | Payment is not yet paid or cancelled |
| `notification_sent_at` does NOT exist | Notification hasn't been sent yet |
| `due_date <= window_end` | Payment is due within the reminder window |

### 3. Reminder Window Calculation

```
window_end = current_time + PAYMENT_REMINDER_DAYS

Example with PAYMENT_REMINDER_DAYS=1:
┌────────────────────────────────────────────────────────────────────┐
│  Current Time: Sep 22, 2024 10:00 AM                                │
│  Window End:   Sep 23, 2024 10:00 AM (1 day ahead)                  │
│                                                                     │
│  Payments with due_date <= Sep 23, 10:00 AM will be notified        │
└────────────────────────────────────────────────────────────────────┘
```

---

## Visual Timeline

```
Today: Sep 22                     Tomorrow: Sep 23              Sep 24
    │                                  │                            │
    ├─── Worker Run (10:00 AM) ────────┼────────────────────────────┤
    │   Checks payments due by:        │                            │
    │   Sep 23, 10:00 AM               │                            │
    │                                  │                            │
    │   ┌─────────────────┐            │                            │
    │   │ Payment A       │            │                            │
    │   │ Due: Sep 22     │ ◄── SEND   │  (due today)               │
    │   │ Status: pending │            │                            │
    │   │ notification_   │            │                            │
    │   │ sent_at: null   │            │                            │
    │   └─────────────────┘            │                            │
    │                                  │                            │
    │   ┌─────────────────┐            │                            │
    │   │ Payment B       │            │                            │
    │   │ Due: Sep 23     │ ───────────┤                            │
    │   │ Status: pending │ ◄── SEND   │  (due tomorrow,            │
    │   │ notification_   │            │   within 1-day window)     │
    │   │ sent_at: null   │            │                            │
    │   └─────────────────┘            │                            │
    │                                  │                            │
    │   ┌─────────────────┐            │                            │
    │   │ Payment C       │ ───────────┼────────────────────────────┤
    │   │ Due: Sep 24     │            │                            │
    │   │ Status: pending │            │ ◄── NOT SENT               │
    │   │ notification_   │            │    (outside window)         │
    │   │ sent_at: null   │            │                            │
    │   └─────────────────┘            │                            │
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NOTIFICATION_INTERVAL_MINUTES` | 30 | How often the worker checks for payments |
| `PAYMENT_REMINDER_DAYS` | 1 | Days before due date to send reminder |

### Setting in `.env`

```bash
# Check for payments every 30 minutes
NOTIFICATION_INTERVAL_MINUTES=30

# Send reminder 1 day before due date
PAYMENT_REMINDER_DAYS=1

# Or send reminder 3 days before due date
PAYMENT_REMINDER_DAYS=3
```

---

## Examples

### Example 1: Payment due today
```
Current: Sep 22, 10:00 AM
PAYMENT_REMINDER_DAYS: 1
Window: Sep 23, 10:00 AM

Payment:
  - Due: Sep 22, 3:00 PM
  - Status: pending
  - notification_sent_at: null

Result: ✅ NOTIFICATION SENT (due_date is within window)
```

### Example 2: Payment due tomorrow
```
Current: Sep 22, 10:00 AM
PAYMENT_REMINDER_DAYS: 1
Window: Sep 23, 10:00 AM

Payment:
  - Due: Sep 23, 9:00 AM
  - Status: pending
  - notification_sent_at: null

Result: ✅ NOTIFICATION SENT (due_date is within window)
```

### Example 3: Payment due in 2 days
```
Current: Sep 22, 10:00 AM
PAYMENT_REMINDER_DAYS: 1
Window: Sep 23, 10:00 AM

Payment:
  - Due: Sep 24, 10:00 AM
  - Status: pending
  - notification_sent_at: null

Result: ❌ NOT SENT (due_date > window_end)
```

### Example 4: Already notified
```
Payment:
  - Due: Sep 22, 3:00 PM
  - Status: pending
  - notification_sent_at: Sep 21, 10:00 AM (already exists!)

Result: ❌ NOT SENT (notification_sent_at already exists)
```

### Example 5: Payment already paid
```
Payment:
  - Due: Sep 22, 3:00 PM
  - Status: paid
  - notification_sent_at: null

Result: ❌ NOT SENT (status is not "pending")
```

---

## After Notification is Sent

Once a notification is sent successfully:

1. The payment's `notification_sent_at` field is set to current timestamp
2. Future worker runs skip this payment
3. The notification won't be sent again (unless you use `force: true`)

---

## Testing

### Check which payments would be notified right now:
```bash
curl -X POST http://localhost:8080/notifications/test \
  -H "Content-Type: application/json" \
  -d '{"dry_run": true}'
```

### Force re-notification (for testing):
```bash
curl -X POST http://localhost:8080/notifications/test \
  -H "Content-Type: application/json" \
  -d '{"dry_run": true, "force": true}'
```

---

## Summary

| Your Config | Payment Due Date | When Notification is Sent |
|-------------|-------------------|---------------------------|
| `PAYMENT_REMINDER_DAYS=1` | Sep 24 | Sep 22 or Sep 23 (1 day before) |
| `PAYMENT_REMINDER_DAYS=3` | Sep 24 | Sep 21, 22, or 23 (3 days before) |

**The worker checks every 30 minutes for any payments that fall within the reminder window.**
