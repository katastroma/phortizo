package resolve

import (
	"context"
	"fmt"
	"time"
)

type stubExchanger struct {
	token string
	err   error
}

func (s *stubExchanger) InstallationToken(_ context.Context, _ int64) (string, time.Time, error) {
	return s.token, time.Time{}, s.err
}

var errExchange = fmt.Errorf("exchange failed")

type stubFactory struct {
	token string
	err   error
}

func (f *stubFactory) Create(_ string, _ []byte) TokenExchanger {
	return &stubExchanger{token: f.token, err: f.err}
}
