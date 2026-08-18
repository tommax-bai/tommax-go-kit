// Package objstore 封装 S3 兼容对象存储（线上 OSS S3 端点 / 本地 MinIO，同一套代码，见 docs/07）。
// 对象路径遵循 docs/04 §1.7: {env}/{domain}/{yyyy}/{mm}/{dd}/{id}.{ext}
package objstore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string `yaml:"endpoint"`  // 如 127.0.0.1:9000
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"useSSL"`
	// PublicBase 拼公开访问 URL（本地为 http://127.0.0.1:9000/{bucket}，线上为 CDN 域名）。
	PublicBase string `yaml:"publicBase"`
	Env        string `yaml:"env"` // dev / staging / prod，进对象路径
}

type Store struct {
	cli *minio.Client
	cfg Config
}

func New(cfg Config) (*Store, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("objstore init: %w", err)
	}
	return &Store{cli: cli, cfg: cfg}, nil
}

// Key 生成规范化对象路径。
func (s *Store) Key(domain, id, ext string) string {
	now := time.Now().UTC()
	return fmt.Sprintf("%s/%s/%04d/%02d/%02d/%s.%s",
		s.cfg.Env, domain, now.Year(), now.Month(), now.Day(), id, ext)
}

// Put 上传对象并返回其 key。
func (s *Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.cli.PutObject(ctx, s.cfg.Bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("objstore put %s: %w", key, err)
	}
	return nil
}

// PublicURL 返回对象公开访问地址（dev 桶开放下载；线上应换签名 URL）。
func (s *Store) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.cfg.PublicBase, key)
}

// PresignGet 返回限时签名 URL。
func (s *Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.cli.PresignedGetObject(ctx, s.cfg.Bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("objstore presign %s: %w", key, err)
	}
	return u.String(), nil
}
