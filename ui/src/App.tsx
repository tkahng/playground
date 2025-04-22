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
import { BrowserRouter, Route, Routes } from "react-router";
import AuthVerify from "./components/auth-verify";
import {
  adminHeaderLinks,
  adminSidebarLinks,
  authenticatedSubHeaderLinks,
  protectedSidebarLinks,
  RouteLinks,
  settingsSidebarLinks,
} from "./components/landing-links";
import { Providers } from "./components/providers";
import NotFoundPage from "./pages/404";
import ConfirmPasswordReset from "./pages/auth/confirm-password-reset";
import ResetPasswordRequestPage from "./pages/auth/reset-password";
import NotAuthorizedPage from "./pages/not-authorized";
import ProjectEdit from "./pages/tasks/task-projects/project-edit";
import ProjectListPage from "./pages/tasks/task-projects/projects-list";
function App() {
  return (
    <>
      <Providers>
        <BrowserRouter>
          <AuthVerify />
          <Routes>
            <Route element={<RootLayout />}>
              <Route path="/" element={<Landing />} />
              <Route path="/home" element={<Landing />} />
              <Route path="/features" element={<Features />} />
              <Route path="/pricing" element={<PricingPage />} />
              <Route path="/about" element={<LandingAboutPage />} />
              <Route path="/contact" element={<LandingContactPage />} />
            </Route>
            {/* <Route element= */}
            {/* Other routes */}
            <Route element={<DashboardLayout />}>
              <Route path="/signin" element={<Signin />} />
              <Route path="/signup" element={<SignupPage />} />
              <Route path="/not-authorized" element={<NotAuthorizedPage />} />
              <Route path="/auth/callback" element={<CallbackComponent />} />
              <Route
                path="/auth/confirm-verification"
                element={<ConfirmVerification />}
              />
              <Route
                path="/password-reset"
                element={<ConfirmPasswordReset />}
              />
              <Route
                path="/forgot-password"
                element={<ResetPasswordRequestPage />}
              />
            </Route>
            <Route element={<AuthenticatedLayoutBase />}>
              <Route path="/payment">
                {/* /payment/success?sessionId */}
                <Route path="success" element={<PaymentSuccessPage />} />
              </Route>
              <Route
                path="/dashboard"
                element={
                  <DashboardLayout
                    // sidebarLinks={dashboardSidebarLinks}
                    headerLinks={authenticatedSubHeaderLinks}
                  />
                }
              >
                <Route index element={<Dashboard />} />
              </Route>

              <Route
                path="/dashboard/projects"
                element={
                  <DashboardLayout headerLinks={authenticatedSubHeaderLinks} />
                }
              >
                {/* <Route path="projects"> */}
                <Route index element={<ProjectListPage />} />
                <Route path=":projectId" element={<ProjectEdit />} />
                {/* </Route> */}
              </Route>
              <Route
                path="/dashboard/protected"
                element={
                  <DashboardLayout
                    sidebarLinks={protectedSidebarLinks}
                    headerLinks={authenticatedSubHeaderLinks}
                  />
                }
              >
                <Route path="basic" element={<BasicRoute />} />
                <Route path="pro" element={<ProRoute />} />
                <Route path="advanced" element={<AdvancedRoute />} />
              </Route>
              <Route
                path="/settings"
                element={
                  <DashboardLayout
                    sidebarLinks={settingsSidebarLinks}
                    headerLinks={authenticatedSubHeaderLinks}
                  />
                }
              >
                <Route path="profile" element={<ProfilePage />} />
                <Route path="account" element={<AccountSettingsPage />} />
                <Route path="billing" element={<BillingSettingPage />} />
              </Route>
            </Route>

            <Route path="/admin" element={<AdminLayoutBase />}>
              <Route
                path="dashboard"
                element={
                  <DashboardLayout
                    sidebarLinks={adminSidebarLinks}
                    headerLinks={adminHeaderLinks}
                    sidebarBackLink={RouteLinks.DASHBOARD_HOME}
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
                  <Route path=":permissionId" element={<PermissionEdit />} />
                </Route>
              </Route>
            </Route>
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </BrowserRouter>
      </Providers>
    </>
  );
}

export default App;
