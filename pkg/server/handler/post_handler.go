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

func Paginate(req *pb.GetPostsRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.GetPage() != 0 {
			offset := int((req.GetPage() - 1) * req.GetLimit())
			return db.Offset(offset).Limit(int(req.GetLimit()))
		}
		return db
	}
}

func (h *PostHandler) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	var posts []db.Post
	result := h.DB.Scopes(Paginate(req)).
		Preload("User").
		Preload("Comments").
		Find(&posts)

	if result.Error != nil {
		return nil, status.Error(codes.Internal, "failed to get posts")
	}

	var pbPosts []*pb.PostSummary
	for _, post := range posts {
		commentCount := len(post.Comments)
		pbPosts = append(pbPosts, &pb.PostSummary{
			Id:           uint32(post.ID),
			UserName:     post.User.Name,
			Title:        post.Title,
			CommentCount: uint32(commentCount),
			CreatedAt:    timestamppb.New(post.CreatedAt),
			UpdatedAt:    timestamppb.New(post.UpdatedAt),
		})
	}

	return &pb.GetPostsResponse{
		PostSummaries: pbPosts,
	}, nil
}

func sortComments(comments []*pb.Comment) []*pb.Comment {
	normalComments := make([]*pb.Comment, 0)
	replyComments := make(map[uint32][]*pb.Comment)

	for _, comment := range comments {
		if comment.HasParent {
			replyComments[comment.ParentId] = append(replyComments[comment.ParentId], comment)
		} else {
			normalComments = append(normalComments, comment)
		}
	}

	sortedComments := make([]*pb.Comment, 0)
	for _, comment := range normalComments {
		sortedComments = append(sortedComments, comment)
		sortedComments = append(sortedComments, replyComments[comment.Id]...)
	}

	return sortedComments
}

func (h *PostHandler) GetPostById(ctx context.Context, req *pb.GetPostByIdRequest) (*pb.GetPostByIdResponse, error) {
	var post db.Post
	result := h.DB.Joins("User").
		Preload("Comments").
		Preload("Comments.User").
		First(&post, req.GetId())

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
			Comments:  sortComments(pbComments),
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
