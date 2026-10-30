/** Parse a positive INR amount without passing through floating point. */
export function parseRupeesToPaise(value: string): number | null {
  const trimmed = value.trim();
  if (!/^(?:0|[1-9]\d*)(?:\.\d{1,2})?$/.test(trimmed)) return null;
  const [rupees, fractional = ''] = trimmed.split('.');
  const paise = Number(rupees) * 100 + Number(fractional.padEnd(2, '0'));
  return Number.isSafeInteger(paise) && paise > 0 && paise <= 100000000000 ? paise : null;
}
