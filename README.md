# SplitPay

**SplitPay is a collaborative expense planning and settlement demo.** A group can agree on costs before they affect balances, discuss a proposed expense, plan recurring bills, and follow a clear settlement history. The application is built with React, TypeScript, Go, and MongoDB.

> The application source is kept in a private development workspace. This GitHub repository publishes the project overview only.

## Why it goes beyond a basic expense splitter

- **Group decisions before debt:** A new expense is proposed first. Each included participant can approve or reject their share; only an approved expense changes balances. Rejected proposals can be edited and resubmitted with their decision history intact.
- **Discussion in context:** Members can comment on an expense or a shared purchase plan instead of using a separate group chat.
- **Plan Together:** Members compare priced purchase options, vote, and discuss them. The selected option can become an expense proposal through the same approval flow.
- **Recurring cost radar:** Monthly schedules show upcoming due dates, estimated shares, and a 30-day forecast. A new bill can be proposed for approval, with an explanation when its amount differs substantially from recent bills or its original estimate.
- **Budget outlook:** A monthly group target compares approved spending, pending proposals, and recurring estimates. A what-if input previews another cost without creating an expense.
- **Settlement history:** The app calculates balances in integer paise, suggests settlements, records partial or full demo payments, and shows a timeline of attempts and completed settlements.

## Collaboration and account experience

Group members have account-linked slots. An owner can create a seven-day invite link for an unclaimed slot, copy it, or open a prefilled Gmail draft to send it. **Gmail does not send automatically:** the owner reviews the draft and clicks Send. Anyone holding a valid invite link can claim its slot, so it should be shared only with the intended person.

Clerk supports email verification codes for sign-up and sign-in. The local project also retains Firebase as a rollback path for existing accounts. Expense activity, payment updates, and reminders appear in an in-app notification view opened from a bell. Members can opt in to settlement reminders; each payer-recipient pair has a 24-hour reminder cooldown.

## Payment demo boundaries

The **Simulated payment** method lets a reviewer try success, failure, and cancellation without a payment provider. An optional **Razorpay Test Mode** checkout creates a test order, verifies the signed checkout response and captured payment on the server, and can recover an interrupted checkout through a signed webhook or a status check. A captured test payment is held for review if the underlying group balance changed before it was recorded.

**Neither method transfers real money to another group member.** Razorpay Test Mode needs test keys and dashboard configuration. Live collection and person-to-person payouts are outside the current product scope.

## Technical approach

| Area | Implementation |
| --- | --- |
| Frontend | React 19, TypeScript, Create React App with CRACO, Tailwind CSS |
| Backend | Go 1.24, Fiber, MongoDB |
| Identity | Clerk email-code sessions; Firebase rollback and migration tooling |
| Money | Integer paise, server-calculated balances, deterministic equal-share rounding |
| Reliability | Balance revision checks, idempotent payment completion, Go tests, frontend production build, CI |

## Current status

The group, approval, discussion, planning, budget, notification, and demo settlement flows are implemented in the local project. Backend tests and the frontend production build pass. Live multi-account review, Clerk migration of an existing installation, and Razorpay Test Mode checkout still depend on the relevant service configuration and test accounts.

The product currently uses equal shares among selected participants. Recurring schedules are monthly. Existing expenses stored as floating-point values remain readable, but a database migration is still needed. Guaranteed delivery of every expense notification would require a transactional outbox or equivalent mechanism.
