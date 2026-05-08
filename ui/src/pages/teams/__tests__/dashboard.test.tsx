import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";

vi.mock("@/hooks/use-team", () => ({
  useTeam: vi.fn(() => ({
    team: { id: "team-1", name: "Acme", slug: "acme" },
    setTeam: vi.fn(),
  })),
}));

vi.mock("@/lib/task-queries", () => ({
  taskProjectList: vi.fn(),
  deleteTaskProject: vi.fn(),
}));

vi.mock("../projects/create-project-dialog", () => ({
  CreateProjectDialog: () => <button>New Project</button>,
}));

vi.mock("../projects/create-project-ai-dialog", () => ({
  CreateProjectAiDialog: () => <button>AI Project</button>,
}));

vi.mock("../projects/project-card", () => ({
  ProjectCard: ({ project }: { project: { name: string; id: string } }) => (
    <div data-testid="project-card">{project.name}</div>
  ),
}));

vi.mock("@/components/data-table-pagination", () => ({
  DataTablePagination: () => <div data-testid="pagination" />,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { taskProjectList } from "@/lib/task-queries";
import TeamDashboard from "../dashboard";

function mockProjects(
  projects: { id: string; name: string; [key: string]: unknown }[] = [],
) {
  vi.mocked(taskProjectList).mockResolvedValue({
    data: projects,
    meta: { total: projects.length, page: 0, per_page: 6 },
  } as any);
}

function renderDashboard() {
  return render(
    <TeamDashboard />,
  );
}

describe("TeamDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("empty state — no projects", () => {
    it("shows 'No projects yet' heading", async () => {
      mockProjects([]);
      renderDashboard();
      await waitFor(() => {
        expect(screen.getByText("No projects yet")).toBeInTheDocument();
      });
    });

    it("shows descriptive empty-state text", async () => {
      mockProjects([]);
      renderDashboard();
      await waitFor(() => {
        expect(
          screen.getByText(/create a project to start managing tasks/i),
        ).toBeInTheDocument();
      });
    });

    it("shows create project buttons in the empty state", async () => {
      mockProjects([]);
      renderDashboard();
      // Both the header toolbar and the empty state render these buttons;
      // assert at least one of each is present.
      await waitFor(() => {
        expect(
          screen.getAllByRole("button", { name: /new project/i }).length,
        ).toBeGreaterThan(0);
        expect(
          screen.getAllByRole("button", { name: /ai project/i }).length,
        ).toBeGreaterThan(0);
      });
    });

    it("does not render project cards", async () => {
      mockProjects([]);
      renderDashboard();
      await waitFor(() => {
        expect(screen.queryByTestId("project-card")).not.toBeInTheDocument();
      });
    });
  });

  describe("populated state — projects exist", () => {
    const fakeProjects = [
      { id: "p-1", name: "Apollo", status: "todo", created_at: "", updated_at: "" },
      { id: "p-2", name: "Artemis", status: "todo", created_at: "", updated_at: "" },
    ];

    it("renders a card for each project", async () => {
      mockProjects(fakeProjects);
      renderDashboard();
      await waitFor(() => {
        expect(screen.getAllByTestId("project-card")).toHaveLength(2);
      });
    });

    it("shows project names", async () => {
      mockProjects(fakeProjects);
      renderDashboard();
      await waitFor(() => {
        expect(screen.getByText("Apollo")).toBeInTheDocument();
        expect(screen.getByText("Artemis")).toBeInTheDocument();
      });
    });

    it("does not show the empty state heading", async () => {
      mockProjects(fakeProjects);
      renderDashboard();
      await waitFor(() => {
        expect(screen.queryByText("No projects yet")).not.toBeInTheDocument();
      });
    });

    it("renders pagination", async () => {
      mockProjects(fakeProjects);
      renderDashboard();
      await waitFor(() => {
        expect(screen.getByTestId("pagination")).toBeInTheDocument();
      });
    });
  });

  describe("team header", () => {
    it("shows the team name", async () => {
      mockProjects([]);
      renderDashboard();
      await waitFor(() => {
        expect(screen.getByText("Acme")).toBeInTheDocument();
      });
    });
  });
});
