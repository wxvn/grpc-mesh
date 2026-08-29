package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wxvn/grpc-mesh/internal/common/interceptors"
	"github.com/wxvn/grpc-mesh/internal/url/domain"
	"github.com/wxvn/grpc-mesh/internal/url/transport"
	pb "github.com/wxvn/grpc-mesh/proto/url"
)

type Server struct {
	pb.UnimplementedURLShortenerServiceServer
	service      transport.URLService
	publicScheme string
	publicHost   string
	publicPort   string
}

func NewServer(service transport.URLService, publicScheme string, publicHost string, publicPort string) *Server {
	return &Server{
		service:      service,
		publicScheme: publicScheme,
		publicHost:   publicHost,
		publicPort:   publicPort,
	}
}

func (s *Server) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetUrl() == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"url is required",
		)
	}

	result, err := s.service.CreateShortURL(
		ctx,
		userID,
		req.GetUrl(),
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return s.toURLResponse(result), nil
}

func (s *Server) GetMyURLs(ctx context.Context, req *pb.GetMyURLsRequest) (*pb.GetMyURLsResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	limit := int(req.GetLimit())
	offset := int(req.GetOffset())

	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	urls, err := s.service.GetMyURLs(
		ctx,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	response := &pb.GetMyURLsResponse{
		Urls: make([]*pb.CreateShortURLResponse, 0, len(urls)),
	}

	for _, url := range urls {
		response.Urls = append(
			response.Urls,
			s.toURLResponse(url),
		)
	}

	return response, nil
}

func (s *Server) GetURLStats(ctx context.Context, req *pb.GetURLStatsRequest) (*pb.GetURLStatsResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid url id",
		)
	}

	result, err := s.service.GetURLStats(
		ctx,
		userID,
		id,
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.GetURLStatsResponse{
		Id:          result.ID.String(),
		ShortCode:   result.ShortCode,
		ShortUrl:    s.buildShortURL(result.ShortCode),
		OriginalUrl: result.OriginalURL,
		Clicks:      result.Clicks,
		CreatedAt:   result.CreatedAt.Format(timeFormat),
	}, nil
}

func (s *Server) DeleteURL(ctx context.Context, req *pb.DeleteURLRequest) (*pb.DeleteURLResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid url id",
		)
	}

	if err := s.service.DeleteURL(
		ctx,
		userID,
		id,
	); err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.DeleteURLResponse{
		Success: true,
	}, nil
}

func getUserID(ctx context.Context) (uuid.UUID, error) {
	userID := interceptors.UserIDFromContext(ctx)

	if userID == "" {
		return uuid.Nil, status.Error(
			codes.Unauthenticated,
			"user id not found",
		)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, status.Error(
			codes.Unauthenticated,
			"invalid user id",
		)
	}

	return id, nil
}

func (s *Server) toURLResponse(url *domain.URL) *pb.CreateShortURLResponse {
	return &pb.CreateShortURLResponse{
		Id:          url.ID.String(),
		ShortCode:   url.ShortCode,
		ShortUrl:    s.buildShortURL(url.ShortCode),
		OriginalUrl: url.OriginalURL,
		Clicks:      url.Clicks,
		CreatedAt:   url.CreatedAt.Format(timeFormat),
	}
}

func (s *Server) buildShortURL(shortCode string) string {
	baseURL := s.publicScheme + "://" + s.publicHost

	if s.publicPort != "" {
		baseURL += ":" + s.publicPort
	}

	return baseURL + "/" + shortCode
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
