import { ThemeProvider } from "@/components/theme-provider";
import { AuthProvider } from "@/context/auth-context";
import SettingLayout from "@/layouts/account-setting-layout";
import AdminDashboardLayout from "@/layouts/admin-dashboard-layout";
import AdminLayoutBase from "@/layouts/admin-layout-base";
import AuthenticatedLayoutBase from "@/layouts/authenticated-layout-base";
import DashboardLayout from "@/layouts/dashboard-layout";
import RootLayout from "@/layouts/root";
import PermissionEdit from "@/pages/admin/permissions/permissions-edit";
import PermissionListPage from "@/pages/admin/permissions/permissions-list";
import RoleEdit from "@/pages/admin/roles/role-edit";
import RolesListPage from "@/pages/admin/roles/roles-list";
import UserEdit from "@/pages/admin/users/user-edit";
import UserListPage from "@/pages/admin/users/user-list";
import CallbackComponent from "@/pages/auth/callback";
import ConfirmVerification from "@/pages/auth/confirm-verification";
import Signin from "@/pages/auth/signin";
import SignupPage from "@/pages/auth/signup";
import Dashboard from "@/pages/dashboard"; // Your protected page
import LandingAboutPage from "@/pages/landing/about";
import LandingContactPage from "@/pages/landing/contact";
import Features from "@/pages/landing/features";
import Landing from "@/pages/landing/landing";
import PricingPage from "@/pages/landing/pricing";
import PaymentSuccessPage from "@/pages/payment/payment-success";
import ProfilePage from "@/pages/profile";
import AdvancedRoute from "@/pages/protected-routes/route-advanced";
import BasicRoute from "@/pages/protected-routes/route-basic";
import ProRoute from "@/pages/protected-routes/route-pro";
import AccountSettingsPage from "@/pages/settings/account-settings";
import BillingSettingPage from "@/pages/settings/billing-settings";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { BrowserRouter, Link, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import {
  adminHeaderLinks,
  adminSidebarLinks,
  dashboardSidebarLinks,
  protectedSidebarLinks,
  tasksSidebarLinks,
} from "./components/landing-links";
import { RouteMap } from "./components/route-map";
import ProjectEdit from "./pages/tasks/task-projects/project-edit";
import ProjectListPage from "./pages/tasks/task-projects/projects-list";

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
                <Route
                  path="/auth/confirm-verification"
                  element={<ConfirmVerification />}
                />
                <Route element={<AuthenticatedLayoutBase />}>
                  <Route path="/payment">
                    {/* /payment/success?sessionId */}
                    <Route path="success" element={<PaymentSuccessPage />} />
                  </Route>
                  <Route
                    path="/dashboard"
                    element={<DashboardLayout links={dashboardSidebarLinks} />}
                  >
                    <Route index element={<Dashboard />} />
                  </Route>

                  <Route
                    path="/dashboard/tasks"
                    element={
                      <DashboardLayout
                        links={tasksSidebarLinks}
                        backLink={
                          <>
                            <Link
                              to={RouteMap.DASHBOARD_HOME}
                              className="flex items-center gap-2 text-sm text-muted-foreground"
                            >
                              <ChevronLeft className="h-4 w-4" />
                              Back to Dashboard
                            </Link>
                          </>
                        }
                      />
                    }
                  >
                    <Route path="projects">
                      <Route index element={<ProjectListPage />} />
                      <Route path=":projectId" element={<ProjectEdit />} />
                    </Route>
                  </Route>
                  <Route
                    path="/dashboard/protected"
                    element={
                      <DashboardLayout
                        links={protectedSidebarLinks}
                        backLink={
                          <>
                            <Link
                              to={RouteMap.DASHBOARD_HOME}
                              className="flex items-center gap-2 text-sm text-muted-foreground"
                            >
                              <ChevronLeft className="h-4 w-4" />
                              Back to Dashboard
                            </Link>
                          </>
                        }
                      />
                    }
                  >
                    <Route path="basic" element={<BasicRoute />} />
                    <Route path="pro" element={<ProRoute />} />
                    <Route path="advanced" element={<AdvancedRoute />} />
                  </Route>
                  <Route path="/settings" element={<SettingLayout />}>
                    <Route path="profile" element={<ProfilePage />} />
                    <Route path="account" element={<AccountSettingsPage />} />
                    <Route path="billing" element={<BillingSettingPage />} />
                  </Route>
                </Route>

                <Route path="/admin" element={<AdminLayoutBase />}>
                  <Route
                    path="dashboard"
                    element={
                      <AdminDashboardLayout
                        links={adminSidebarLinks}
                        headerLinks={adminHeaderLinks}
                      />
                    }
                  >
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
