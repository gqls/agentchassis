package datahelpers

import (
	"time"

	"go.uber.org/zap"
)

// DuplicateDetectionLogger provides structured logging for duplicate detection
type DuplicateDetectionLogger struct {
	logger *zap.Logger
}

func NewDuplicateDetectionLogger(logger *zap.Logger) *DuplicateDetectionLogger {
	return &DuplicateDetectionLogger{
		logger: logger.With(zap.String("subsystem", "duplicate_detection")),
	}
}

func (d *DuplicateDetectionLogger) LogRequestReceived(requestID, orchestrationID, podName string) {
	d.logger.Info("REQUEST_RECEIVED",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", orchestrationID),
		zap.String("pod_name", podName),
		zap.Time("timestamp", time.Now()),
	)
}

func (d *DuplicateDetectionLogger) LogClaimAttempt(requestID, podName string) {
	d.logger.Info("CLAIM_ATTEMPT",
		zap.String("request_id", requestID),
		zap.String("pod_name", podName),
		zap.Time("timestamp", time.Now()),
	)
}

func (d *DuplicateDetectionLogger) LogClaimSuccess(requestID, orchestrationID, podName string) {
	d.logger.Info("CLAIM_SUCCESS",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", orchestrationID),
		zap.String("pod_name", podName),
		zap.Time("timestamp", time.Now()),
	)
}

func (d *DuplicateDetectionLogger) LogDuplicateDetected(requestID, podName string) {
	d.logger.Warn("DUPLICATE_DETECTED",
		zap.String("request_id", requestID),
		zap.String("pod_name", podName),
		zap.Time("timestamp", time.Now()),
	)
}

func (d *DuplicateDetectionLogger) LogProcessingComplete(requestID, orchestrationID string, duration time.Duration) {
	d.logger.Info("PROCESSING_COMPLETE",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", orchestrationID),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	)
}
