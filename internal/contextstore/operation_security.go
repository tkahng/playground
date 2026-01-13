package contextstore

import (
	"context"
)

const (
	contextKeyOperationSecurity contextKey = "operation_security"
)

func SetContextOperationSecurity(ctx context.Context, info []map[string][]string) context.Context {
	return context.WithValue(ctx, contextKeyOperationSecurity, info)
}
func GetContextOperationSecurity(ctx context.Context) []map[string][]string {
	if team, ok := ctx.Value(contextKeyOperationSecurity).([]map[string][]string); ok {
		return team
	} else {
		return nil
	}
}
