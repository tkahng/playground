import ProtectedRoute from "@/components/protected-route";
import Dashboard from "@/pages/dashboard"; // Your protected page
import Login from "@/pages/login2";
import { BrowserRouter, Route, Routes } from "react-router";
import { AuthProvider } from "./context/auth-context";

function App() {
  return (
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
  );
}

export default App;
