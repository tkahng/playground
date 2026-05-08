import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function StripeTestCardNotice() {
  return (
    <Alert className="border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-100">
      <AlertTitle>Test environment — no real charges</AlertTitle>
      <AlertDescription>
        Use Stripe test card{" "}
        <code className="font-mono font-semibold">4242 4242 4242 4242</code>,
        any future expiry, any 3-digit CVC, and any ZIP.{" "}
        <a
          href="https://docs.stripe.com/testing?testing-method=card-numbers#visa"
          target="_blank"
          rel="noopener noreferrer"
          className="underline underline-offset-2"
        >
          More test cards
        </a>
      </AlertDescription>
    </Alert>
  );
}
