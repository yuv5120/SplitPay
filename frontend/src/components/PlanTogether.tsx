import { useState } from 'react';
import axios from 'axios';
import { toast } from 'sonner';
import { api, Group, PurchasePlan } from '../lib/api';
import { parseRupeesToPaise } from '../lib/money';
import { formatCurrency, formatDateTime } from '../lib/utils';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Checkbox } from './ui/checkbox';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';

type OptionForm = { label: string; amount: string };
const initialOptions = (): OptionForm[] => [{ label: '', amount: '' }, { label: '', amount: '' }];
const money = (paise: number) => formatCurrency(paise / 100);
const message = (error: unknown) => axios.isAxiosError(error) ? error.response?.data?.message || error.message : error instanceof Error ? error.message : 'Please try again';

export function PlanTogether({ group, onChanged, onExpenseCreated }: { group: Group; onChanged: () => Promise<void>; onExpenseCreated: (id: string) => void }) {
  const [createOpen, setCreateOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [details, setDetails] = useState('');
  const [days, setDays] = useState('3');
  const [options, setOptions] = useState<OptionForm[]>(initialOptions);
  const [comments, setComments] = useState<Record<string, string>>({});
  const [converting, setConverting] = useState<PurchasePlan | null>(null);
  const [payer, setPayer] = useState('');
  const [participants, setParticipants] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const joined = group.members.filter(member => member.isJoined);

  const run = async (work: () => Promise<void>, success: string) => {
    setBusy(true);
    try { await work(); await onChanged(); toast.success(success); }
    catch (error) { toast.error(message(error)); }
    finally { setBusy(false); }
  };

  const create = async () => {
    const parsed = options.map(option => ({ label: option.label.trim(), amountPaise: parseRupeesToPaise(option.amount) }));
    const votingDays = Number(days);
    if (!title.trim() || parsed.some(option => !option.label || option.amountPaise === null) || !Number.isInteger(votingDays) || votingDays < 1 || votingDays > 14) {
      toast.error('Add a title, two or more priced options, and 1–14 voting days'); return;
    }
    await run(async () => {
      await api.createPurchasePlan(group.id, { title: title.trim(), details: details.trim(), votingDays,
        options: parsed.map(option => ({ label: option.label, amountPaise: option.amountPaise! })) });
      setCreateOpen(false); setTitle(''); setDetails(''); setDays('3'); setOptions(initialOptions());
    }, 'Purchase plan shared with the group');
  };

  const openConvert = (plan: PurchasePlan) => {
    setConverting(plan);
    setPayer(group.currentMemberId || joined[0]?.id || '');
    setParticipants(joined.map(member => member.id));
  };
  const toggleParticipant = (id: string) => setParticipants(previous => previous.includes(id) ? previous.filter(value => value !== id) : [...previous, id]);
  const convert = async () => {
    if (!converting || !payer || !participants.length) { toast.error('Choose a payer and at least one participant'); return; }
    const planID = converting.id;
    await run(async () => {
      const expense = await api.convertPurchasePlan(group.id, planID, payer, participants);
      setConverting(null);
      onExpenseCreated(expense.id);
    }, 'Chosen option sent through expense approval');
  };

  return <>
    <Card className="glass-card mb-8">
      <CardHeader><div className="flex flex-wrap items-start justify-between gap-3"><div><CardTitle>Plan Together</CardTitle><p className="text-sm text-muted-foreground mt-1">Compare ideas and agree before anyone spends.</p></div><Button disabled={joined.length < 2 || (group.purchasePlans?.length || 0) >= 50} onClick={() => setCreateOpen(true)}>Propose a purchase</Button></div></CardHeader>
      <CardContent className="space-y-5">
        {!group.purchasePlans?.length && <p className="text-sm text-muted-foreground">No purchase plans yet. Once two members join, compare priced options and vote together.</p>}
        {group.purchasePlans?.map(plan => {
          const selected = plan.options.find(option => option.id === plan.selectedOptionId);
          return <section key={plan.id} className="rounded-xl border bg-card/40 p-4 space-y-3" aria-label={plan.title}>
            <div className="flex flex-wrap items-start justify-between gap-2"><div><h3 className="font-semibold text-lg">{plan.title}</h3>{plan.details && <p className="text-sm text-muted-foreground">{plan.details}</p>}</div><span className="rounded-full border px-2 py-1 text-xs capitalize">{plan.status}</span></div>
            <p className="text-xs text-muted-foreground">Voting ends {formatDateTime(plan.deadlineAt)}. The lower-cost option wins a tied vote.</p>
            <div className="grid gap-2 sm:grid-cols-2">{plan.options.map(option => {
              const projected = (group.budget?.projectedPaise || 0) + option.amountPaise;
              const over = (group.budget?.limitPaise || 0) > 0 && projected > group.budget.limitPaise;
              const chosen = plan.selectedOptionId === option.id;
              return <div key={option.id} className={`rounded-lg border p-3 space-y-1 ${chosen ? 'border-primary bg-primary/5' : ''}`}>
                <div className="flex items-start justify-between gap-2"><p className="font-medium">{option.label}</p><strong>{money(option.amountPaise)}</strong></div>
                <p className="text-xs text-muted-foreground">{plan.voteCounts?.[option.id] || 0} vote{(plan.voteCounts?.[option.id] || 0) === 1 ? '' : 's'}{chosen ? ' · selected' : ''}</p>
                {group.budget?.limitPaise > 0 && <p className={`text-xs ${over ? 'text-amber-700' : 'text-muted-foreground'}`}>If bought this month: {over ? `${money(projected - group.budget.limitPaise)} over target` : `${money(group.budget.limitPaise - projected)} left`}</p>}
                {plan.canVote && <Button size="sm" variant={plan.myVote === option.id ? 'premium' : 'outline'} disabled={busy} onClick={() => run(() => api.votePurchasePlan(group.id, plan.id, option.id), 'Vote saved')}>{plan.myVote === option.id ? 'Your vote' : 'Vote for this'}</Button>}
              </div>;
            })}</div>
            <div className="flex flex-wrap gap-2">{plan.canClose && <Button size="sm" disabled={busy} onClick={() => run(() => api.closePurchasePlan(group.id, plan.id), 'Voting closed')}>Close voting</Button>}{plan.canCancel && <Button size="sm" variant="outline" disabled={busy} onClick={() => { if (window.confirm(`Cancel ${plan.title}?`)) run(() => api.cancelPurchasePlan(group.id, plan.id), 'Plan cancelled'); }}>Cancel plan</Button>}{plan.canConvert && <Button size="sm" disabled={busy} onClick={() => openConvert(plan)}>Create expense proposal</Button>}{plan.convertedExpenseId && <p className="text-sm text-muted-foreground">Expense proposal created from {selected?.label || 'selected option'}.</p>}</div>
            <div className="border-t pt-3 space-y-2"><h4 className="text-sm font-semibold">Discussion</h4>{plan.comments?.length ? plan.comments.map(comment => <div key={comment.id} className="rounded-lg bg-muted/40 p-2 text-sm"><span className="font-medium">{comment.authorName}</span> <span className="text-xs text-muted-foreground">{formatDateTime(comment.createdAt)}</span><p className="whitespace-pre-wrap break-words">{comment.text}</p></div>) : <p className="text-sm text-muted-foreground">No comments yet.</p>}{group.currentMemberId && plan.status !== 'cancelled' && (plan.comments?.length || 0) < 100 && <div className="flex gap-2"><Input aria-label={`Comment on ${plan.title}`} value={comments[plan.id] || ''} maxLength={1000} onChange={event => setComments(previous => ({ ...previous, [plan.id]: event.target.value }))} placeholder="Add a thought for the group" /><Button size="sm" variant="outline" disabled={busy || !comments[plan.id]?.trim()} onClick={() => run(async () => { await api.commentPurchasePlan(group.id, plan.id, comments[plan.id]); setComments(previous => ({ ...previous, [plan.id]: '' })); }, 'Comment added')}>Post</Button></div>}</div>
          </section>;
        })}
      </CardContent>
    </Card>

    <Dialog open={createOpen} onOpenChange={setCreateOpen}><DialogContent onClose={() => setCreateOpen(false)} className="max-w-xl"><DialogHeader><DialogTitle>Propose a purchase</DialogTitle></DialogHeader><div className="space-y-4 mt-3"><div><Label htmlFor="plan-title">What are you deciding?</Label><Input id="plan-title" maxLength={100} value={title} onChange={event => setTitle(event.target.value)} placeholder="Weekend stay" /></div><div><Label htmlFor="plan-details">Context (optional)</Label><Input id="plan-details" maxLength={500} value={details} onChange={event => setDetails(event.target.value)} placeholder="Close to the station, for four people" /></div><div><Label htmlFor="plan-days">Voting days</Label><Input id="plan-days" type="number" min={1} max={14} value={days} onChange={event => setDays(event.target.value)} /></div><div className="space-y-2"><Label>Options and estimated group cost</Label>{options.map((option, index) => <div key={index} className="flex flex-wrap gap-2"><Input aria-label={`Option ${index + 1} name`} className="flex-1 min-w-32" maxLength={100} value={option.label} onChange={event => setOptions(previous => previous.map((item, i) => i === index ? { ...item, label: event.target.value } : item))} placeholder={`Option ${index + 1}`} /><Input aria-label={`Option ${index + 1} amount in rupees`} className="w-32" inputMode="decimal" value={option.amount} onChange={event => setOptions(previous => previous.map((item, i) => i === index ? { ...item, amount: event.target.value } : item))} placeholder="₹ amount" />{options.length > 2 && <Button size="sm" variant="outline" onClick={() => setOptions(previous => previous.filter((_, i) => i !== index))}>Remove</Button>}</div>)}{options.length < 5 && <Button size="sm" variant="outline" onClick={() => setOptions(previous => [...previous, { label: '', amount: '' }])}>Add option</Button>}</div><Button disabled={busy} onClick={create}>Share plan</Button></div></DialogContent></Dialog>

    <Dialog open={!!converting} onOpenChange={open => !open && setConverting(null)}><DialogContent onClose={() => setConverting(null)}><DialogHeader><DialogTitle>Create expense proposal</DialogTitle></DialogHeader>{converting && <div className="space-y-4 mt-3"><p className="text-sm">{converting.options.find(option => option.id === converting.selectedOptionId)?.label} · {money(converting.options.find(option => option.id === converting.selectedOptionId)?.amountPaise || 0)}</p><p className="text-xs text-muted-foreground">Use this after the purchase. The expense still needs participant approval before it changes balances.</p><div><Label htmlFor="plan-payer">Who paid?</Label><select id="plan-payer" className="w-full rounded-md border bg-background p-2" value={payer} onChange={event => setPayer(event.target.value)}>{joined.map(member => <option key={member.id} value={member.id}>{member.name}</option>)}</select></div><div><Label>Include in the equal split</Label>{joined.map(member => <label key={member.id} className="flex items-center gap-2 py-1"><Checkbox checked={participants.includes(member.id)} onCheckedChange={() => toggleParticipant(member.id)} />{member.name}</label>)}</div><Button disabled={busy} onClick={convert}>Create proposal</Button></div>}</DialogContent></Dialog>
  </>;
}
