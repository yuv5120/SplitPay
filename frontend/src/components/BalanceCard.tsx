import { FC } from 'react';
import { Card } from './ui/card';
import { Avatar, AvatarFallback } from './ui/avatar';
import { formatCurrency, cn } from '../lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface Member {
  id: string;
  name: string;
}

interface BalanceCardProps {
  member: Member;
  balance: number;
}

export const BalanceCard: FC<BalanceCardProps> = ({ member, balance }) => {
  const getBalanceStatus = (): 'positive' | 'negative' | 'settled' => {
    if (balance > 0.01) return 'positive';
    if (balance < -0.01) return 'negative';
    return 'settled';
  };

  const status = getBalanceStatus();

  return (
    <Card className="hover:shadow-md transition-all">
      <div className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Avatar className="w-12 h-12">
              <AvatarFallback className="text-base">
                {member.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="font-semibold">{member.name}</p>
              <p className="text-xs text-muted-foreground">
                {status === 'positive' && 'Gets back'}
                {status === 'negative' && 'Owes'}
                {status === 'settled' && 'Settled up'}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {status === 'positive' && (
              <div className="p-2 rounded-lg bg-success/10">
                <TrendingUp className="w-5 h-5 text-success" />
              </div>
            )}
            {status === 'negative' && (
              <div className="p-2 rounded-lg bg-destructive/10">
                <TrendingDown className="w-5 h-5 text-destructive" />
              </div>
            )}
            {status === 'settled' && (
              <div className="p-2 rounded-lg bg-muted">
                <Minus className="w-5 h-5 text-muted-foreground" />
              </div>
            )}
            <p className={cn(
              'text-lg font-bold',
              status === 'positive' && 'text-success',
              status === 'negative' && 'text-destructive',
              status === 'settled' && 'text-muted-foreground'
            )}>
              {formatCurrency(Math.abs(balance))}
            </p>
          </div>
        </div>
      </div>
    </Card>
  );
};
