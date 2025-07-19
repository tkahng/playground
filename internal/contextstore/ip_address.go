package contextstore

import "context"

const (
	contextKeyIPAddress contextKey = "ip_address"
)

func SetContextIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, contextKeyIPAddress, ip)
}

func GetContextIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value(contextKeyIPAddress).(string); ok {
		return ip
	} else {
		return ""
	} // ctx.Value(contextKeyIPAddress).(string)
}
