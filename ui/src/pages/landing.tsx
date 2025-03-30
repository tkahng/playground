import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function Landing() {
  return (
    <div className="flex h-screen items-center justify-center">
      <Card className="max-w-[700px]">
        <CardHeader>
          <CardTitle>Card Title</CardTitle>
          <CardDescription>Card Description</CardDescription>
        </CardHeader>
        <CardContent>
          <p>Card Content</p>
        </CardContent>
        <CardFooter>
          <CardAction>Card Action</CardAction>
        </CardFooter>
      </Card>
    </div>
  );
}
