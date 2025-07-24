import { Facebook, Github, Linkedin, Twitter } from "lucide-react";
import { NavLink } from "react-router";
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
                  <NavLink to="#" className="text-sm hover:underline">
                    Features
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Pricing
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    API
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Integrations
                  </NavLink>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Resources</h4>
              <ul className="space-y-2">
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Documentation
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Tutorials
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Blog
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Support
                  </NavLink>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Company</h4>
              <ul className="space-y-2">
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    About
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Careers
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Privacy Policy
                  </NavLink>
                </li>
                <li>
                  <NavLink to="#" className="text-sm hover:underline">
                    Terms of Service
                  </NavLink>
                </li>
              </ul>
            </div>
          </div>
          <div className="space-y-6 flex justify-center">
            <div className="max-w-[160px]">
              <h4 className="text-lg font-semibold mb-4">Social</h4>
              <ul className="space-y-2">
                <li>
                  <NavLink
                    to="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Twitter className="mr-2 h-5 w-5" /> Twitter
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Facebook className="mr-2 h-5 w-5" /> Facebook
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Linkedin className="mr-2 h-5 w-5" /> LinkedIn
                  </NavLink>
                </li>
                <li>
                  <NavLink
                    to="#"
                    className="flex items-center text-sm hover:underline"
                  >
                    <Github className="mr-2 h-5 w-5" /> GitHub
                  </NavLink>
                </li>
              </ul>
            </div>
          </div>
        </div>
        <div className="mt-8 border-t pt-8">
          <p className="text-center text-xs text-gray-500 dark:text-gray-400">
            © 2023 Playground. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
