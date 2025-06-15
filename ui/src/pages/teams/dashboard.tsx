import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useTeam } from "@/hooks/use-team";
import {
  Activity,
  Code,
  MessageSquare,
  MoreHorizontal,
  TrendingUp,
  UserPlus,
  Users,
  Zap,
} from "lucide-react";

type TeamMember = {
  id: string;
  name: string;
  email: string;
  role: string;
  avatar?: string;
  lastActive: string;
};

export default function TeamDashboard() {
  const { team, isLoading, error } = useTeam();
  const teamMembers: TeamMember[] = [
    {
      id: "1",
      name: "John Doe",
      email: "john@acme.com",
      role: "Owner",
      lastActive: "2 minutes ago",
    },
    {
      id: "2",
      name: "Jane Smith",
      email: "jane@acme.com",
      role: "Admin",
      lastActive: "1 hour ago",
    },
    {
      id: "3",
      name: "Bob Johnson",
      email: "bob@acme.com",
      role: "Member",
      lastActive: "3 hours ago",
    },
    {
      id: "4",
      name: "Alice Brown",
      email: "alice@acme.com",
      role: "Member",
      lastActive: "1 day ago",
    },
  ];

  const teamStats = {
    totalMembers: teamMembers.length,
    activeChats: 24,
    apiCalls: 15420,
    codeGenerations: 89,
    usageLimit: 25000,
    currentUsage: 15420,
  };
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!team) {
    return <div>No team selected</div>;
  }
  return (
    // <div className="flex flex-col min-h-screen">
    //   <main className="flex-1 py-6 bg-gray-50 dark:bg-gray-900">
    <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
      {/* <div className="space-y-6"> */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            {team.name} Dashboard
          </h1>
          <p className="text-muted-foreground">
            Manage your team's AI usage and collaboration
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {/* <Badge variant="secondary">{team.plan} Plan</Badge> */}
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Invite Member
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Team Members</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{teamStats.totalMembers}</div>
            <p className="text-xs text-muted-foreground">+2 from last month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Chats</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{teamStats.activeChats}</div>
            <p className="text-xs text-muted-foreground">+12% from last week</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">API Calls</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {teamStats.apiCalls.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">+8% from last month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Code Generated
            </CardTitle>
            <Code className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {teamStats.codeGenerations}
            </div>
            <p className="text-xs text-muted-foreground">
              +15% from last month
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Usage Progress */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Usage This Month</CardTitle>
          <CardDescription>
            {teamStats.currentUsage.toLocaleString()} of{" "}
            {teamStats.usageLimit.toLocaleString()} API calls used
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Progress
            value={(teamStats.currentUsage / teamStats.usageLimit) * 100}
            className="w-full"
          />
          <div className="flex justify-between text-sm text-muted-foreground mt-2">
            <span>
              {Math.round(
                (teamStats.currentUsage / teamStats.usageLimit) * 100
              )}
              % used
            </span>
            <span>
              {(teamStats.usageLimit - teamStats.currentUsage).toLocaleString()}{" "}
              remaining
            </span>
          </div>
        </CardContent>
      </Card>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="members">Members</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Recent Activity</CardTitle>
                <CardDescription>Latest team interactions</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[
                    {
                      user: "John Doe",
                      action: "Generated code snippet",
                      time: "2 minutes ago",
                    },
                    {
                      user: "Jane Smith",
                      action: "Started new chat session",
                      time: "1 hour ago",
                    },
                    {
                      user: "Bob Johnson",
                      action: "Used TextGenius model",
                      time: "3 hours ago",
                    },
                    {
                      user: "Alice Brown",
                      action: "Invited new member",
                      time: "1 day ago",
                    },
                  ].map((activity, index) => (
                    <div key={index} className="flex items-center">
                      <Activity className="h-4 w-4 mr-2 text-muted-foreground" />
                      <div className="flex-1">
                        <p className="text-sm font-medium">{activity.user}</p>
                        <p className="text-xs text-muted-foreground">
                          {activity.action}
                        </p>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {activity.time}
                      </span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Usage Trends</CardTitle>
                <CardDescription>API usage over time</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[
                    { period: "This week", usage: 3420, trend: "+12%" },
                    { period: "Last week", usage: 3050, trend: "+8%" },
                    { period: "2 weeks ago", usage: 2820, trend: "+15%" },
                    { period: "3 weeks ago", usage: 2450, trend: "+5%" },
                  ].map((period, index) => (
                    <div
                      key={index}
                      className="flex items-center justify-between"
                    >
                      <div>
                        <p className="text-sm font-medium">{period.period}</p>
                        <p className="text-xs text-muted-foreground">
                          {period.usage.toLocaleString()} calls
                        </p>
                      </div>
                      <div className="flex items-center">
                        <TrendingUp className="h-4 w-4 mr-1 text-green-500" />
                        <span className="text-sm text-green-500">
                          {period.trend}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="members" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Team Members</CardTitle>
              <CardDescription>
                Manage your team members and their roles
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {teamMembers.map((member) => (
                  <div
                    key={member.id}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center space-x-4">
                      <Avatar>
                        <AvatarImage
                          src={member.avatar || "/placeholder.svg"}
                        />
                        <AvatarFallback>
                          {member.name
                            .split(" ")
                            .map((n) => n[0])
                            .join("")}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="text-sm font-medium">{member.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {member.email}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Badge variant="outline">{member.role}</Badge>
                      <span className="text-xs text-muted-foreground">
                        {member.lastActive}
                      </span>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>Edit Role</DropdownMenuItem>
                          <DropdownMenuItem>View Activity</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-destructive">
                            Remove Member
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
            <CardFooter>
              <Button className="w-full">
                <UserPlus className="mr-2 h-4 w-4" />
                Invite New Member
              </Button>
            </CardFooter>
          </Card>
        </TabsContent>

        <TabsContent value="activity" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Activity Log</CardTitle>
              <CardDescription>
                Detailed activity history for your team
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {[
                  {
                    user: "John Doe",
                    action: "Generated 150 lines of Python code",
                    model: "CodeAssist",
                    time: "2 minutes ago",
                  },
                  {
                    user: "Jane Smith",
                    action: "Started chat session about API integration",
                    model: "TextGenius",
                    time: "1 hour ago",
                  },
                  {
                    user: "Bob Johnson",
                    action: "Processed image recognition task",
                    model: "ImageMaster",
                    time: "3 hours ago",
                  },
                  {
                    user: "Alice Brown",
                    action: "Generated voice synthesis",
                    model: "VoiceWizard",
                    time: "5 hours ago",
                  },
                  {
                    user: "John Doe",
                    action: "Analyzed data patterns",
                    model: "DataInsight",
                    time: "1 day ago",
                  },
                ].map((activity, index) => (
                  <div key={index} className="border-l-2 border-muted pl-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium">{activity.user}</p>
                        <p className="text-sm text-muted-foreground">
                          {activity.action}
                        </p>
                        <Badge variant="secondary" className="mt-1 text-xs">
                          {activity.model}
                        </Badge>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {activity.time}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="settings" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Team Settings</CardTitle>
              <CardDescription>Configure your team preferences</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4">
                <div className="grid grid-cols-3 items-center gap-4">
                  <label htmlFor="team-name" className="text-sm font-medium">
                    Team Name
                  </label>
                  <input
                    id="team-name"
                    defaultValue={team.name}
                    className="col-span-2 flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  />
                </div>
                <div className="grid grid-cols-3 items-center gap-4">
                  <label htmlFor="plan" className="text-sm font-medium">
                    Current Plan
                  </label>
                  <div className="col-span-2">
                    <Badge>{"hello"}</Badge>
                  </div>
                </div>
              </div>
            </CardContent>
            <CardFooter>
              <Button>Save Changes</Button>
            </CardFooter>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
    //    </main>

    // </div>
  );
}
