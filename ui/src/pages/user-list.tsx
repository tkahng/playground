// import { RouteMap } from "@/components/route-map";
// import { useAuthProvider } from "@/hooks/use-auth-provider";
// import { userPaginate } from "@/lib/api";
// import { useEffect, useState } from "react";
// import { useNavigate, useSearchParams } from "react-router";

// export default function UserListPage() {
//   const { user } = useAuthProvider();
//   const [loading, setLoading] = useState(false);
//   const navigate = useNavigate(); // Get navigation function
//   const [searchParams, setSearchParams] = useSearchParams();

//   if (!user) {
//     navigate(RouteMap.SIGNIN);
//     return;
//   }
//   useEffect(() => {
//     const fetch = async () => {
//       try {
//         const response = await userPaginate(user.tokens.refresh_token, {});
//         setUsers(data);
//       } catch (error) {
//         console.error("Error fetching users:", error);
//       }
//     };

//     return () => {
//       second;
//     };
//   }, [third]);
// }
