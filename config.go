package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.yaml.in/yaml/v2"
)

// Config 存储配置信息
type LocalCert struct {
	PublicKeyPath string `yaml:"public_key_path"`
	// PrivateKeyPath string `yaml:"private_key_path"`
}

type Config struct {
	LogLevel   string      `yaml:"log_level"`
	Interval   int         `yaml:"interval"`
	Remote     []string    `yaml:"remote"`
	LocalCerts []LocalCert `yaml:"local"`
}

var Cfg *Config

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
)

// LoadConfig 从文件加载配置
func LoadConfig() error {
	// 检查文件是否存在
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", *configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return err
	}

	// 初始化配置
	Cfg = &Config{}

	// 解析YAML
	if err = yaml.Unmarshal(data, Cfg); err != nil {
		return err
	}

	// 设置默认值
	if Cfg.LogLevel == "" {
		Cfg.LogLevel = "info"
	}
	if Cfg.Interval == 0 {
		Cfg.Interval = 60 // 默认60秒采集一次
	}
	if Cfg.Remote == nil {
		Cfg.Remote = []string{}
	}
	if Cfg.LocalCerts == nil {
		Cfg.LocalCerts = []LocalCert{}
	}

	zapLogLevel, err := zap.ParseAtomicLevel(Cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", Cfg.LogLevel)
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapLogLevel,
	))
	zap.ReplaceGlobals(logger)
	zap.L().Info("config loaded", zap.Any("config", Cfg))
	return nil
}
