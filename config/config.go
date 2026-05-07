package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	DB      DBConfig      `yaml:"db"`
	Redis   RedisConfig   `yaml:"redis"`
	JWT     JWTConfig     `yaml:"jwt"`
	Wechat  WechatConfig  `yaml:"wechat"`
	Alipay  AlipayConfig  `yaml:"alipay"`
	Payment PaymentConfig `yaml:"payment"`
}

type PaymentConfig struct {
	Sandbox bool `yaml:"sandbox"` // 沙盒模式，所有支付直接返回成功
}

type AlipayConfig struct {
	AppID      string `yaml:"app_id"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
	NotifyURL  string `yaml:"notify_url"`
}

type ServerConfig struct {
	Port             int      `yaml:"port"`
	Mode             string   `yaml:"mode"` // debug, release
	SqlConsole       bool     `yaml:"sql_console"`
	TrustedProxies   []string `yaml:"trusted_proxies"`
	CorsOrigins      []string `yaml:"cors_origins"`
	MetricsPath      string   `yaml:"metrics_path"`       // 如 /internal/metrics，空则关闭 Prometheus
	RateLimitBackend string   `yaml:"rate_limit_backend"` // auto | redis | memory — auto 在 Redis 可用时用集群限流
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // hours
}

type WechatConfig struct {
	AppID      string `yaml:"app_id"`
	AppSecret  string `yaml:"app_secret"` // 小程序密钥
	MchID      string `yaml:"mch_id"`
	MchAPIKey  string `yaml:"mch_api_key"` // APIv3 密钥
	SerialNo   string `yaml:"serial_no"`   // 商户证书序列号
	PrivateKey string `yaml:"private_key"` // 商户私钥文件路径
	NotifyURL  string `yaml:"notify_url"`  // 支付回调地址
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
