import React, { FC } from 'react';
import { Card } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { formatCurrency, formatDate } from '../lib/utils';
import { Receipt, Users, Trash2 } from 'lucide-react';

interface Member {
  id: string;
  name: string;
}

interface ExpenseCardProps {
  expense: {
    id: string;
    description: string;
    amount: number;
    amountPaise: number;
    date: string;
    paidBy: string;
    participants: string[];
    status?: string;
    approvals?: { status: string }[];
    anomalyExplanation?: string;
  };
  members: Member[];
  onDelete?: (expenseId: string) => void;
  onOpen?: (expenseId: string) => void;
}

export const ExpenseCard: FC<ExpenseCardProps> = ({ expense, members, onDelete, onOpen }) => {
  const paidByMember = members.find(m => m.id === expense.paidBy);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (window.confirm(`Are you sure you want to delete "${expense.description}"? This action cannot be undone.`)) {
      onDelete?.(expense.id);
    }
  };

  return (
    <Card className="hover:shadow-elegant transition-all group">
      <div className="p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-10 h-10 rounded-lg bg-gradient-primary flex items-center justify-center text-primary-foreground shadow-md">
                <Receipt className="w-5 h-5" />
              </div>
              <div className="flex-1">
                <h4 className="font-semibold text-base group-hover:text-primary transition-colors">
                  {expense.description}
                </h4>
                <p className="text-xs text-muted-foreground">
                  {formatDate(expense.date)}
                </p>
              </div>
              
              {onDelete && (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleDelete}
                  className="text-destructive hover:text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-2 mt-3">
              <Badge variant="outline" className="text-xs capitalize">{expense.status || 'active'}</Badge>
              {expense.anomalyExplanation && <Badge variant="outline" className="text-xs border-amber-400/50">Unusual amount</Badge>}
              {expense.status === 'proposed' && <span className="text-xs text-muted-foreground">{expense.approvals?.filter(item => item.status === 'pending').length || 0} awaiting approval</span>}
              <Badge variant="outline" className="text-xs">
                <Users className="w-3 h-3 mr-1" />
                {expense.participants.length} people
              </Badge>
              <span className="text-xs text-muted-foreground">
                Paid by <span className="font-medium text-foreground">{paidByMember?.name}</span>
              </span>
              {onOpen && <button type="button" className="text-xs text-primary hover:underline" onClick={() => onOpen(expense.id)}>View approvals and discussion →</button>}
            </div>
          </div>
          
          <div className="text-right">
            <p className="text-2xl font-bold gradient-text">{formatCurrency(expense.amount)}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {expense.participants.length > 0 && expense.amountPaise % expense.participants.length === 0
                ? `${formatCurrency(expense.amountPaise / expense.participants.length / 100)}/person`
                : 'Equal split; remainder assigned in participant order'}
            </p>
          </div>
        </div>
      </div>
    </Card>
  );
};
