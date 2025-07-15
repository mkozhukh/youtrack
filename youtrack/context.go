package youtrack

import (
	"context"
)

type YouTrackContext struct {
	ctx    context.Context
	APIKey string
}

func NewYouTrackContext(ctx context.Context, apiKey string) *YouTrackContext {
	return &YouTrackContext{
		ctx:    ctx,
		APIKey: apiKey,
	}
}

func (y *YouTrackContext) Context() context.Context {
	return y.ctx
}