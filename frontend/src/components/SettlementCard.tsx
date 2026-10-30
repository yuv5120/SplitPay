import { FC } from 'react';
import { Card } from './ui/card';
import { Avatar, AvatarFallback } from './ui/avatar';
import { Button } from './ui/button';
import { formatCurrency } from '../lib/utils';
import { ArrowRight, CheckCircle } from 'lucide-react';

interface Member {
  id: string;
  name: string;
}

interface Settlement {
  from: string;
  to: string;
  amount: number;
}

interface SettlementCardProps {
  settlement: Settlement;
  members: Member[];
  onSettle?: (settlement: Settlement) => void;
  onRemind?: (settlement: Settlement) => void;
}

export const SettlementCard: FC<SettlementCardProps> = ({ settlement, members, onSettle, onRemind }: SettlementCardProps) => {
  const fromMember = members.find((m: Member) => m.id === settlement.from);
  const toMember = members.find((m: Member) => m.id === settlement.to);

  if (!fromMember || !toMember) return null;

  return (
    <Card className="hover:shadow-md transition-all">
      <div className="p-4">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3 flex-1">
            <Avatar className="w-10 h-10">
              <AvatarFallback>
                {fromMember.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="font-medium text-sm">{fromMember.name}</span>
          </div>

          <div className="flex flex-col items-center gap-1">
            <ArrowRight className="w-5 h-5 text-primary" />
            <span className="text-lg font-bold gradient-text">
              {formatCurrency(settlement.amount)}
            </span>
          </div>

          <div className="flex items-center gap-3 flex-1 justify-end">
            <span className="font-medium text-sm text-right">{toMember.name}</span>
            <Avatar className="w-10 h-10">
              <AvatarFallback>
                {toMember.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
          </div>
        </div>

        {(onSettle || onRemind) && (
          <div className="mt-3 pt-3 border-t flex gap-2">
            {onSettle && <Button size="sm" className="flex-1" onClick={() => onSettle(settlement)}><CheckCircle className="w-4 h-4 mr-2" />Settle up</Button>}
            {onRemind && <Button size="sm" variant="outline" className="flex-1" onClick={() => onRemind(settlement)}>Send reminder</Button>}
          </div>
        )}
      </div>
    </Card>
  );
};
