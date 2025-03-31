import ProtectedRoute from "@/components/protected-route";
import Dashboard from "@/pages/dashboard"; // Your protected page
import Login from "@/pages/login";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import RootLayout from "./layouts/root";
import CallbackComponent from "./pages/callback";
import Landing from "./pages/landing";

function App() {
  return (
    <>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route element={<RootLayout />}>
                <Route path="/" element={<Landing />} />
                <Route path="/signin" element={<Login />} />
                <Route path="/signup" element={<Login />} />
              </Route>
              {/* Other routes */}
              <Route path="/auth/callback" element={<CallbackComponent />} />

              <Route element={<RootLayout />}>
                <Route
                  path="/dashboard"
                  element={
                    <ProtectedRoute>
                      <Dashboard />
                    </ProtectedRoute>
                  }
                />
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
