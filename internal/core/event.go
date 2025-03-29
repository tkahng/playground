package core

import (
	"context"

	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/tools/hook"
)

type BaseEventData struct {
	tags []string
}

func (e *BaseEventData) Tags() []string {
	if e.tags == nil {
		return nil
	}

	return e.tags
}

type BaseEvent struct {
	App App
	hook.Event
	BaseEventData
	Context context.Context
	// Could be any of the ModelEventType* constants, like:
	// - create
	// - update
	// - delete
	// - validate
	Type string
}

type SignupEvent struct {
	App App
	hook.Event
	BaseEventData
	Context context.Context
	// Could be any of the ModelEventType* constants, like:
	// - create
	// - update
	// - delete
	// - validate
	Type string

	User    *models.User
	Account *models.UserAccount
}
