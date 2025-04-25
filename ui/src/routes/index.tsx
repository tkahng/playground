import RootLayout from "@/layouts/root";
import LandingAboutPage from "@/pages/landing/about";
import LandingContactPage from "@/pages/landing/contact";
import Features from "@/pages/landing/features";
import Landing from "@/pages/landing/landing";
import PricingPage from "@/pages/landing/pricing";

export const routes = [
  {
    element: <RootLayout />,
    children: [
      {
        path: "/",
        element: <Landing />,
      },
      {
        path: "/home",
        element: <Landing />,
      },
      {
        path: "/features",
        element: <Features />,
      },
      {
        path: "/pricing",
        element: <PricingPage />,
      },
      {
        path: "/about",
        element: <LandingAboutPage />,
      },
      {
        path: "/contact",
        element: <LandingContactPage />,
      },
    ],
  },
];
