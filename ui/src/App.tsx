import Dashboard from "@/pages/dashboard"; // Your protected page
import Signin from "@/pages/signin";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import AuthenticatedLayout from "./layouts/authenticated-layout";
import RootLayout from "./layouts/root";
import CallbackComponent from "./pages/callback";
import Landing from "./pages/landing";
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
                  <Route path="/signin" element={<Signin />} />
                  <Route path="/signup" element={<SignupPage />} />
                </Route>
                {/* Other routes */}
                <Route path="/auth/callback" element={<CallbackComponent />} />

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
