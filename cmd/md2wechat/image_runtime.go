package main

import (
	"github.com/geekjourneyx/md2wechat-skill/internal/image"
	"github.com/geekjourneyx/md2wechat-skill/internal/wechat"
)

func newRuntimeImageProcessor() *image.Processor {
	svc := wechat.NewService(cfg, log)
	return image.NewProcessor(
		cfg,
		log,
		image.WithDownloadFunc(wechat.DownloadFile),
		image.WithUploadFunc(func(filePath string) (*image.UploadResult, error) {
			result, err := svc.UploadMaterialWithRetry(filePath, 3)
			if err != nil {
				return nil, err
			}
			return &image.UploadResult{
				MediaID:   result.MediaID,
				WechatURL: result.WechatURL,
				Width:     result.Width,
				Height:    result.Height,
			}, nil
		}),
	)
}
