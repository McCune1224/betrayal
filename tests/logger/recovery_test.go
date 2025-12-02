package logger

import (
	"errors"
	"time"

	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/rs/zerolog"
)

// TestRecoverWithLogNoPanic tests recovery when no panic occurs
func (lts *LoggerTestSuite) TestRecoverWithLogNoPanic() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	func() {
		defer logger.RecoverWithLog(*log)
		// No panic
	}()
}

// TestRecoverWithLogHandlesPanic tests recovery catches and logs panic
func (lts *LoggerTestSuite) TestRecoverWithLogHandlesPanic() {
	cfg := logger.Config{
		Environment: "production",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	// This should not crash the test
	func() {
		defer logger.RecoverWithLog(*log)
		panic("test panic")
	}()
}

// TestRecoverWithLogRePanicsInProduction tests that production re-panics
func (lts *LoggerTestSuite) TestRecoverWithLogRePanicsInProduction() {
	cfg := logger.Config{
		Environment: "production",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	// Lock down the log level
	logger.SetLevel(zerolog.InfoLevel)

	// In production, we expect a re-panic
	defer func() {
		if r := recover(); r == nil {
			lts.Fail("Expected re-panic in production")
		}
	}()

	func() {
		defer logger.RecoverWithLog(*log)
		panic("test panic")
	}()
}

// TestSafeGoSuccess tests SafeGo with successful function
func (lts *LoggerTestSuite) TestSafeGoSuccess() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan bool)

	logger.SafeGo(*log, "test_goroutine", func() error {
		done <- true
		return nil
	})

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		lts.Fail("SafeGo goroutine timed out")
	}
}

// TestSafeGoWithError tests SafeGo handles function errors
func (lts *LoggerTestSuite) TestSafeGoWithError() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan bool)

	logger.SafeGo(*log, "test_goroutine_error", func() error {
		done <- true
		return errors.New("test error")
	})

	select {
	case <-done:
		// Success - error was logged
	case <-time.After(2 * time.Second):
		lts.Fail("SafeGo goroutine timed out")
	}
}

// TestSafeGoWithPanic tests SafeGo recovers from panics
func (lts *LoggerTestSuite) TestSafeGoWithPanic() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan bool)

	logger.SafeGo(*log, "test_goroutine_panic", func() error {
		defer func() { done <- true }()
		panic("test panic in goroutine")
	})

	select {
	case <-done:
		// Success - panic was recovered
	case <-time.After(2 * time.Second):
		lts.Fail("SafeGo goroutine timed out")
	}
}

// TestSafeGoVoidSuccess tests SafeGoVoid with successful function
func (lts *LoggerTestSuite) TestSafeGoVoidSuccess() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan bool)

	logger.SafeGoVoid(*log, "test_void_goroutine", func() {
		done <- true
	})

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		lts.Fail("SafeGoVoid goroutine timed out")
	}
}

// TestSafeGoVoidWithPanic tests SafeGoVoid recovers from panics
func (lts *LoggerTestSuite) TestSafeGoVoidWithPanic() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan bool)

	logger.SafeGoVoid(*log, "test_void_panic", func() {
		defer func() { done <- true }()
		panic("test panic in void goroutine")
	})

	select {
	case <-done:
		// Success - panic was recovered
	case <-time.After(2 * time.Second):
		lts.Fail("SafeGoVoid goroutine timed out")
	}
}

// TestSafeGoMultipleConcurrent tests multiple concurrent SafeGo calls
func (lts *LoggerTestSuite) TestSafeGoMultipleConcurrent() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	done := make(chan int, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			logger.SafeGo(*log, "concurrent_task", func() error {
				done <- idx
				return nil
			})
		}(i)
	}

	count := 0
	timeout := time.After(5 * time.Second)
	for count < 10 {
		select {
		case <-done:
			count++
		case <-timeout:
			lts.Fail("SafeGo concurrent goroutines timed out")
			return
		}
	}
}
