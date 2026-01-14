import React, { FC } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardHeader, CardTitle, CardContent } from './ui/card';
import { Badge } from './ui/badge';
import { Avatar, AvatarFallback } from './ui/avatar';
import { Button } from './ui/button';
import { formatCurrency } from '../lib/utils';
import { Users, Receipt, TrendingUp, Trash2 } from 'lucide-react';

interface Member {
  id: string;
  name: string;
}

interface Expense {
  id: string;
  description: string;
  amount: number;
  date: string;
  paidBy: string;
  participants: string[];
}

interface GroupCardProps {
  group: {
    id: string;
    name: string;
    members: Member[];
    expenses?: Expense[];
  };
  onDelete?: (groupId: string) => void;
}

export const GroupCard: FC<GroupCardProps> = ({ group, onDelete }) => {
  const navigate = useNavigate();
  
  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (window.confirm(`Are you sure you want to delete "${group.name}"? This action cannot be undone.`)) {
      onDelete?.(group.id);
    }
  };
  
  const totalExpenses = group.expenses?.reduce((sum, exp) => sum + exp.amount, 0) || 0;
  const expenseCount = group.expenses?.length || 0;

  return (
    <Card 
      className="cursor-pointer hover:scale-[1.02] transition-all hover:shadow-elegant group"
      onClick={() => navigate(`/group/${group.id}`)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-xl mb-2 group-hover:gradient-text transition-all">
              {group.name}
            </CardTitle>
            <div className="flex flex-wrap gap-2">
              <Badge variant="glass" className="text-xs">
                <Users className="w-3 h-3 mr-1" />
                {group.members.length} members
              </Badge>
              <Badge variant="glass" className="text-xs">
                <Receipt className="w-3 h-3 mr-1" />
                {expenseCount} expenses
              </Badge>
            </div>
          </div>
          <div className="w-12 h-12 rounded-xl gradient-secondary flex items-center justify-center shadow-md">
            <TrendingUp className="w-6 h-6 text-secondary-foreground" />
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Total Expenses</span>
            <span className="text-lg font-bold gradient-text">{formatCurrency(totalExpenses)}</span>
          </div>
          
          <div className="flex items-center justify-between gap-2">
            <div className="flex -space-x-2 flex-1">
              {group.members.slice(0, 5).map((member) => (
                <Avatar key={member.id} className="border-2 border-card">
                  <AvatarFallback className="text-xs">
                    {member.name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              ))}
              {group.members.length > 5 && (
                <div className="w-10 h-10 rounded-full border-2 border-card bg-muted flex items-center justify-center">
                  <span className="text-xs font-medium">+{group.members.length - 5}</span>
                </div>
              )}
            </div>
            
            {onDelete ? (
              <Button
                variant="ghost"
                size="icon"
                onClick={handleDelete}
                className="text-destructive hover:text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            ) : null}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
