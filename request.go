package main

import "github.com/21strive/commonuser"

type NativeRegister struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	commonuser.DeviceInfo
}
type VerifyRegistration struct {
	VerificationCode string `json:"verificationCode"`
}

type LoginWithEmail struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	commonuser.DeviceInfo
}

type LoginWithUsername struct {
	Username string `json:"username"`
	Password string `json:"password"`
	commonuser.DeviceInfo
}

type AuthWithGoogle struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	commonuser.DeviceInfo
}

type UpdateAccount struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type UpdateEmail struct {
	NewEmail string `json:"newEmail"`
}

type ValidateUpdateEmail struct {
	AccountUUID string `json:"accountUUID"`
	Token       string `json:"token"`
}

type RevokeUpdateEmail struct {
	AccountUUID string `json:"accountUUID"`
	RevokeToken string `json:"revokeToken"`
}

type UpdatePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	AccountUUID string `json:"accountUUID"`
	Token       string `json:"token"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=255"`
}
