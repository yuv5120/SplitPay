<div align="center">

# SplitPay

### Plan together. Split fairly. Settle clearly.

SplitPay brings shared expenses, group decisions, and settlement tracking into one place. It helps roommates, friends, and travel groups agree on costs before they become debts.

![React](https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?logo=typescript&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-47A248?logo=mongodb&logoColor=white)

</div>

## ✨ Features

### Shared groups and invitations

Create a group, add member slots, and invite people with a seven-day link. Invitations can be copied or opened as a prefilled Gmail draft. Each member signs in and joins the group through their link.

### Expense approval and discussion

Choose who paid and who participated in an expense. The expense becomes part of the balance only after the included members approve it. Members can reject a proposal, discuss it in comments, and review its decision history when it is revised.

### Balances and settlements

SplitPay calculates equal shares in exact paise, shows who owes whom, and suggests settlements. Members can record full or partial demo payments and follow every attempt in a settlement timeline. The checkout supports a simulated method and optional Razorpay Test Mode; test payments do not transfer money to group members.

### Recurring cost radar

Schedule monthly costs such as rent or utilities. See upcoming due dates, estimated shares, and a 30-day forecast. When an actual bill differs significantly from its estimate or recent bills, SplitPay explains the change before the group approves it.

### Plan Together

Compare priced options for a shared purchase, vote, and discuss the choice with the group. The selected option can be turned into an expense proposal, so the normal approval flow still applies.

### Budget outlook

Set a monthly group target and compare it with approved costs, pending proposals, and recurring estimates. Try a possible new cost in a what-if field to preview its impact without creating an expense.

### Notifications and reminders

A bell opens group notifications for expense decisions, comments, and payments. Members can opt in to settlement reminders, with a 24-hour cooldown for each payer-recipient pair.

### Account and profile

Sign up or sign in with an email verification code using Clerk. Members can view and update their profile details and manage their session from one profile page.

## 🧭 How it works

1. Create a group and invite its members.
2. Propose an expense and discuss it with the people included.
3. Members approve their shares; the group balance updates.
4. Review settlement suggestions, record a demo payment, and follow the timeline.
5. Plan upcoming bills and shared purchases before they affect the balance.

## 🛠 Tech stack

| Layer | Technologies |
| --- | --- |
| Frontend | React 19, TypeScript, Tailwind CSS, CRACO |
| Backend | Go 1.24, Fiber |
| Database | MongoDB |
| Authentication | Clerk email-code sessions; Firebase retained for account migration and rollback |
| Payment demo | Simulated checkout and Razorpay Test Mode |
| Quality | Go tests and frontend production build |

## 🎨 Design

The interface uses responsive cards, compact member and expense views, and clear status labels for proposals, approvals, and payments. A group dashboard brings budgets, recurring costs, purchase plans, balances, and the settlement timeline together while keeping each action close to its context.
