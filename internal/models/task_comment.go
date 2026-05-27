package models

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

type TaskComment struct {
	_                 struct{}    `db:"task_comments" schema:"task" json:"-"`
	ID                uuid.UUID   `db:"id" json:"id"`
	TaskID            uuid.UUID   `db:"task_id" json:"task_id"`
	CreatedByMemberID uuid.UUID   `db:"created_by_member_id" json:"created_by_member_id"`
	Content           string      `db:"content" json:"content"`
	CreatedAt         time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time   `db:"updated_at" json:"updated_at"`
	CreatedByMember   *TeamMember `db:"created_by_member" src:"created_by_member_id" dest:"id" table:"org.team_members" json:"created_by_member,omitempty"`
	Task              *Task       `db:"task" src:"task_id" dest:"id" table:"task.tasks" json:"task,omitempty"`
}

var mentionRe = regexp.MustCompile(`@\[([^\]]+)\]\(([0-9a-f-]{36})\)`)

// ParseMentionedMemberIDs extracts unique team member UUIDs from @[Name](uuid) mention syntax.
func ParseMentionedMemberIDs(content string) []uuid.UUID {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	seen := make(map[uuid.UUID]struct{})
	result := make([]uuid.UUID, 0, len(matches))
	for _, m := range matches {
		id, err := uuid.Parse(m[2])
		if err != nil {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}
