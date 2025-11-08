package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sanket9162/grpc-api/internal/models"
	"github.com/sanket9162/grpc-api/internal/repositories/mongodb"
	pb "github.com/sanket9162/grpc-api/proto/gen"
	"github.com/sanket9162/grpc-api/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	for _, teacher := range req.GetExecs() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "request is in incorrect format: non-empty ID fields are not allowed")
		}
	}

	addedExec, err := mongodb.AddExecsToDb(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: addedExec}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	filter, err := filter(req.Exec, &models.Exec{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal err")
	}
	sortOption := sortOptions(req.GetSortBy())

	execs, err := mongodb.GetExecsFromDB(ctx, sortOption, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil
}

func (s *Server) UpdateExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {

	updatedExecs, err := mongodb.UpdateExecsInDB(ctx, req.Execs)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: updatedExecs}, nil
}

func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeachersConfirmation, error) {
	ids := req.GetIds()
	var teacherIdsToDelete []string
	for _, v := range ids {
		teacherIdsToDelete = append(teacherIdsToDelete, v.Id)
	}
	deletedIds, err := mongodb.DeleteTeacherFromDB(ctx, teacherIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteTeachersConfirmation{
		Status:     "Teachers successfully deleted",
		DeletedIds: deletedIds,
	}, nil

}

func (s *Server) Login(ctx context.Context, req *pb.ExecLoginRequest) (*pb.ExecLoginResponse, error) {
	exec, err := mongodb.GetUserByUsername(ctx, req)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	if exec.InactiveStatus {
		return nil, status.Error(codes.Unauthenticated, "Account is inactive")
	}

	err = utils.VerifyPassword(req.GetPassword(), exec.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "incorrect username/password")
	}

	tokenString, err := utils.SignToken(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Could not create token.")
	}

	return &pb.ExecLoginResponse{Status: true, Token: tokenString}, nil
}

func (s *Server) UpdatePasswrod(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	username, userRole, err := mongodb.UpdatePassowordInDB(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	token, err := utils.SignToken(req.Id, username, userRole)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Interanl error")
	}

	return &pb.UpdatePasswordResponse{
		PasswordUpdated: true,
		Token:           token,
	}, nil
}

func (s *Server) DeactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {
	result, err := mongodb.DeactivateUserInDB(ctx, req.GetIds())
	if err != nil {
		return nil, err
	}

	return &pb.Confirmation{
		Confirmation: result.ModifiedCount > 0,
	}, nil
}

func (s *Server) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPassowrdResponse, error) {
	email := req.GetEmail()

	message, err := mongodb.ForgotpasswordDb(ctx, email)
	if err != nil {
		return nil, err
	}

	return &pb.ForgotPassowrdResponse{
		Confiramtion: true,
		Message:      message,
	}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.Confirmation, error) {
	token := req.GetResetCode()

	if req.GetNew_Password() != req.GetConfirmPassword() {
		return nil, status.Error(codes.InvalidArgument, "passowrds do not match")
	}

	bytes, err := hex.DecodeString(token)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	hashedToken := sha256.Sum256(bytes)
	tokenInDb := hex.EncodeToString(hashedToken[:])

	err = mongodb.ResetPasswordDB(ctx, tokenInDb, req.GetNew_Password())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Confirmation{
		Confirmation: true,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.EmptyRequest) (*pb.ExecLogoutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthrized Access")
	}

	val, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorization Access")
	}

	token := strings.TrimPrefix(val[0], "Bearer ")

	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorization Access")
	}

	expiryTimeStamp := ctx.Value(utils.ContextKey("exporesAt"))
	expiryTimeStr := fmt.Sprintf("%v", expiryTimeStamp)

	expiryTimeint, err := strconv.ParseInt(expiryTimeStr, 10, 64)
	if err != nil {
		utils.ErrorHandler(err, "")
		return nil, status.Error(codes.Internal, "internal error")
	}

	expiryTime := time.Unix(expiryTimeint, 0)

	utils.JwtStore.AddToken(token, expiryTime)

	return &pb.ExecLogoutResponse{
		LoggedOut: true,
	}, nil
}
