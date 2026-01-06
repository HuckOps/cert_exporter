package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	listenAddr  = flag.String("web.listen-address", ":9101", "Address to listen on for web interface and telemetry.")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
)

func main() {
	flag.Parse()

	// 加载配置
	if err := LoadConfig(); err != nil {
		zap.L().Fatal("Error loading config", zap.Error(err))
	}

	// 创建并注册收集器
	certCollector := NewCertCollector()
	prometheus.MustRegister(certCollector)

	// 设置HTTP路由
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Certificate Exporter</title></head>
			<body>
				<h1>Certificate Exporter</h1>
				<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
		</html>`))
	})

	zap.L().Info("Starting server", zap.String("addr", *listenAddr))
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		zap.L().Fatal("Error starting server", zap.Error(err))
	}
}
