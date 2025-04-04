import Dashboard from "@/pages/dashboard"; // Your protected page
import Signin from "@/pages/signin";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import AuthenticatedLayout from "./layouts/authenticated-layout";
import RootLayout from "./layouts/root";
import CallbackComponent from "./pages/callback";
import Landing from "./pages/landing";
import SignupPage from "./pages/signup";
import UserListPage from "./pages/user-list";

function App() {
  return (
    <>
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
                <Route path="users" element={<UserListPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </AuthProvider>
        <Toaster />
      </ThemeProvider>
    </>
  );
}

export default App;
