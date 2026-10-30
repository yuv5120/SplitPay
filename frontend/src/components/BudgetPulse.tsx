import { useState } from 'react';
import { api, Group } from '../lib/api';
import { parseRupeesToPaise } from '../lib/money';
import { formatCurrency } from '../lib/utils';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { toast } from 'sonner';

export function BudgetPulse({ group, onChanged }: { group: Group; onChanged: () => Promise<void> }) {
  const budget = group.budget;
  const [open, setOpen] = useState(false);
  const [amount, setAmount] = useState('');
  const [scenario, setScenario] = useState('');
  const [busy, setBusy] = useState(false);
  const currency = (paise: number) => formatCurrency(paise / 100);
  const openEditor = () => { setAmount(budget.limitPaise ? (budget.limitPaise / 100).toFixed(2) : ''); setOpen(true); };
  const save = async (remove = false) => {
    const paise = remove ? 0 : parseRupeesToPaise(amount);
    if (paise === null) { toast.error('Enter a valid monthly target'); return; }
    setBusy(true);
    try { await api.updateBudget(group.id, paise); await onChanged(); setOpen(false); toast.success(remove ? 'Monthly target removed' : 'Monthly target saved'); }
    catch { toast.error('Could not update monthly target'); }
    finally { setBusy(false); }
  };
  const ratio = budget.limitPaise > 0 ? Math.min(100, Math.round(budget.projectedPaise / budget.limitPaise * 100)) : 0;
  const scenarioPaise = scenario.trim() ? parseRupeesToPaise(scenario) : null;
  const scenarioTotal = scenarioPaise === null ? null : budget.projectedPaise + scenarioPaise;
  const month = new Date(`${budget.period}-01T00:00:00+05:30`).toLocaleString('en-IN', { month: 'long', year: 'numeric', timeZone: 'Asia/Kolkata' });

  return <>
    <Card className="glass-card mb-8">
      <CardHeader><div className="flex items-start justify-between gap-4"><div><CardTitle>Monthly budget outlook</CardTitle><p className="text-sm text-muted-foreground mt-1">{month} · group spending plan</p></div>{group.canManage && <Button variant="outline" onClick={openEditor}>{budget.limitPaise ? 'Edit target' : 'Set target'}</Button>}</div></CardHeader>
      <CardContent className="space-y-4">
        {budget.limitPaise > 0 ? <><div className="flex flex-wrap items-baseline justify-between gap-2"><p className="text-2xl font-bold">{currency(budget.projectedPaise)} <span className="text-sm font-normal text-muted-foreground">projected of {currency(budget.limitPaise)}</span></p><p className={budget.overBudgetPaise ? 'text-amber-700 font-medium' : 'text-green-700 font-medium'}>{budget.overBudgetPaise ? `${currency(budget.overBudgetPaise)} over target` : `${currency(budget.remainingPaise)} left`}</p></div><div className="h-3 rounded-full bg-muted overflow-hidden" role="progressbar" aria-label="Projected monthly spending" aria-valuenow={ratio} aria-valuemin={0} aria-valuemax={100}><div className={`h-full rounded-full ${budget.overBudgetPaise ? 'bg-amber-500' : 'bg-primary'}`} style={{ width: `${ratio}%` }} /></div></> : <p className="text-sm text-muted-foreground">{group.canManage ? 'Set a monthly target to see whether approved bills, proposals, and scheduled costs fit.' : 'The group owner has not set a monthly target yet.'}</p>}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm"><div className="rounded-lg border p-3"><p className="text-muted-foreground">Approved</p><strong>{currency(budget.approvedPaise)}</strong></div><div className="rounded-lg border p-3"><p className="text-muted-foreground">Awaiting approval</p><strong>{currency(budget.pendingPaise)}</strong></div><div className="rounded-lg border p-3"><p className="text-muted-foreground">Scheduled estimate</p><strong>{currency(budget.scheduledPaise)}</strong></div></div>
        <div className="rounded-lg border p-4 space-y-2"><Label htmlFor="budget-what-if">What if we add another group cost?</Label><div className="flex flex-wrap items-center gap-3"><Input id="budget-what-if" inputMode="decimal" className="max-w-48" value={scenario} onChange={event => setScenario(event.target.value)} placeholder="Amount in ₹" />{scenarioTotal !== null && <p className="text-sm" aria-live="polite">Projected: <strong>{currency(scenarioTotal)}</strong>{budget.limitPaise > 0 && <> · {scenarioTotal > budget.limitPaise ? `${currency(scenarioTotal - budget.limitPaise)} over target` : `${currency(budget.limitPaise - scenarioTotal)} left`}</>}</p>}</div><p className="text-xs text-muted-foreground">Try an amount without creating a bill or changing anyone’s balance.</p></div>
        <p className="text-xs text-muted-foreground">Planning only. Pending and scheduled costs are estimates; they do not change balances. A recurring bill replaces its schedule estimate once proposed.</p>
      </CardContent>
    </Card>
    <Dialog open={open} onOpenChange={setOpen}><DialogContent onClose={() => setOpen(false)}><DialogHeader><DialogTitle>Monthly group target</DialogTitle></DialogHeader><div className="space-y-4 mt-3"><p className="text-sm text-muted-foreground">A shared planning target for this calendar month and future months. Only the group owner can change it.</p><div><Label htmlFor="budget-target">Target amount (₹)</Label><Input id="budget-target" inputMode="decimal" value={amount} onChange={event => setAmount(event.target.value)} placeholder="20000.00" /></div><div className="flex gap-2"><Button disabled={busy} onClick={() => save()}>Save target</Button>{budget.limitPaise > 0 && <Button disabled={busy} variant="outline" onClick={() => save(true)}>Remove target</Button>}</div></div></DialogContent></Dialog>
  </>;
}
