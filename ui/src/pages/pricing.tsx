import { useAuthProvider } from "@/hooks/use-auth-provider";

export default function Pricing() {
  const { user } = useAuthProvider();
  //   const
  return (
    <div>
      <h1 className="text-3xl font-bold tracking-tighter sm:text-5xl xl:text-6xl/none">
        Pricing
      </h1>
    </div>
  );
}
