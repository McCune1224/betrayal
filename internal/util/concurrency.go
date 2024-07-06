package util

import "context"

// Helper to run a database query in a goroutine and return the result
// this should be run in a goroutine and the resultChan should be buffered
func DbTask[T any](ctx context.Context, resultChan chan T, dbFunc func() (T, error)) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	select {
	case <-ctx.Done():
		close(resultChan)
		cancel()
	default:
		result, err := dbFunc()
		if err != nil {
			close(resultChan)
			cancel()
			return
		}
		resultChan <- result
		return
	}
}
