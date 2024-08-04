package backend

import (
	"log"
	"sync/atomic"
	"time"
)

type Metrics struct {
	requestCount       uint64
	bytesSent          uint64
	bytesReceived      uint64
	totalBytesSent     uint64
	totalBytesReceived uint64
	lastResetTime      time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		lastResetTime: time.Now(),
	}
}

func (m *Metrics) IncrementRequests() {
	atomic.AddUint64(&m.requestCount, 1)
}

func (m *Metrics) AddBytesSent(bytes uint64) {
	atomic.AddUint64(&m.bytesSent, bytes)
}

func (m *Metrics) AddBytesReceived(bytes uint64) {
	atomic.AddUint64(&m.bytesReceived, bytes)
}

func (m *Metrics) AddTotalBytesSent(bytes uint64) {
	atomic.AddUint64(&m.totalBytesSent, bytes)
}

func (m *Metrics) AddTotalBytesReceived(bytes uint64) {
	atomic.AddUint64(&m.totalBytesReceived, bytes)
}

func (m *Metrics) GetTotalTraffic() (totalSent, totalReceived uint64) {
	totalSent = atomic.LoadUint64(&m.totalBytesSent)
	totalReceived = atomic.LoadUint64(&m.totalBytesReceived)
	return
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	sentBytes := uint64(n)
	w.metrics.AddBytesSent(sentBytes)
	w.metrics.AddTotalBytesSent(sentBytes)
	return n, err
}

func (m *Metrics) GetAndReset() (requests, bytesSent, bytesReceived uint64, duration time.Duration) {
	now := time.Now()
	duration = now.Sub(m.lastResetTime)

	requests = atomic.SwapUint64(&m.requestCount, 0)
	bytesSent = atomic.SwapUint64(&m.bytesSent, 0)
	bytesReceived = atomic.SwapUint64(&m.bytesReceived, 0)

	m.lastResetTime = now
	return
}

func (p *ReverseProxy) LogMetrics() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		requests, bytesSent, bytesReceived, duration := p.metrics.GetAndReset()
		totalSent, totalReceived := p.metrics.GetTotalTraffic()
		log.Printf("Requests/s: %.2f, Traffic sent: %d bytes/s, Traffic received: %d bytes/s, Total sent: %d bytes, Total received: %d bytes",
			float64(requests)/duration.Seconds(),
			uint64(float64(bytesSent)/duration.Seconds()),
			uint64(float64(bytesReceived)/duration.Seconds()),
			totalSent,
			totalReceived)
	}
}
