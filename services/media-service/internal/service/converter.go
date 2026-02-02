package service

import (
	"github.com/zunokit/zuno-marketplace-api/services/media-service/internal/client"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
)

// MediaResponseToProto converts client.MediaResponse to pb.MediaInfo
func MediaResponseToProto(result *client.MediaResponse) *pb.MediaInfo {
	if result == nil {
		return nil
	}

	return &pb.MediaInfo{
		Id:           result.ID,
		FileName:     result.FileName,
		FileSize:     result.FileSize,
		MimeType:     result.MimeType,
		MediaType:    result.MediaType,
		Url:          result.URL,
		ThumbnailUrl: result.ThumbnailURL,
		Width:        int32(result.Width),
		Height:       int32(result.Height),
		IpfsHash:     result.IPFSHash,
		IpfsUrl:      result.IPFSURL,
		IsPinned:     result.IsPinned,
		PinnedAt:     result.PinnedAt,
		CreatedAt:    result.CreatedAt,
	}
}

// MediaResponsesToProto converts a slice of client.MediaResponse to pb.MediaInfo
func MediaResponsesToProto(results []*client.MediaResponse) []*pb.MediaInfo {
	if results == nil {
		return nil
	}

	mediaInfos := make([]*pb.MediaInfo, 0, len(results))
	for _, result := range results {
		mediaInfos = append(mediaInfos, MediaResponseToProto(result))
	}

	return mediaInfos
}

// ExtractFileData extracts file data and filenames from pb.FileData slice
func ExtractFileData(files []*pb.FileData) ([][]byte, []string) {
	fileData := make([][]byte, 0, len(files))
	filenames := make([]string, 0, len(files))

	for _, f := range files {
		fileData = append(fileData, f.Data)
		filenames = append(filenames, f.Filename)
	}

	return fileData, filenames
}
