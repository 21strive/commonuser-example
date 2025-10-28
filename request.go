package main

type NativeRegister struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
}
type VerifyRegistration struct {
	AccountUUID      string `json:"accountUUID"`
	VerificationCode string `json:"verificationCode"`
}

type LoginWithEmail struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginWithUsername struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateAccount struct {
	AccountUUID string `json:"accountUUID"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
}

type UpdateEmail struct {
	AccountUUID string `json:"accountUUID"`
	NewEmail    string `json:"newEmail"`
}

type ValidateUpdateEmail struct {
	AccountUUID string `json:"accountUUID"`
	Token       string `json:"token"`
}

type RevokeUpdateEmail struct {
	AccountUUID string `json:"accountUUID"`
}

type UpdatePassword struct {
	AccountUUID string `json:"accountUUID"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	Token       string `json:"token"`
	NewPassword string `json:"<PASSWORD>"`
}
