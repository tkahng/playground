import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function TermsOfServicePage() {
  return (
    <div className="container px-4 md:px-6">
      <div className="mx-auto max-w-3xl">
        <h1 className="mb-6 text-3xl font-bold">Terms and Conditions</h1>
        <Card>
          <CardHeader>
            <CardTitle>NexusAI Terms and Conditions</CardTitle>
            <CardDescription>Last updated: June 1, 2023</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <p>
              Welcome to NexusAI. These Terms and Conditions govern your use of
              our website and AI services. By accessing or using NexusAI, you
              agree to be bound by these Terms. Please read them carefully.
            </p>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                1. Acceptance of Terms
              </h2>
              <p>
                By accessing or using NexusAI services, you agree to comply with
                and be bound by these Terms and Conditions. If you do not agree
                to these Terms, please do not use our services.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                2. Description of Service
              </h2>
              <p>
                NexusAI provides artificial intelligence services, including but
                not limited to machine learning models, APIs, and data analysis
                tools. The specific features and functionality may change over
                time.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">3. User Accounts</h2>
              <p>
                To access certain features of NexusAI, you may be required to
                create an account. You are responsible for maintaining the
                confidentiality of your account information and for all
                activities that occur under your account.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">4. Acceptable Use</h2>
              <p>
                You agree not to use NexusAI for any unlawful purpose or in any
                way that:
              </p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>Violates any applicable laws or regulations</li>
                <li>Infringes on the rights of others</li>
                <li>
                  Interferes with or disrupts the integrity of our services
                </li>
                <li>
                  Attempts to gain unauthorized access to our systems or user
                  accounts
                </li>
              </ul>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                5. Intellectual Property
              </h2>
              <p>
                All content and materials available on NexusAI, including but
                not limited to text, graphics, logos, and software, are the
                property of NexusAI or its licensors and are protected by
                copyright and other intellectual property laws.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                6. Limitation of Liability
              </h2>
              <p>
                NexusAI and its affiliates shall not be liable for any indirect,
                incidental, special, consequential, or punitive damages
                resulting from your use of or inability to use our services.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                7. Modifications to Service
              </h2>
              <p>
                We reserve the right to modify, suspend, or discontinue any part
                of NexusAI services at any time without notice or liability.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">8. Governing Law</h2>
              <p>
                These Terms shall be governed by and construed in accordance
                with the laws of [Your Jurisdiction], without regard to its
                conflict of law provisions.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                9. Changes to Terms
              </h2>
              <p>
                We may update these Terms from time to time. We will notify you
                of any changes by posting the new Terms on this page and
                updating the "Last updated" date at the top of this document.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">10. Contact Us</h2>
              <p>
                If you have any questions about these Terms, please contact us
                at legal~nexusai.com.
              </p>
            </section>

            <div className="mt-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">
                By using NexusAI services, you acknowledge that you have read
                and understood these Terms and Conditions and agree to be bound
                by them.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
