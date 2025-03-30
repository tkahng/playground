import ProtectedRoute from "@/components/protected-route";
import Dashboard from "@/pages/dashboard"; // Your protected page
import Login from "@/pages/login2";
import { BrowserRouter, Route, Routes } from "react-router";
import { Toaster } from "sonner";
import { AuthProvider } from "./context/auth-context";

function App() {
  return (
    <div>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
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
    </div>
  );
}

export default App;
