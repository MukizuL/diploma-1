package dto

type AuthForm struct {
	Login    string `json:"login" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}
