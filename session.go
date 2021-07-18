package main

import (
	"context"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	ignore "github.com/sabhiram/go-gitignore"
)

type Session struct {
	Project  *Project
	S3       *s3.Client
	Uploader *manager.Uploader
	Ignore   *ignore.GitIgnore
}

func NewSession(p *Project) (*Session, error) {
	s := &Session{
		Project: p,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	region, err := manager.GetBucketRegion(context.TODO(), client, p.Bucket)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	s.S3 = s3.NewFromConfig(cfg)
	s.Uploader = manager.NewUploader(s.S3)

	if len(p.Ignore) != 0 {
		s.Ignore = ignore.CompileIgnoreLines(p.Ignore...)
	}

	return s, nil
}

func (s *Session) Run() error {
	return s.scan("")
}

func (s *Session) scan(start string) error {
	if s.Ignore.MatchesPath(start) {
		return nil
	}

	list, err := ioutil.ReadDir(path.Join(s.Project.Root, start))
	if err != nil {
		return err
	}

	for _, info := range list {
		path := path.Join(start, info.Name())

		if info.IsDir() {
			err = s.scan(path + "/")
		} else {
			err = s.backup(path, info)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) backup(file string, info fs.FileInfo) error {
	if s.Ignore.MatchesPath(file) {
		return nil
	}

	duration := startedAt.Sub(info.ModTime())
	if duration.Hours() < 24*7 {
		log.Printf("skipping %s: too young", file)
		return nil
	}

	key := path.Join(s.Project.Prefix, file)
	res, err := s.S3.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.Project.Bucket),
		Key:    aws.String(key),
	})
	if err == nil && !info.ModTime().After(*res.LastModified) {
		log.Printf("skipping %s: not modified", file)
		return nil
	}

	log.Printf("uploading %s", file)

	f, err := os.Open(path.Join(s.Project.Root, file))
	if err != nil {
		return err
	}
	defer f.Close()

	req := &s3.PutObjectInput{
		Bucket:        aws.String(s.Project.Bucket),
		Key:           aws.String(key),
		Body:          f,
		ContentLength: info.Size(),
		StorageClass:  types.StorageClassDeepArchive,
	}
	if req.ContentLength < 100*1024*1024 {
		_, err = s.S3.PutObject(context.TODO(), req)
	} else {
		_, err = s.Uploader.Upload(context.TODO(), req)
	}
	if err != nil {
		return err
	}

	return nil
}
