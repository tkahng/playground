import { getTeamTeamMembers } from "@/lib/team-queries";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import {
  MentionsInput,
  Mention,
  type SuggestionDataItem,
  type OnChangeHandlerFunc,
} from "react-mentions";
import { cn } from "@/lib/utils";

interface MentionTextareaProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
}

const mentionStyle = {
  control: {
    fontSize: 14,
    lineHeight: 1.5,
  },
  "&multiLine": {
    control: {
      minHeight: 80,
    },
    highlighter: {
      padding: "8px 12px",
      border: "1px solid transparent",
    },
    input: {
      padding: "8px 12px",
      border: "1px solid hsl(var(--border))",
      borderRadius: "calc(var(--radius) - 2px)",
      outline: "none",
      background: "hsl(var(--background))",
      color: "hsl(var(--foreground))",
    },
  },
  suggestions: {
    list: {
      background: "hsl(var(--popover))",
      border: "1px solid hsl(var(--border))",
      borderRadius: "calc(var(--radius) - 2px)",
      fontSize: 14,
      boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
      maxHeight: 200,
      overflowY: "auto" as const,
      zIndex: 50,
    },
    item: {
      padding: "6px 12px",
      cursor: "pointer",
      "&focused": {
        background: "hsl(var(--accent))",
        color: "hsl(var(--accent-foreground))",
      },
    },
  },
};

export function MentionTextarea({
  value,
  onChange,
  placeholder,
  className,
  disabled,
}: MentionTextareaProps) {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();

  const fetchMembers = async (
    search: string,
    callback: (data: SuggestionDataItem[]) => void,
  ) => {
    if (!user?.tokens.access_token || !teamMember?.team_id) {
      callback([]);
      return;
    }
    try {
      const result = await getTeamTeamMembers({
        token: user.tokens.access_token,
        teamId: teamMember.team_id,
        search,
        perPage: 20,
        active: true,
      });
      const items: SuggestionDataItem[] = (result?.data ?? []).map((m) => ({
        id: m.id,
        display: m.user?.name ?? m.user?.email ?? m.id,
      }));
      callback(items);
    } catch {
      callback([]);
    }
  };

  const handleChange: OnChangeHandlerFunc = (_e, newValue) => {
    onChange(newValue);
  };

  return (
    <MentionsInput
      value={value}
      onChange={handleChange}
      placeholder={placeholder}
      disabled={disabled}
      style={mentionStyle}
      className={cn(className)}
      a11ySuggestionsListLabel="Suggested team members"
    >
      <Mention
        trigger="@"
        data={fetchMembers}
        markup="@[__display__](__id__)"
        displayTransform={(_id, display) => `@${display}`}
        renderSuggestion={(suggestion) => (
          <span>{suggestion.display}</span>
        )}
      />
    </MentionsInput>
  );
}
