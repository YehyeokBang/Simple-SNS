package handler

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/YehyeokBang/Simple-SNS/pkg/api/v1/comment"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type CommentHandler struct {
	pb.UnimplementedCommentServiceServer
	DB  *gorm.DB
	JWT *auth.JWT
}

func NewCommentHandler(db *gorm.DB, jwt *auth.JWT) *CommentHandler {
	return &CommentHandler{
		DB:  db,
		JWT: jwt,
	}
}

func (h *CommentHandler) WriteComment(ctx context.Context, req *pb.WriteCommentRequest) (*pb.WriteCommentResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	comment := db.Comment{
		UserID:  uint(userIDUint),
		PostID:  uint(req.GetPostId()),
		Content: req.GetContent(),
	}

	user := db.User{}
	h.DB.First(&user, comment.UserID)
	if user.ID == 0 {
		return nil, status.Error(codes.NotFound, "user is not exists")
	}

	result := h.DB.Create(&comment)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to write comment")
	}

	return &pb.WriteCommentResponse{
		Comment: &pb.Comment{
			Id:        uint32(comment.ID),
			PostId:    uint32(comment.PostID),
			UserName:  user.Name,
			Content:   comment.Content,
			CreatedAt: timestamppb.New(comment.CreatedAt),
			UpdatedAt: timestamppb.New(comment.UpdatedAt),
		},
	}, nil
}

func (h *CommentHandler) WriteReply(ctx context.Context, req *pb.WriteReplyRequest) (*pb.WriteReplyResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	parentComment := db.Comment{}
	h.DB.First(&parentComment, req.GetParentCommentId())

	// 부모 댓글이 이미 대댓글인 경우 에러를 반환
	if parentComment.ParentCommentID != nil {
		return nil, status.Error(codes.InvalidArgument, "nested replies are not allowed")
	}

	parentCommentID := uint(req.GetParentCommentId())
	reply := db.Comment{
		UserID:          uint(userIDUint),
		PostID:          uint(req.GetPostId()),
		ParentCommentID: &parentCommentID,
		Content:         req.GetContent(),
		ParentComment:   &parentComment,
	}

	user := db.User{}
	h.DB.First(&user, reply.UserID)
	if user.ID == 0 {
		return nil, status.Error(codes.NotFound, "user is not exists")
	}

	result := h.DB.Create(&reply)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to write reply")
	}

	return &pb.WriteReplyResponse{
		Reply: &pb.Comment{
			Id:        uint32(reply.ID),
			PostId:    uint32(reply.PostID),
			UserName:  user.Name,
			Content:   reply.Content,
			CreatedAt: timestamppb.New(reply.CreatedAt),
			UpdatedAt: timestamppb.New(reply.UpdatedAt),
		},
	}, nil
}

func (h *CommentHandler) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	var comment db.Comment
	result := h.DB.Where("id = ?", req.GetCommentId()).First(&comment)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "comment is not exists")
	}

	if comment.UserID != uint(userIDUint) {
		return nil, status.Error(codes.PermissionDenied, "you don't have permission")
	}

	user := h.DB.First(&db.User{}, comment.UserID)
	if user.Error != nil {
		return nil, status.Error(codes.NotFound, "user is not exists")
	}

	comment.Content = req.GetContent()
	result = h.DB.Save(&comment)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to update comment")
	}

	return &pb.UpdateCommentResponse{
		Comment: &pb.Comment{
			Id:        uint32(comment.ID),
			PostId:    uint32(comment.PostID),
			Content:   comment.Content,
			UserName:  user.Name(),
			CreatedAt: timestamppb.New(comment.CreatedAt),
			UpdatedAt: timestamppb.New(comment.UpdatedAt),
		},
	}, nil
}

func (h *CommentHandler) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	var comment db.Comment
	result := h.DB.Where("id = ?", req.GetCommentId()).First(&comment)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "comment is not exists")
	}

	if comment.UserID != uint(userIDUint) {
		return nil, status.Error(codes.PermissionDenied, "you don't have permission")
	}

	result = h.DB.Delete(&comment)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to delete comment")
	}

	return &pb.DeleteCommentResponse{
		Message: fmt.Sprintf("comment %d is deleted", comment.ID),
	}, nil
}
