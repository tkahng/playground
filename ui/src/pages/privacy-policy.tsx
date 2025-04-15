import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function PrivacyPolicyPage() {
  return (
    <div className="container px-4 md:px-6">
      <div className="mx-auto max-w-3xl">
        <h1 className="mb-6 text-3xl font-bold">Privacy Policy</h1>
        <Card>
          <CardHeader>
            <CardTitle>NexusAI Privacy Policy</CardTitle>
            <CardDescription>Last updated: June 1, 2023</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <p>
              At NexusAI, we are committed to protecting your privacy and
              ensuring the security of your personal information. This Privacy
              Policy explains how we collect, use, disclose, and safeguard your
              data when you use our AI services and visit our website.
            </p>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                1. Information We Collect
              </h2>
              <p>We collect the following types of information:</p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>
                  Personal information (e.g., name, email address) when you
                  create an account
                </li>
                <li>
                  Usage data related to your interactions with our AI models and
                  APIs
                </li>
                <li>Payment information when you subscribe to our services</li>
                <li>
                  Cookies and similar tracking technologies on our website
                </li>
              </ul>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                2. How We Use Your Information
              </h2>
              <p>We use your information to:</p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>Provide, maintain, and improve our AI services</li>
                <li>Process payments and prevent fraud</li>
                <li>Send you important updates and communications</li>
                <li>Analyze usage patterns to enhance user experience</li>
                <li>Comply with legal obligations</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                3. Data Sharing and Disclosure
              </h2>
              <p>We may share your information with:</p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>Service providers who assist in our operations</li>
                <li>
                  Law enforcement or government agencies when required by law
                </li>
                <li>Business partners with your consent</li>
              </ul>
              <p className="mt-2">
                We do not sell your personal information to third parties.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">4. Data Security</h2>
              <p>
                We implement robust security measures to protect your data,
                including:
              </p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>Encryption of data in transit and at rest</li>
                <li>Regular security audits and penetration testing</li>
                <li>Access controls and authentication mechanisms</li>
                <li>Employee training on data protection</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                5. Your Rights and Choices
              </h2>
              <p>You have the right to:</p>
              <ul className="mt-2 list-disc space-y-1 pl-6">
                <li>Access and update your personal information</li>
                <li>Request deletion of your data</li>
                <li>Opt-out of certain data processing activities</li>
                <li>Withdraw consent where applicable</li>
              </ul>
              <p className="mt-2">
                Contact us at privacy~nexusai.com to exercise these rights.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">
                6. Changes to This Policy
              </h2>
              <p>
                We may update this Privacy Policy from time to time. We will
                notify you of any changes by posting the new Privacy Policy on
                this page and updating the "Last updated" date at the top of
                this policy.
              </p>
            </section>

            <section>
              <h2 className="mb-2 text-xl font-semibold">7. Contact Us</h2>
              <p>
                If you have any questions or concerns about our privacy
                practices, please contact us at privacy~nexusai.com.
              </p>
            </section>

            <div className="mt-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">
                By using NexusAI services, you agree to the terms outlined in
                this Privacy Policy. We encourage you to review this policy
                regularly for any changes.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
