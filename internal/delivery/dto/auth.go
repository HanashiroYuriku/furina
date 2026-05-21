package dto

type UserRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=20,unique=users->username,whitespace,username" example:"hanashiroyuriku"`
	Email       string `json:"email" validate:"required,max=100,email,unique=users->email,whitespace" example:"hanashiroyuriku@gmail.com"`
	DisplayName string `json:"displayName" example:"Hanashiro Yuriku"`
	Password    string `json:"password" validate:"required,complexpassword" example:"P4$$word"`
}

type UserVerificationRequest struct {
	Email string `json:"email" validate:"required,incolumn=user_verifications->email"`
}

type LoginResponse struct {
	TokenResponse
	User UserResponse `json:"user"`
}

type LoginRequest struct {
	EmailUsername string `json:"emailUsername" validate:"required" example:"hanashiroyuriku"`
	Password      string `json:"password" validate:"required" example:"P4$$w0rd"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type TokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}