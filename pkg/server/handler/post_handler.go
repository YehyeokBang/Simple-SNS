package handler

import (
	"context"
	"strconv"
	"time"

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
		Title:   req.GetTitle(),
		Content: req.GetContent(),
	}

	result := h.DB.Create(&post)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to write post")
	}

	h.DB.Joins("User").First(&post)

	return &pb.WritePostResponse{
		Message: "success posting",
	}, nil
}

func (h *PostHandler) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	type Result struct {
		PostID    uint
		UserName  string
		Title     string
		CreatedAt time.Time
		UpdatedAt time.Time
		Count     int64
	}

	var results []Result
	query := h.DB.Model(&db.Post{}).
		Select("posts.id as post_id, users.name as user_name, posts.title, posts.created_at, posts.updated_at, count(comments.id) as count").
		Joins("left join users on users.id = posts.user_id").
		Joins("left join comments on comments.post_id = posts.id").
		Group("posts.id, users.name, posts.content, posts.created_at, posts.updated_at").
		Order("posts.updated_at desc")

	if req.GetPage() != 0 {
		query = query.Limit(int(req.GetLimit())).Offset(int((req.GetPage() - 1) * req.GetLimit()))
	}
	result := query.Find(&results)

	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to get posts")
	}

	var pbPosts []*pb.PostSummary
	for _, result := range results {
		pbPosts = append(pbPosts, &pb.PostSummary{
			Id:           uint32(result.PostID),
			UserName:     result.UserName,
			Title:        result.Title,
			CommentCount: uint32(result.Count),
			CreatedAt:    timestamppb.New(result.CreatedAt),
			UpdatedAt:    timestamppb.New(result.UpdatedAt),
		})
	}

	return &pb.GetPostsResponse{
		PostSummaries: pbPosts,
	}, nil
}

func (h *PostHandler) GetPostById(ctx context.Context, req *pb.GetPostByIdRequest) (*pb.GetPostByIdResponse, error) {
	var post db.Post
	result := h.DB.Joins("User").Preload("Comments").Preload("Comments.User").First(&post, req.GetId())
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "post is not exists")
	}

	var pbComments []*pb.Comment
	for _, comment := range post.Comments {
		parentID := uint32(0)
		if comment.ParentCommentID != nil {
			parentID = uint32(*comment.ParentCommentID)
		}

		pbComments = append(pbComments, &pb.Comment{
			Id:        uint32(comment.ID),
			PostId:    uint32(comment.PostID),
			UserId:    uint32(comment.UserID),
			HasParent: comment.ParentCommentID != nil,
			ParentId:  parentID,
			Content:   comment.Content,
			UserName:  comment.User.Name,
			CreatedAt: timestamppb.New(comment.CreatedAt),
			UpdatedAt: timestamppb.New(comment.UpdatedAt),
		})
	}

	return &pb.GetPostByIdResponse{
		Post: &pb.Post{
			Id:        uint32(post.ID),
			UserName:  post.User.Name,
			Title:     post.Title,
			Content:   post.Content,
			Comments:  pbComments,
			CreatedAt: timestamppb.New(post.CreatedAt),
			UpdatedAt: timestamppb.New(post.UpdatedAt),
		},
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

	if req.GetTitle() != "" {
		post.Title = req.GetTitle()
	}

	if req.GetContent() != "" {
		post.Content = req.GetContent()
	}

	result = h.DB.Save(&post)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to update post")
	}

	var loadedPost db.Post
	h.DB.Preload("User").First(&loadedPost, post.ID)

	return &pb.UpdatePostResponse{
		Message: "success updating",
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
