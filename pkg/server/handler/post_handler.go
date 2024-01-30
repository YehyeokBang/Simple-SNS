package handler

import (
	"context"
	"strconv"

	pb "github.com/YehyeokBang/Simple-SNS/pkg/api/v1/post"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type PostHandler struct {
	pb.UnimplementedPostServiceServer
	DB  *gorm.DB
	JWT *auth.JWT
}

func NewPostHandler(db *gorm.DB, jwt *auth.JWT) *PostHandler {
	return &PostHandler{
		DB:  db,
		JWT: jwt,
	}
}

func (h *PostHandler) WritePost(ctx context.Context, req *pb.WritePostRequest) (*pb.WritePostResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	post := db.Post{
		UserID:  uint(userIDUint),
		Content: req.GetContent(),
	}

	result := h.DB.Create(&post)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to write post")
	}

	var loadedPost db.Post
	h.DB.Preload("User").First(&loadedPost, post.ID)

	return &pb.WritePostResponse{
		Post: &pb.Post{
			Id:        uint32(post.ID),
			UserName:  post.User.Name,
			Content:   post.Content,
			CreatedAt: timestamppb.New(post.CreatedAt),
			UpdatedAt: timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

func (h *PostHandler) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	var posts []db.Post
	if req.GetPage() != 0 {
		result := h.DB.Preload("User").Limit(int(req.Limit)).Offset(int(req.Limit*req.Page - 1)).Find(&posts)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, "failed to get posts")
		}
	} else {
		result := h.DB.Preload("User").Find(&posts)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, "failed to get posts")
		}
	}

	var pbPosts []*pb.Post
	for _, post := range posts {
		pbPosts = append(pbPosts, &pb.Post{
			Id:        uint32(post.ID),
			UserName:  post.User.Name,
			Content:   post.Content,
			CreatedAt: timestamppb.New(post.CreatedAt),
			UpdatedAt: timestamppb.New(post.UpdatedAt),
		})
	}

	return &pb.GetPostsResponse{
		Posts: pbPosts,
	}, nil
}

func (h *PostHandler) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	var post db.Post
	result := h.DB.Where("id = ?", req.GetId()).First(&post)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "post is not exists")
	}

	if post.UserID != uint(userIDUint) {
		return nil, status.Error(codes.PermissionDenied, "you don't have permission")
	}

	post.Content = req.GetContent()
	result = h.DB.Save(&post)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to update post")
	}

	var loadedPost db.Post
	h.DB.Preload("User").First(&loadedPost, post.ID)

	return &pb.UpdatePostResponse{
		Post: &pb.Post{
			Id:        uint32(loadedPost.ID),
			UserName:  loadedPost.User.Name,
			Content:   loadedPost.Content,
			CreatedAt: timestamppb.New(loadedPost.CreatedAt),
			UpdatedAt: timestamppb.New(loadedPost.UpdatedAt),
		},
	}, nil
}

func (h *PostHandler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	var post db.Post
	result := h.DB.Where("id = ?", req.GetId()).First(&post)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "post is not exists")
	}

	if post.UserID != uint(userIDUint) {
		return nil, status.Error(codes.PermissionDenied, "you don't have permission")
	}

	result = h.DB.Delete(&post)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to delete post")
	}

	return &pb.DeletePostResponse{
		Status: true,
	}, nil
}
