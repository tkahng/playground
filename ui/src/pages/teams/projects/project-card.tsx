import * as React from "react";
import { Link } from "react-router";
import { LayoutGrid, Clock, MoreHorizontal, Trash2 } from "lucide-react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Project, Team } from "@/schema.types";
import { ConfirmDialog } from "@/hooks/use-dialog";
import {
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import { useState } from "react";

interface ProjectCardProps {
  team: Team;
  project: Project;
  onDelete(projectId: string): void;
}

export function ProjectCard({ project, team, onDelete }: ProjectCardProps) {
  const [dropdownOpen, setDropdownOpen] = useState(false);

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const handleDelete = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onDelete(project.id);
  };

  return (
    <Link to={`/teams/${team?.slug}/projects/${project.id}`} className="block">
      <Card className="group cursor-pointer transition-all hover:shadow-md hover:border-primary/20">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-base leading-snug line-clamp-2 flex-1">
              {project.name}
            </CardTitle>
            <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
              <DropdownMenuTrigger asChild onClick={(e) => e.preventDefault()}>
                <Button variant="ghost" size="icon" className="size-8 shrink-0">
                  <MoreHorizontal className="size-4" />
                  <span className="sr-only">Open menu</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-40">
                {/* <DropdownMenuItem onClick={handleEdit}> */}
                {/*   <Pencil className="size-4" /> */}
                {/*   Edit */}
                {/* </DropdownMenuItem> */}
                {/* <DropdownMenuSeparator /> */}
                <DropdownMenuItem onClick={handleDelete} variant="destructive">
                  <Trash2 className="size-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <ConfirmDialog
              dialogProps={{
                open: deleteConfirmOpen,
                onOpenChange: setDeleteConfirmOpen,
              }}
            >
              <>
                <DialogHeader>
                  <DialogTitle>Are you absolutely sure?</DialogTitle>
                </DialogHeader>
                {/* Dialog Content */}
                <DialogDescription>
                  This action cannot be undone.
                </DialogDescription>
                <DialogFooter>
                  <DialogClose asChild>
                    <Button
                      variant="outline"
                      onClick={() => {
                        console.log("cancel");
                        // editDialog.props.onOpenChange(false);
                      }}
                    >
                      Cancel
                    </Button>
                  </DialogClose>
                  <DialogClose asChild>
                    <Button variant="destructive" onClick={handleDelete}>
                      Delete
                    </Button>
                  </DialogClose>
                </DialogFooter>
              </>
            </ConfirmDialog>
          </div>
          <CardDescription className="flex items-center gap-4 pt-1">
            <span className="flex items-center gap-1.5">
              <LayoutGrid className="size-3.5" />
              {project.status}
            </span>
            {/* <span className="flex items-center gap-1.5"> */}
            {/*   <Users className="size-3.5" /> */}
            {/*   {project.status} members */}
            {/* </span> */}
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <Clock className="size-3.5" />
            Updated {project.updated_at}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
