package service

import (
	"fmt"
	"time"

	"pickup/internal/config"
	"pickup/internal/model"
	"pickup/internal/repository"
	"pickup/internal/utils"

	"go.uber.org/zap"
)

// AuthService 认证服务接口
type AuthService interface {
	WechatLogin(code, phoneCode string) (*WechatLoginResponse, error)
	GetUserInfo(userID uint) (*model.User, error)
	UpdateLastLogin(userID uint) error
}

// WechatLoginResponse 微信登录响应
type WechatLoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// authService 认证服务实现
type authService struct {
	userRepo     repository.UserRepository
	wechatClient *utils.WechatClient
	jwtUtil      *utils.JWTUtil
	cryptoUtil   *utils.CryptoUtil
	logger       *zap.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	wechatConfig *config.WechatConfig,
	jwtConfig *config.JWTConfig,
	cryptoConfig *config.CryptoConfig,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		wechatClient: utils.NewWechatClient(wechatConfig.AppID, wechatConfig.AppSecret),
		jwtUtil:      utils.NewJWTUtil(jwtConfig.Secret, jwtConfig.ExpireTime, jwtConfig.Issuer),
		cryptoUtil:   utils.NewCryptoUtil(cryptoConfig.Key),
		logger:       logger,
	}
}

// WechatLogin 微信登录
func (s *authService) WechatLogin(code, phoneCode string) (*WechatLoginResponse, error) {
	// 1. 调用微信接口获取openid和session_key
	sessionResp, err := s.wechatClient.JSCode2Session(code)
	if err != nil {
		s.logger.Error("failed to get wechat session", zap.Error(err))
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 2. 获取手机号（这里简化处理，实际需要access_token）
	// 在实际项目中，需要先获取access_token，然后用phone_code获取手机号
	// 这里假设phoneCode就是解密后的手机号
	phone := phoneCode // 简化处理

	// 3. 查找或创建用户
	user, err := s.userRepo.GetByOpenID(sessionResp.OpenID)
	if err != nil {
		// 用户不存在，创建新用户
		user = &model.User{
			OpenID:   sessionResp.OpenID,
			UnionID:  sessionResp.UnionID,
			Phone:    phone,
			Nickname: "微信用户",
			Role:     model.RolePassenger,
			Status:   model.UserStatusActive,
		}

		if err := s.userRepo.Create(user); err != nil {
			s.logger.Error("failed to create user", zap.Error(err))
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
	} else {
		// 更新用户信息
		user.UnionID = sessionResp.UnionID
		user.Phone = phone
		user.LastLoginAt = &time.Time{}
		*user.LastLoginAt = time.Now()

		if err := s.userRepo.Update(user); err != nil {
			s.logger.Error("failed to update user", zap.Error(err))
			return nil, fmt.Errorf("更新用户信息失败: %w", err)
		}
	}

	// 4. 生成JWT token
	token, err := s.jwtUtil.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 5. 更新最后登录时间
	if err := s.userRepo.UpdateLastLoginAt(user.ID); err != nil {
		s.logger.Warn("failed to update last login time", zap.Error(err))
	}

	s.logger.Info("user login successful", zap.Uint("user_id", user.ID))

	return &WechatLoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *authService) GetUserInfo(userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

// UpdateLastLogin 更新最后登录时间
func (s *authService) UpdateLastLogin(userID uint) error {
	return s.userRepo.UpdateLastLoginAt(userID)
}
