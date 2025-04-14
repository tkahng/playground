import Dashboard from "@/pages/dashboard"; // Your protected page
import Signin from "@/pages/signin";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import SettingLayout from "./layouts/account-setting-layout";
import AdminDashboardLayout from "./layouts/admin-dashboard-layout";
import AdminLayoutBase from "./layouts/admin-layout-base";
import AuthenticatedLayoutBase from "./layouts/authenticated-layout-base";
import DashboardLayout from "./layouts/dashboard-layout";
import RootLayout from "./layouts/root";
import LandingAboutPage from "./pages/about";
import AccountSettingsPage from "./pages/account-settings";
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
                {/* <Route element= */}
                {/* Other routes */}
                <Route path="/auth/callback" element={<CallbackComponent />} />
                <Route element={<AuthenticatedLayoutBase />}>
                  <Route path="/payment">
                    {/* /payment/success?sessionId */}
                    <Route path="success" element={<PaymentSuccessPage />} />
                  </Route>
                  <Route path="/dashboard" element={<DashboardLayout />}>
                    <Route index element={<Dashboard />} />
                  </Route>
                  <Route path="/settings" element={<SettingLayout />}>
                    <Route path="account" element={<AccountSettingsPage />} />
                  </Route>
                </Route>

                <Route path="/admin" element={<AdminLayoutBase />}>
                  <Route path="dashboard" element={<AdminDashboardLayout />}>
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
                      <Route
                        path=":permissionId"
                        element={<PermissionEdit />}
                      />
                    </Route>
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
