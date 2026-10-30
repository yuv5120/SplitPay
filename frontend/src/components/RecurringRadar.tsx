import { useState } from 'react';
import { api, Group, RecurringCost, RecurringInput } from '../lib/api';
import { formatCurrency, formatDate } from '../lib/utils';
import { parseRupeesToPaise } from '../lib/money';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Checkbox } from './ui/checkbox';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { toast } from 'sonner';

interface Props {
  group: Group;
  onChanged: () => Promise<void>;
}

type Form = { description: string; estimate: string; dueDay: string; paidBy: string; participants: string[] };
const emptyForm: Form = { description: '', estimate: '', dueDay: '1', paidBy: '', participants: [] };

export function RecurringRadar({ group, onChanged }: Props) {
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<RecurringCost | null>(null);
  const [form, setForm] = useState<Form>(emptyForm);
  const [proposing, setProposing] = useState<RecurringCost | null>(null);
  const [actualAmount, setActualAmount] = useState('');
  const [busy, setBusy] = useState(false);
  const joined = group.members.filter(member => member.isJoined);
  const name = (id: string) => group.members.find(member => member.id === id)?.name || 'Former member';

  const openForm = (plan?: RecurringCost) => {
    setEditing(plan || null);
    setForm(plan ? {
      description: plan.description, estimate: (plan.estimatePaise / 100).toFixed(2),
      dueDay: String(plan.dueDay), paidBy: plan.paidBy, participants: [...plan.participants]
    } : emptyForm);
    setFormOpen(true);
  };

  const toggleParticipant = (id: string) => setForm(previous => ({
    ...previous, participants: previous.participants.includes(id)
      ? previous.participants.filter(item => item !== id)
      : [...previous.participants, id]
  }));

  const save = async () => {
    const estimatePaise = parseRupeesToPaise(form.estimate);
    const dueDay = Number(form.dueDay);
    if (!form.description.trim() || estimatePaise === null || !Number.isInteger(dueDay) || dueDay < 1 || dueDay > 28 || !form.paidBy || form.participants.length === 0) {
      toast.error('Enter a name, positive estimate, due day 1–28, payer, and participants');
      return;
    }
    const input: RecurringInput = { description: form.description.trim(), estimatePaise, dueDay, paidBy: form.paidBy, participants: form.participants };
    setBusy(true);
    try {
      if (editing) await api.updateRecurring(group.id, editing.id, input);
      else await api.createRecurring(group.id, input);
      await onChanged();
      setFormOpen(false);
      toast.success(editing ? 'Schedule updated' : 'Recurring cost scheduled');
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not save schedule'); }
    finally { setBusy(false); }
  };

  const archive = async (plan: RecurringCost) => {
    if (!window.confirm(`Archive ${plan.description}? Its previous expense proposals will remain visible.`)) return;
    setBusy(true);
    try {
      await api.archiveRecurring(group.id, plan.id);
      await onChanged();
      toast.success('Schedule archived');
    } catch { toast.error('Could not archive schedule'); }
    finally { setBusy(false); }
  };

  const openProposal = (plan: RecurringCost) => {
    setProposing(plan);
    setActualAmount((plan.estimatePaise / 100).toFixed(2));
  };

  const actualPaise = parseRupeesToPaise(actualAmount);
  const baseline = proposing ? (proposing.recentCount > 0 ? proposing.recentAveragePaise : proposing.estimatePaise) : 0;
  const delta = actualPaise !== null ? Math.abs(actualPaise - baseline) : 0;
  const largeChange = actualPaise !== null && baseline > 0 && delta >= 1000 && delta * 100 >= baseline * 25;
  const changeMessage = largeChange && proposing && actualPaise !== null
    ? `${formatCurrency(actualPaise / 100)} is ${Math.floor(delta * 100 / baseline)}% ${actualPaise > baseline ? 'higher' : 'lower'} than the ${proposing.recentCount > 0 ? 'recent approved average' : 'scheduled estimate'} of ${formatCurrency(baseline / 100)}. Review before approving.`
    : '';

  const submitProposal = async () => {
    if (!proposing || actualPaise === null) { toast.error('Enter a positive actual amount'); return; }
    setBusy(true);
    try {
      const expense = await api.proposeRecurring(group.id, proposing.id, actualPaise);
      await onChanged();
      setProposing(null);
      toast.success(expense.status === 'active' ? 'Scheduled expense is active' : 'Scheduled expense proposed for approval');
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not propose scheduled bill'); }
    finally { setBusy(false); }
  };

  return <>
    <Card className="glass-card mb-8">
      <CardHeader><div className="flex flex-wrap items-center justify-between gap-3"><div><CardTitle>Recurring Cost Radar</CardTitle><p className="text-sm text-muted-foreground mt-1">Forecasts only. Upcoming costs do not change balances.</p></div><Button disabled={!group.currentMemberId} onClick={() => openForm()}>Schedule a cost</Button></div></CardHeader>
      <CardContent>
        <div className="rounded-lg border bg-accent/10 p-4 mb-5"><p className="text-sm text-muted-foreground">Estimated next 30 days</p><p className="text-2xl font-bold">{formatCurrency((group.monthAheadForecastPaise || 0) / 100)}</p><p className="text-xs text-muted-foreground">Based on active monthly schedules, not debt owed.</p></div>
        {group.recurring?.length ? <div className="grid gap-4 md:grid-cols-2">{group.recurring.map(plan => <div key={plan.id} className="rounded-lg border p-4 space-y-2">
          <div className="flex justify-between gap-3"><h3 className="font-semibold">{plan.description}</h3><strong>{formatCurrency(plan.estimatePaise / 100)}</strong></div>
          <p className="text-sm text-muted-foreground">Next due {formatDate(plan.nextDueDate)} · monthly on day {plan.dueDay}</p>
          <p className="text-sm">Expected payer: {name(plan.paidBy)}</p>
          <div className="text-sm"><strong>Estimated shares</strong><ul className="mt-1">{plan.participants.map(id => <li key={id}>{name(id)}: {formatCurrency((plan.estimatedSharesPaise[id] || 0) / 100)}</li>)}</ul></div>
          {plan.recentCount > 0 && <p className="text-xs text-muted-foreground">Recent approved average ({plan.recentCount}): {formatCurrency(plan.recentAveragePaise / 100)}</p>}
          <div className="flex flex-wrap gap-2 pt-2"><Button size="sm" disabled={!group.currentMemberId || !plan.canPropose || busy} onClick={() => openProposal(plan)}>{plan.canPropose ? 'Enter actual bill' : 'Current bill proposed'}</Button>{plan.canManage && <><Button size="sm" variant="outline" onClick={() => openForm(plan)}>Edit</Button><Button size="sm" variant="outline" disabled={busy} onClick={() => archive(plan)}>Archive</Button></>}</div>
        </div>)}</div> : <p className="text-sm text-muted-foreground">No monthly costs scheduled yet. Add rent, utilities, or another expected group cost.</p>}
      </CardContent>
    </Card>

    <Dialog open={formOpen} onOpenChange={setFormOpen}><DialogContent onClose={() => setFormOpen(false)} className="max-w-lg"><DialogHeader><DialogTitle>{editing ? 'Edit monthly cost' : 'Schedule monthly cost'}</DialogTitle></DialogHeader><div className="space-y-3 max-h-[70vh] overflow-y-auto mt-3 pr-1">
      <p className="text-sm text-muted-foreground">A schedule is a forecast. Actual bills require participant approval.</p>
      <div><Label htmlFor="recurring-name">Name</Label><Input id="recurring-name" maxLength={200} value={form.description} onChange={event => setForm({ ...form, description: event.target.value })} placeholder="Rent or electricity" /></div>
      <div><Label htmlFor="recurring-estimate">Estimated amount (₹)</Label><Input id="recurring-estimate" value={form.estimate} onChange={event => setForm({ ...form, estimate: event.target.value })} /></div>
      <div><Label htmlFor="recurring-day">Due day each month (1–28)</Label><Input id="recurring-day" type="number" min={1} max={28} value={form.dueDay} onChange={event => setForm({ ...form, dueDay: event.target.value })} /></div>
      <div><Label htmlFor="recurring-payer">Expected payer</Label><select id="recurring-payer" className="w-full rounded-md border bg-background p-2" value={form.paidBy} onChange={event => setForm({ ...form, paidBy: event.target.value })}><option value="">Select payer</option>{joined.map(member => <option key={member.id} value={member.id}>{member.name}</option>)}</select></div>
      <div><Label>Split between</Label>{joined.map(member => <label key={member.id} className="flex items-center gap-2 py-1"><Checkbox checked={form.participants.includes(member.id)} onCheckedChange={() => toggleParticipant(member.id)} />{member.name}</label>)}</div>
      <Button disabled={busy} onClick={save}>{editing ? 'Save changes' : 'Add to radar'}</Button>
    </div></DialogContent></Dialog>

    <Dialog open={!!proposing} onOpenChange={open => !open && setProposing(null)}><DialogContent onClose={() => setProposing(null)} className="max-w-lg"><DialogHeader><DialogTitle>Enter actual bill</DialogTitle></DialogHeader>{proposing && <div className="space-y-4 mt-3"><p className="text-sm">{proposing.description} · due {formatDate(proposing.nextDueDate)}</p><p className="text-sm text-muted-foreground">The actual bill becomes a proposal. It does not affect balances until all included members approve.</p><div><Label htmlFor="actual-bill">Actual amount (₹)</Label><Input id="actual-bill" value={actualAmount} onChange={event => setActualAmount(event.target.value)} /></div>{changeMessage && <p role="alert" className="text-sm rounded-md border border-amber-400/50 bg-amber-400/10 p-3">{changeMessage}</p>}<Button disabled={busy} onClick={submitProposal}>Propose bill for approval</Button></div>}</DialogContent></Dialog>
  </>;
}
