package user

import (
	"context"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/auth"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req *RegisterReq) error
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
}

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) Service {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, req *RegisterReq) error {
	exist, err := s.repo.GetUserByName(ctx, req.Username)
	if err == nil && exist != nil {
		return errno.ErrUserAlreadyExist
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		Username:     req.Username,
		PasswordHash: string(hashedPwd),
		Email:        req.Email,
	}
	return s.repo.RegisterUser(ctx, user)
}

func (s *userService) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	user, err := s.repo.GetUserByName(ctx, req.Username)
	if err != nil || user == nil {
		return nil, errno.ErrUserNotFound
	}

	if hashErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); hashErr != nil {
		return nil, errno.ErrPasswordIncorrect
	}
	expire := 24 * time.Hour
	token, err := auth.GenerateToken(user.ID, expire, user.Username)
	if err != nil {
		return nil, err
	}

	return &LoginResp{
		AccessToken: token,
		ExpiresIn:   int(expire.Seconds()),
		UserID:      user.ID,
		Username:    user.Username,
	}, nil
}
