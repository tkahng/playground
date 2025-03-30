import ProtectedRoute from "@/components/protected-route";
import Dashboard from "@/pages/dashboard"; // Your protected page
import Login from "@/pages/login";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { ThemeProvider } from "./components/theme-provider";
import { AuthProvider } from "./context/auth-context";
import RootLayout from "./layouts/root";
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
              </Route>
              <Route path="/" element={<Login />} />
              <Route path="/signin" element={<Login />} />
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <Dashboard />
                  </ProtectedRoute>
                }
              />
            </Routes>
          </BrowserRouter>
        </AuthProvider>
        <Toaster />
      </ThemeProvider>
    </>
  );
}

export default App;
