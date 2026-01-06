package main

import (
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/HuckOps/cert_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// RemoteCacheEntry Stores certificate information retrieved from a remote source.

type RemoteCacheEntry struct {
	certInfos []utils.CertificateInfo
	err       error
	elapsed   time.Duration
}

type LocalCacheEntry struct {
	certInfo utils.CertificateInfo
	domain   string
	err      error
}

// CertCollector implements the prometheus.Collector interface.
type CertCollector struct {
	certExpiryDays   *prometheus.Desc
	certValid        *prometheus.Desc
	certSubject      *prometheus.Desc
	certCheckStatus  *prometheus.Desc
	certCheckLatency *prometheus.Desc

	// remoteCache Stores certificate information retrieved from a remote source.
	remoteCache map[string]RemoteCacheEntry
	// localCache Stores certificate information retrieved from a local source.
	localCache map[string]LocalCacheEntry
	mutex      sync.RWMutex
	interval   time.Duration
	ticker     *time.Ticker
	wg         sync.WaitGroup
	shutdown   chan struct{}
}

// NewCertCollector Creates a new certificate collector with the specified interval.
func NewCertCollector() *CertCollector {
	interval := time.Duration(Cfg.Interval) * time.Second

	collector := &CertCollector{
		certCheckStatus: prometheus.NewDesc(
			"certificate_check_status",
			"Whether the certificate check was successful (1 = success, 0 = failure)",
			[]string{"source_type", "source"}, nil,
		),

		certExpiryDays: prometheus.NewDesc(
			"certificate_expiry_days",
			"Number of days until certificate expiry",
			[]string{"domain", "sn", "algorithm", "source_type", "source"}, nil,
		),
		certValid: prometheus.NewDesc(
			"certificate_valid",
			"Whether the certificate is valid (1 = valid, 0 = invalid)",
			[]string{"domain", "sn", "algorithm", "source_type", "source"}, nil,
		),
		certSubject: prometheus.NewDesc(
			"certificate_subject",
			"Certificate subject information",
			[]string{"domain", "sn", "algorithm", "subject", "source_type", "source"}, nil,
		),
		certCheckLatency: prometheus.NewDesc(
			"certificate_check_latency_milliseconds",
			"Time taken to check the certificate in milliseconds",
			[]string{"source_type", "source"}, nil,
		),
		// 初始化异步采集相关字段
		remoteCache: make(map[string]RemoteCacheEntry),
		localCache:  make(map[string]LocalCacheEntry),
		interval:    interval,
		shutdown:    make(chan struct{}),
	}

	// Start the first collection immediately
	collector.collectCertificates()

	// Start the background collection goroutine
	collector.startBackgroundCollection()

	return collector
}

// Describe Sends metric descriptors to the provided channel.
func (c *CertCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.certExpiryDays
	ch <- c.certValid
	ch <- c.certSubject
	ch <- c.certCheckStatus
	ch <- c.certCheckLatency
}

// startBackgroundCollection Starts the background collection goroutine
func (c *CertCollector) startBackgroundCollection() {
	c.ticker = time.NewTicker(c.interval)

	c.wg.Go(func() {
		for {
			select {
			case <-c.ticker.C:
				c.collectCertificates()
			case <-c.shutdown:
				c.ticker.Stop()
				return
			}
		}
	})
}

// Stop Stops the background collection goroutine
func (c *CertCollector) Stop() {
	close(c.shutdown)
	c.wg.Wait()
}

// collectCertificates Collects all certificate information
func (c *CertCollector) collectCertificates() {
	if len(Cfg.Remote) != 0 {
		go collectRemote(c)
	}
	if len(Cfg.LocalCerts) != 0 {
		go collectLocal(c)
	}
}

// collectRemote Collects all remote certificate information
func collectRemote(c *CertCollector) {
	// Check all certificates concurrently
	var wg sync.WaitGroup
	results := make(chan struct {
		remote    string
		elapsed   time.Duration
		certInfos []utils.CertificateInfo
		err       error
	}, len(Cfg.Remote))
	// Start a goroutine to check each certificate
	for _, remote := range Cfg.Remote {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			elapsed, certInfos, err := utils.CheckRemoteCertificate(r)

			if err == nil {
				zap.L().Debug("Checked certificate", zap.String("remote", r), zap.Any("certInfos", certInfos))
			} else {
				zap.L().Error("Error checking certificate", zap.String("remote", r), zap.Error(err))
			}
			// Always send the result to the channel
			results <- struct {
				remote    string
				elapsed   time.Duration
				certInfos []utils.CertificateInfo
				err       error
			}{r, elapsed, certInfos, err}
		}(remote)
	}

	// Wait for all goroutines to finish and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Update the cache with the results
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for result := range results {
		c.remoteCache[result.remote] = RemoteCacheEntry{
			certInfos: result.certInfos,
			err:       result.err,
			elapsed:   result.elapsed,
		}
	}
}

