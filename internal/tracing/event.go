//revive:disable:package-comments
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EventType identifies what triggered the event.
type EventType string

const (
	// EventTypeWebhook is a webhook-triggered event.
	EventTypeWebhook EventType = "webhook"

	// EventTypeManual is a manual retrieve event.
	EventTypeManual EventType = "manual"

	// EventTypeReplay is a replay event.
	EventTypeReplay EventType = "replay"
)

const (
	// EventSpanName is the root span name for all events.
	EventSpanName = "event"

	// EventTypeAttribute is the span attribute key for the event type.
	EventTypeAttribute = "event.type"

	// TenantAttribute is the span attribute key for the tenant namespace.
	TenantAttribute = "tenant"

	// WatchTargetSpanName is the span name for watch target processing.
	WatchTargetSpanName = "watch_target"

	// WatchTargetNameAttribute is the span attribute key for the ConfigMap name.
	WatchTargetNameAttribute = "watch_target.name"

	// WatchTargetRepoURLAttribute is the span attribute key for the repository URL.
	WatchTargetRepoURLAttribute = "watch_target.repo_url"

	// WatchTargetRefAttribute is the span attribute key for the git ref.
	WatchTargetRefAttribute = "watch_target.ref"

	// WatchTargetPathAttribute is the span attribute key for the path.
	WatchTargetPathAttribute = "watch_target.path"
)

// StartEvent creates a tracer, starts a root event span with the given type
// and namespace, and returns the context and tracer. The tracer is used by
// downstream watch target processors to create child spans.
func StartEvent(ctx context.Context, eventType EventType, namespace string) (context.Context, trace.Tracer) {
	tracer := otel.Tracer("phortizo")
	ctx, _ = tracer.Start(ctx, EventSpanName, trace.WithAttributes(
		attribute.String(EventTypeAttribute, string(eventType)),
		attribute.String(TenantAttribute, namespace),
	))

	return ctx, tracer
}
