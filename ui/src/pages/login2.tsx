import { AuthContext } from "@/context/auth-context";
import { SigninInput } from "@/schema.types";
import { useContext, useState } from "react";
import { useNavigate } from "react-router";

export default function Login() {
  const [input, setInput] = useState<SigninInput>({ email: "", password: "" });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate(); // Get navigation function
  const { login } = useContext(AuthContext);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    setLoading(true);

    // Simulate API call (replace with actual authentication)
    try {
      await login({ email: input.email, password: input.password });
      setLoading(false);
      navigate("/dashboard");
    } catch (error) {
      setError("Invalid email or password");
      setLoading(false);
    }
  };
  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const key = e.target.id;
    const value = e.target.value;
    setInput((values) => ({
      ...values,
      [key]: value,
    }));
  }
  return (
    <div>
      <h2>Login</h2>
      {error && <p style={{ color: "red" }}>{error}</p>}

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="email">Email:</label>
          <input
            id="email"
            type="email"
            name="email"
            value={input.email}
            onChange={handleChange}
            required
          />
        </div>
        <div>
          <label htmlFor="password">Password:</label>
          <input
            id="password"
            type="password"
            name="password"
            value={input.password}
            onChange={handleChange}
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? "Logging in..." : "Login"}
        </button>
      </form>
    </div>
  );
}
