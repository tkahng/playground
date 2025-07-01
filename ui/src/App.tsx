import AdminLayoutBase from "@/layouts/admin-layout-base";
import AuthenticatedLayoutOutlet from "@/layouts/authenticated-layout-outlet";
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
import LandingAboutPage from "@/pages/landing/about";
import LandingContactPage from "@/pages/landing/contact";
import Features from "@/pages/landing/features";
import Landing from "@/pages/landing/landing";
import PricingPage from "@/pages/landing/pricing";
import PaymentSuccessPage from "@/pages/payment/payment-success";
import BillingSettingPage from "@/pages/settings/billing-settings";
import AccountSettingsPage from "@/pages/settings/general-settings";
import { BrowserRouter, Route, Routes } from "react-router";
import {
  adminHeaderLinks,
  authenticatedSubHeaderLinks,
  userDashboardLinks,
} from "./components/links";
import { Providers } from "./components/providers";
import { RouteMap } from "./components/route-map";
import AdminLayout from "./layouts/admin-layout";
import PageSectionLayout from "./layouts/page-section";
import PublicLayout from "./layouts/public-layout";
import TeamDashboardLayout from "./layouts/team-dashboard-layout";
import NotFoundPage from "./pages/404";
import AdminDashboardPage from "./pages/admin/admin-dashboard";
import ProductEditPage from "./pages/admin/products/products-edit";
import ProductsListPage from "./pages/admin/products/products-list";
import SubscriptionsListPage from "./pages/admin/subscriptions/subscription-list";
import ConfirmPasswordReset from "./pages/auth/confirm-password-reset";
import ResetPasswordRequestPage from "./pages/auth/reset-password";
import DashboardPage from "./pages/dashboard";
import NotAuthorizedPage from "./pages/not-authorized";
import ProtectedRouteLayout from "./pages/protected-routes/protected-layout";
import ProtectedRoutePage from "./pages/protected-routes/protected-route-page";
import ProtectedRouteIndex from "./pages/protected-routes/route-index";
import TeamDashboard from "./pages/teams/dashboard";
import ProjectEdit from "./pages/teams/projects/project-edit";
import ProjectListPage from "./pages/teams/projects/projects-list";
import TaskLayout from "./pages/teams/projects/task-layout";
import TeamBillingSettingPage from "./pages/teams/settings/team-billing-settings";
import TeamSettingsPage from "./pages/teams/settings/team-general-settings";
import TeamMembersSettingPage from "./pages/teams/settings/team-members-settings";
import TeamListPage from "./pages/teams/team-list";
import UserTeamInvitationRedirectPage from "./pages/teams/user-team-invitation-redirect-page";

function TeamRoutes() {
  return (
    <>
      <Route element={<AuthenticatedLayoutOutlet />}>
        <Route element={<TeamDashboardLayout />}>
          <Route
            path={`/teams/:teamSlug/dashboard`}
            element={<TeamDashboard />}
          />
          <Route element={<PageSectionLayout title="Projects" />}>
            <Route element={<TaskLayout />}>
              <Route
                path={`/teams/:teamSlug/projects`}
                element={<ProjectListPage />}
              />
              <Route
                path={`/teams/:teamSlug/projects/:projectId`}
                element={<ProjectEdit />}
              />
            </Route>
          </Route>

          <Route
            path={`/teams/:teamSlug/settings`}
            element={<PageSectionLayout title="Settings" />}
          >
            <Route index element={<TeamSettingsPage />} />
            <Route path="billing" element={<TeamBillingSettingPage />} />
            <Route path="members" element={<TeamMembersSettingPage />} />
          </Route>
        </Route>

        <Route
          path={`/protected`}
          element={
            <DashboardLayout headerLinks={authenticatedSubHeaderLinks} />
          }
        >
          <Route element={<PageSectionLayout title="Protected" />}>
            <Route element={<ProtectedRouteLayout />}>
              <Route index element={<ProtectedRouteIndex />} />
              <Route path=":permission" element={<ProtectedRoutePage />} />
            </Route>
          </Route>
        </Route>
      </Route>
    </>
  );
}

function App() {
  return (
    <>
      <Providers>
        <BrowserRouter>
          {/* <AuthVerify /> */}
          <Routes>
            {/* Landing page */}
            <Route element={<RootLayout />}>
              <Route path="/" element={<Landing />} />
              <Route path="/home" element={<Landing />} />
              <Route path="/features" element={<Features />} />
              <Route path="/pricing" element={<PricingPage />} />
              <Route path="/about" element={<LandingAboutPage />} />
              <Route path="/contact" element={<LandingContactPage />} />
            </Route>
            {/* Dashboard routes */}
            <Route element={<PublicLayout />}>
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

            <Route element={<AuthenticatedLayoutOutlet />}>
              <Route
                // dashboard
                children={
                  <Route path={`/dashboard`} element={<DashboardPage />} />
                }
                element={<DashboardLayout headerLinks={userDashboardLinks} />}
              />
              <Route path={RouteMap.PAYMENT}>
                <Route path="success" element={<PaymentSuccessPage />} />
              </Route>

              <Route
                // dashboard
                children={
                  <Route
                    path={`/team-invitation`}
                    element={<UserTeamInvitationRedirectPage />}
                  />
                }
                element={<DashboardLayout />}
              />
            </Route>

            <Route element={<AuthenticatedLayoutOutlet />}>
              <Route
                // dashboard
                path={`/teams`}
                element={<DashboardLayout headerLinks={userDashboardLinks} />}
              >
                <Route element={<PageSectionLayout title="Teams" />}>
                  <Route index element={<TeamListPage />} />
                </Route>
              </Route>
              <Route
                path={RouteMap.SETTINGS}
                element={<DashboardLayout headerLinks={userDashboardLinks} />}
              >
                <Route element={<PageSectionLayout title="Settings" />}>
                  <Route index element={<AccountSettingsPage />} />
                  <Route path="billing" element={<BillingSettingPage />} />
                </Route>
              </Route>
            </Route>
            {TeamRoutes()}
            <Route path={"/admin"} element={<AdminLayoutBase />}>
              <Route element={<AdminLayout headerLinks={adminHeaderLinks} />}>
                <Route index element={<AdminDashboardPage />} />
                <Route
                  path="users"
                  element={<PageSectionLayout title="Users" />}
                >
                  <Route index element={<UserListPage />} />
                  <Route path=":userId" element={<UserEdit />} />
                </Route>
                <Route
                  path="roles"
                  element={<PageSectionLayout title="Roles" />}
                >
                  <Route index element={<RolesListPage />} />
                  <Route path=":roleId" element={<RoleEdit />} />
                </Route>
                <Route
                  path="permissions"
                  element={<PageSectionLayout title="Permissions" />}
                >
                  <Route index element={<PermissionListPage />} />
                  <Route path=":permissionId" element={<PermissionEdit />} />
                </Route>
                <Route
                  path="subscriptions"
                  element={<PageSectionLayout title="Subscriptions" />}
                >
                  <Route index element={<SubscriptionsListPage />} />
                  <Route
                    path=":subscriptionId"
                    element={<div>Product Edit</div>}
                  />
                </Route>
                <Route
                  path="products"
                  element={<PageSectionLayout title="Products" />}
                >
                  <Route index element={<ProductsListPage />} />
                  <Route path=":productId" element={<ProductEditPage />} />
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
