// Debt simplification algorithm
// Minimizes the number of transactions needed to settle all debts

interface Transaction {
  from: string;
  to: string;
  amount: number;
}

interface Balances {
  [memberId: string]: number;
}

interface Creditor {
  memberId: string;
  amount: number;
}

interface Debtor {
  memberId: string;
  amount: number;
}

export function simplifyDebts(balances: Balances): Transaction[] {
  // balances is an object: { memberId: balance }
  // Positive balance means person is owed money
  // Negative balance means person owes money
  
  const creditors: Creditor[] = []; // People who are owed money
  const debtors: Debtor[] = [];   // People who owe money
  
  // Separate into creditors and debtors
  Object.entries(balances).forEach(([memberId, balance]) => {
    if (balance > 0.01) {
      creditors.push({ memberId, amount: balance });
    } else if (balance < -0.01) {
      debtors.push({ memberId, amount: Math.abs(balance) });
    }
  });
  
  const transactions: Transaction[] = [];
  
  // Sort for optimal matching
  creditors.sort((a, b) => b.amount - a.amount);
  debtors.sort((a, b) => b.amount - a.amount);
  
  let i = 0, j = 0;
  
  while (i < creditors.length && j < debtors.length) {
    const creditor = creditors[i];
    const debtor = debtors[j];
    
    const settleAmount = Math.min(creditor.amount, debtor.amount);
    
    if (settleAmount > 0.01) {
      transactions.push({
        from: debtor.memberId,
        to: creditor.memberId,
        amount: settleAmount,
      });
    }
    
    creditor.amount -= settleAmount;
    debtor.amount -= settleAmount;
    
    if (creditor.amount < 0.01) i++;
    if (debtor.amount < 0.01) j++;
  }
  
  return transactions;
}

interface Member {
  id: string;
  [key: string]: any;
}

interface Expense {
  paidBy: string;
  amount: number;
  participants: string[];
}

export function calculateBalances(expenses: Expense[], members: Member[]): Balances {
  // Initialize balances
  const balances: Balances = {};
  members.forEach(member => {
    balances[member.id] = 0;
  });
  
  // Calculate balances from expenses
  expenses.forEach(expense => {
    const { paidBy, amount, participants } = expense;
    const splitAmount = amount / participants.length;
    
    // Person who paid gets credited
    balances[paidBy] += amount;
    
    // Each participant gets debited their share
    participants.forEach(participantId => {
      balances[participantId] -= splitAmount;
    });
  });
  
  return balances;
}
