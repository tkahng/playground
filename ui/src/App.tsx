import Dashboard from "@/pages/dashboard"; // Your protected page
import Signin from "@/pages/signin";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import AuthenticatedLayout from "./layouts/authenticated-layout";
import AuthenticatedLayoutBase from "./layouts/authenticated-layout-base";
import RootLayout from "./layouts/root";
import LandingAboutPage from "./pages/about";
import CallbackComponent from "./pages/callback";
import LandingContactPage from "./pages/contact";
import Features from "./pages/features";
import Landing from "./pages/landing";
import PaymentSuccessPage from "./pages/payment-success";
import PermissionEdit from "./pages/permissions-edit";
import PermissionListPage from "./pages/permissions-list";
import PricingPage from "./pages/pricing";
import RoleEdit from "./pages/role-edit";
import RolesListPage from "./pages/roles-list";
import SignupPage from "./pages/signup";
import UserEdit from "./pages/user-edit";
import UserListPage from "./pages/user-list";

const queryClient = new QueryClient();
function App() {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
          <AuthProvider>
            <BrowserRouter>
              <Routes>
                <Route element={<RootLayout />}>
                  <Route path="/" element={<Landing />} />
                  <Route path="/home" element={<Landing />} />
                  <Route path="/features" element={<Features />} />
                  <Route path="/pricing" element={<PricingPage />} />
                  <Route path="/about" element={<LandingAboutPage />} />
                  <Route path="/contact" element={<LandingContactPage />} />
                  <Route path="/signin" element={<Signin />} />
                  <Route path="/signup" element={<SignupPage />} />
                </Route>
                {/* Other routes */}
                <Route path="/auth/callback" element={<CallbackComponent />} />
                <Route path="/payment" element={<AuthenticatedLayoutBase />}>
                  {/* /payment/success?sessionId */}
                  <Route path="success" element={<PaymentSuccessPage />} />
                </Route>
                <Route path="/dashboard" element={<AuthenticatedLayout />}>
                  <Route index element={<Dashboard />} />
                  <Route path="users">
                    <Route index element={<UserListPage />} />
                    <Route path=":userId" element={<UserEdit />} />
                  </Route>
                  <Route path="roles">
                    <Route index element={<RolesListPage />} />
                    <Route path=":roleId" element={<RoleEdit />} />
                  </Route>
                  <Route path="permissions">
                    <Route index element={<PermissionListPage />} />
                    <Route path=":permissionId" element={<PermissionEdit />} />
                  </Route>
                </Route>
              </Routes>
            </BrowserRouter>
          </AuthProvider>
          <Toaster />
        </ThemeProvider>
      </QueryClientProvider>
    </>
  );
}

export default App;
