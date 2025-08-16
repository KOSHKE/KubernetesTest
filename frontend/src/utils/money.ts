// Currency utilities and money formatting in minor units
// Single Responsibility: isolate currency exponent rules and formatting

const CURRENCY_FRACTION_DIGITS: Record<string, number> = {
  JPY: 0,
  KRW: 0,
  VND: 0,
  KWD: 3,
  BHD: 3,
  JOD: 3,
  OMR: 3,
  TND: 3,
};

export function getFractionDigits(currency?: string): number {
  if (!currency) return 2;
  return CURRENCY_FRACTION_DIGITS[currency.toUpperCase()] ?? 2;
}

export function formatMoneyMinor(amountMinor?: number, currency?: string, locale = 'en-US'): string {
  const amount = Number(amountMinor || 0);
  const digits = getFractionDigits(currency);
  const divisor = Math.pow(10, digits);
  const major = amount / divisor;
  if (currency) {
    try {
      return new Intl.NumberFormat(locale, { style: 'currency', currency, minimumFractionDigits: digits, maximumFractionDigits: digits }).format(major);
    } catch {
      // Fallback if Intl doesn't know the currency
      return `${major.toFixed(digits)} ${currency}`.trim();
    }
  }
  return major.toFixed(digits);
}

export function sumMinorWithSameCurrency(values: Array<{ amountMinor: number; currency?: string }>): { amountMinor: number; currency?: string } {
  if (values.length === 0) return { amountMinor: 0 };
  const currency = values[0].currency;
  const total = values.reduce((acc, v) => acc + Number(v.amountMinor || 0), 0);
  return { amountMinor: total, currency };
}