func collectLocal(c *CertCollector) {
	// Check all certificates concurrently
	var wg sync.WaitGroup
	results := make(chan struct {
		local    string
		certInfo utils.CertificateInfo
		err      error
	}, len(Cfg.LocalCerts))
	// Start a goroutine to check each certificate
	for _, local := range Cfg.LocalCerts {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			certInfo, err := utils.CheckLocalCertificate(l)
			result := struct {
				local    string
				certInfo utils.CertificateInfo
				err      error
			}{l, utils.CertificateInfo{}, err}

			if err == nil {
				zap.L().Debug("Checked certificate", zap.String("local", l), zap.Any("certInfo", certInfo))
				result.certInfo = *certInfo
			} else {
				zap.L().Error("Error checking certificate", zap.String("local", l), zap.Error(err))
			}
			// Always send the result to the channel
			results <- result
		}(local.PublicKeyPath)
	}
	// Wait for all goroutines to finish and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Update the cache with the results
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for result := range results {
		c.localCache[result.local] = LocalCacheEntry{
			certInfo: result.certInfo,
			err:      result.err,
		}
	}
}

// Collect Collects metrics from the cache and sends them to the provided channel.
func (c *CertCollector) Collect(ch chan<- prometheus.Metric) {

	// Read certificate information from the cache
	zap.L().Debug("Collecting certificate information cache in remote detection mode")
	c.mutex.RLock()
	remoteCacheCopy := make(map[string]RemoteCacheEntry, len(c.remoteCache))
	maps.Copy(remoteCacheCopy, c.remoteCache)
	c.mutex.RUnlock()

	// Send metrics for each certificate in the cache
	for remote, entry := range remoteCacheCopy {
		if entry.err != nil {
			// If certificate check failed, send valid status as 0
			ch <- prometheus.MustNewConstMetric(c.certValid, prometheus.GaugeValue, 0, "remote", remote)
			continue
		}
		// Send certificate check status metric
		ch <- prometheus.MustNewConstMetric(c.certCheckStatus, prometheus.GaugeValue, 1, "remote", remote)
		// Send certificate check latency metric
		ch <- prometheus.MustNewConstMetric(c.certCheckLatency, prometheus.GaugeValue, float64(entry.elapsed.Microseconds()), "remote", remote)

		for _, info := range entry.certInfos {
			fmt.Println(info.Domains)
			// Send certificate expiry metric
			ch <- prometheus.MustNewConstMetric(c.certExpiryDays, prometheus.GaugeValue, info.ExpiryDays, strings.Join(info.Domains, ","), info.SN, info.Algorithm, "remote", remote)
			// Send certificate validity metric
			ch <- prometheus.MustNewConstMetric(c.certValid, prometheus.GaugeValue, info.Valid, strings.Join(info.Domains, ","), info.SN, info.Algorithm, "remote", remote)
			// Send certificate subject metric with subject value
			ch <- prometheus.MustNewConstMetric(c.certSubject, prometheus.GaugeValue, 1, strings.Join(info.Domains, ","), info.SN, info.Algorithm, info.Subject, "remote", remote)
		}
	}

	// Read certificate information from the cache
	zap.L().Debug("Collecting certificate information cache in local detection mode")
	c.mutex.RLock()
	localCacheCopy := make(map[string]LocalCacheEntry, len(c.localCache))
	maps.Copy(localCacheCopy, c.localCache)
	c.mutex.RUnlock()

	// Send metrics for each certificate in the cache
	for local, entry := range localCacheCopy {
		if entry.err != nil {
			// If certificate check failed, send valid status as 0
			ch <- prometheus.MustNewConstMetric(c.certCheckStatus, prometheus.GaugeValue, 0, "local", local)
			continue
		}
		// Send certificate check status metric
		fmt.Println(strings.Join(entry.certInfo.Domains, ","))
		ch <- prometheus.MustNewConstMetric(c.certCheckStatus, prometheus.GaugeValue, 1, "local", local)
		// Send certificate expiry metric
		ch <- prometheus.MustNewConstMetric(c.certExpiryDays, prometheus.GaugeValue, entry.certInfo.ExpiryDays, strings.Join(entry.certInfo.Domains, ","), entry.certInfo.SN, entry.certInfo.Algorithm, "local", local)
		// Send certificate validity metric
		ch <- prometheus.MustNewConstMetric(c.certValid, prometheus.GaugeValue, entry.certInfo.Valid, strings.Join(entry.certInfo.Domains, ","), entry.certInfo.SN, entry.certInfo.Algorithm, "local", local)
		// Send certificate subject metric with subject value
		ch <- prometheus.MustNewConstMetric(c.certSubject, prometheus.GaugeValue, 1, strings.Join(entry.certInfo.Domains, ","), entry.certInfo.SN, entry.certInfo.Algorithm, entry.certInfo.Subject, "local", local)
	}

}
