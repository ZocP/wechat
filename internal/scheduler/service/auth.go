package service

import (
	"fmt"
	"strings"

	"pickup/internal/config"
	"pickup/internal/scheduler/models"
	"pickup/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	db           *gorm.DB
	wechatClient *utils.WechatClient
	jwtUtil      *utils.JWTUtil
	adminPhone   string
	adminOpenID  string
	logger       *zap.Logger
}

func NewAuthService(db *gorm.DB, wechatCfg *config.WechatConfig, jwtCfg *config.JWTConfig, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:           db,
		wechatClient: utils.NewWechatClient(wechatCfg.AppID, wechatCfg.AppSecret),
		jwtUtil:      utils.NewJWTUtil(jwtCfg.Secret, jwtCfg.ExpireTime, jwtCfg.Issuer),
		adminPhone:   normalizePhone(wechatCfg.AdminPhone),
		adminOpenID:  strings.TrimSpace(wechatCfg.AdminOpenID),
		logger:       logger,
	}
}

type LoginResult struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func normalizePhone(phone string) string {
	value := strings.TrimSpace(phone)
	if strings.HasPrefix(value, "+86") {
		value = strings.TrimSpace(strings.TrimPrefix(value, "+86"))
	}
	return value
}

func (s *AuthService) LoginWithWechatCode(code string) (*LoginResult, error) {
	session, err := s.wechatClient.JSCode2Session(code)
	if err != nil {
		return nil, fmt.Errorf("wechat login failed: %w", err)
	}

	var user models.User
	err = s.db.Where("open_id = ?", session.OpenID).First(&user).Error
	isAdminByOpenID := s.adminOpenID != "" && session.OpenID == s.adminOpenID
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		role := models.UserRoleStudent
		if isAdminByOpenID {
			role = models.UserRoleAdmin
		}
		user = models.User{
			OpenID: session.OpenID,
			Name:   "wx_user",
			Role:   role,
		}
		if createErr := s.db.Create(&user).Error; createErr != nil {
			return nil, createErr
		}
	} else if isAdminByOpenID && user.Role != models.UserRoleAdmin {
		user.Role = models.UserRoleAdmin
		if updateErr := s.db.Model(&models.User{}).Where("id = ?", user.ID).Update("role", models.UserRoleAdmin).Error; updateErr != nil {
			return nil, updateErr
		}
	}

	token, err := s.jwtUtil.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, User: user}, nil
}

func (s *AuthService) BindPhone(userID uint, phoneCode string) error {
	accessToken, err := s.wechatClient.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get wechat access token failed: %w", err)
	}

	phoneResp, err := s.wechatClient.GetPhoneNumber(accessToken, phoneCode)
	if err != nil {
		return fmt.Errorf("get wechat phone number failed: %w", err)
	}

	phone := strings.TrimSpace(phoneResp.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(phoneResp.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return fmt.Errorf("wechat phone number is empty")
	}

	updates := map[string]any{"phone": phone}
	if s.adminPhone != "" && normalizePhone(phone) == s.adminPhone {
		updates["role"] = models.UserRoleAdmin
	}
	if s.adminOpenID != "" {
		var current models.User
		if err := s.db.Select("open_id").Where("id = ?", userID).First(&current).Error; err == nil && current.OpenID == s.adminOpenID {
			updates["role"] = models.UserRoleAdmin
		}
	}
	if _, ok := updates["role"]; ok {
		updates["role"] = models.UserRoleAdmin
	}

	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (s *AuthService) GetMe(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
