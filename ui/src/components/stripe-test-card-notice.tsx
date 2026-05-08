import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function StripeTestCardNotice() {
  return (
    <Alert>
      <AlertTitle>Test environment — no real charges</AlertTitle>
      <AlertDescription>
        <p>
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
        </p>
      </AlertDescription>
    </Alert>
  );
}
