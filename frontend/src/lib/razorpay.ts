export interface RazorpayCheckoutResponse {
  razorpay_order_id: string;
  razorpay_payment_id: string;
  razorpay_signature: string;
}

interface RazorpayInstance { open(): void }
interface RazorpayWindow extends Window {
  Razorpay?: new (options: {
    key: string;
    amount: number;
    currency: string;
    name: string;
    description: string;
    order_id: string;
    handler: (response: RazorpayCheckoutResponse) => void;
    theme: { color: string };
  }) => RazorpayInstance;
}

let scriptPromise: Promise<void> | null = null;

export async function openRazorpayCheckout(options: ConstructorParameters<NonNullable<RazorpayWindow['Razorpay']>>[0]): Promise<void> {
  const browser = window as RazorpayWindow;
  if (!browser.Razorpay) {
    if (!scriptPromise) {
      scriptPromise = new Promise<void>((resolve, reject) => {
        const script = document.createElement('script');
        script.src = 'https://checkout.razorpay.com/v1/checkout.js';
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => { script.remove(); scriptPromise = null; reject(new Error('Could not load Razorpay checkout')); };
        document.body.appendChild(script);
      });
    }
    await scriptPromise;
  }
  if (!browser.Razorpay) throw new Error('Razorpay checkout is unavailable');
  new browser.Razorpay(options).open();
}
