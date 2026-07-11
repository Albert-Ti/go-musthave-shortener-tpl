package config

import (
	"flag"
	"os"
)

type Options struct {
	RunAddr         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	JWTSecretKey    string
	AuditFile       string
	AuditURL        string
}

// pattern Builder
func NewOptions(opts ...func(*Options)) *Options {
	o := &Options{
		RunAddr:      "localhost:8080",
		BaseURL:      "http://localhost:8080",
		JWTSecretKey: "jwt_secret_key",
	}

	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithRunAddr(v string) func(*Options)         { return func(o *Options) { o.RunAddr = v } }
func WithBaseURL(v string) func(*Options)         { return func(o *Options) { o.BaseURL = v } }
func WithFileStoragePath(v string) func(*Options) { return func(o *Options) { o.FileStoragePath = v } }
func WithDatabaseDSN(v string) func(*Options)     { return func(o *Options) { o.DatabaseDSN = v } }
func WithJWTSecretKey(v string) func(*Options)    { return func(o *Options) { o.JWTSecretKey = v } }
func WithAuditFile(v string) func(*Options)       { return func(o *Options) { o.AuditFile = v } }
func WithAuditURL(v string) func(*Options)        { return func(o *Options) { o.AuditURL = v } }

func Build() *Options {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var raw Options
	fs.StringVar(&raw.RunAddr, "a", "localhost:8080", "address and port to run server")
	fs.StringVar(&raw.BaseURL, "b", "http://localhost:8080", "Base URL")
	fs.StringVar(&raw.FileStoragePath, "f", "", "file storage")
	fs.StringVar(&raw.DatabaseDSN, "d", "", "connection string to DB")
	fs.StringVar(&raw.AuditFile, "audit-file", "", "путь к файлу-приёмнику")
	fs.StringVar(&raw.AuditURL, "audit-url", "", "URL удаленного сервера-приёмника")

	_ = fs.Parse(os.Args[1:])

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	pick := func(flagName string, flagVal string, envStr string) string {
		if explicit[flagName] {
			return flagVal
		}
		if v := os.Getenv(envStr); v != "" {
			return v
		}
		return flagVal
	}

	dsn := pick("d", raw.DatabaseDSN, "DATABASE_CONN_STRING")
	if explicit["f"] {
		dsn = raw.DatabaseDSN
	}

	return NewOptions(
		WithRunAddr(pick("a", raw.RunAddr, "SERVER_ADDRESS")),
		WithBaseURL(pick("b", raw.BaseURL, "BASE_URL")),
		WithFileStoragePath(pick("f", raw.FileStoragePath, "FILE_STORAGE_PATH")),
		WithDatabaseDSN(dsn),
		WithAuditFile(pick("audit-file", raw.AuditFile, "AUDIT_FILE")),
		WithAuditURL(pick("audit-url", raw.AuditURL, "AUDIT_URL")),
	)
}
