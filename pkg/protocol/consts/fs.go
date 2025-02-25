package consts

import "time"

const (
	// files bigger than this size are sent with sendfile
	MaxSmallFileSize = 2 * 4096

	// FSHandlerCacheDuration is the default expiration duration for inactive
	// file handlers opened by FS.
	FSHandlerCacheDuration = 10 * time.Second

	// FSCompressedFileSuffix is the suffix FS adds to the original file names
	// when trying to store compressed file under the new file name.
	// See FS.Compress for details.
	FSCompressedFileSuffix    = ".hertz.gz"
	FsMinCompressRatio        = 0.8
	FsMaxCompressibleFileSize = 8 * 1024 * 1024
)
