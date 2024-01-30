package handler

import (
	"context"
	"fmt"
	"time"

	pb "github.com/YehyeokBang/Simple-SNS/pkg/api/v1/user"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	DB  *gorm.DB
	JWT *auth.JWT
}

func NewUserHandler(db *gorm.DB, jwt *auth.JWT) *UserHandler {
	return &UserHandler{
		DB:  db,
		JWT: jwt,
	}
}

func (h *UserHandler) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	hashedPassword, err := auth.HashPassword(req.GetPassword())
	if err != nil {
		return nil, err
	}

	var birthday *time.Time
	if req.GetBirthday() != nil {
		t := req.GetBirthday().AsTime()
		birthday = &t
	}

	user := db.User{
		UserId:    req.GetUserId(),
		Password:  hashedPassword,
		Name:      req.GetName(),
		Age:       req.GetAge(),
		Sex:       req.GetSex(),
		Birthday:  birthday,
		Introduce: req.GetIntroduce(),
	}

	result := h.DB.Create(&user)
	if result.Error != nil {
		if result.Error.Error() == "Error 1062: Duplicate entry" {
			return nil, status.Error(codes.AlreadyExists, "user id is already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.SignUpResponse{
		UserId: user.UserId,
	}, nil
}

func (h *UserHandler) LogIn(ctx context.Context, req *pb.LogInRequest) (*pb.LogInResponse, error) {
	var user db.User
	result := h.DB.Where("user_id = ?", req.GetUserId()).First(&user)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "user is not exists or password is not correct")
	}

	if !auth.CheckPasswordHash(req.GetPassword(), user.Password) {
		return nil, status.Error(codes.NotFound, "user is not exists or password is not correct")
	}

	userID := fmt.Sprintf("%d", user.ID)
	token, err := h.JWT.CreateToken(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	return &pb.LogInResponse{
		JwtToken: token,
	}, nil
}

func ExtractUserIDFromContext(ctx context.Context) (string, error) {
	userID := ctx.Value(auth.UserIDKey).(string)
	if userID == "" {
		return "", status.Error(codes.Unauthenticated, "user id is not provided")
	}

	return userID, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID, err := ExtractUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var user db.User
	result := h.DB.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, status.Error(codes.NotFound, "user is not exists")
	}

	var birthday *timestamppb.Timestamp
	if user.Birthday != nil {
		birthday = timestamppb.New(*user.Birthday)
	}

	return &pb.GetUserResponse{
		UserId:    user.UserId,
		Name:      user.Name,
		Age:       user.Age,
		Sex:       user.Sex,
		Birthday:  birthday,
		Introduce: user.Introduce,
	}, nil
}
