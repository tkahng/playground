import { Facebook, Github, Linkedin, Twitter } from "lucide-react";

// container mx-auto flex h-14 items-center justify-between  lg:px-6
export function PlaygroundFooter() {
  return (
    <footer className="border-t w-full mt-auto bg-background">
      <div className="container mx-auto flex-col items-center justify-between px-4 py-12 md:px-6 lg:px-8">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4 justify-center">
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Product</h4>
              <ul className="space-y-2">
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Features
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Pricing
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    API
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Integrations
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Resources</h4>
              <ul className="space-y-2">
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Documentation
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Tutorials
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Blog
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Support
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Company</h4>
              <ul className="space-y-2">
                <li>
                  <a href="#" className="text-sm hover:underline">
                    About
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Careers
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Privacy Policy
                  </a>
                </li>
                <li>
                  <a href="#" className="text-sm hover:underline">
                    Terms of Service
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Social</h4>
              <ul className="space-y-2">
                <li>
                  <a
                    href="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Twitter className="mr-2 h-5 w-5" /> Twitter
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Facebook className="mr-2 h-5 w-5" /> Facebook
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Linkedin className="mr-2 h-5 w-5" /> LinkedIn
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Github className="mr-2 h-5 w-5" /> GitHub
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>
        <div className="mt-8 border-t pt-8">
          <p className="text-center text-xs text-gray-500 dark:text-gray-400">
            © 2026 Playground. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
