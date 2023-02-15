package service

import (
	"net/http"

	"github.com/vibeitco/go-utils/server"
)

type handler struct {
	config Config
}
type AuthenticationResponse struct {
	UserID string `json:"userId"`
	Token  string `json:"token"`
}

func NewHandler(conf Config) (*handler, error) {
	h := &handler{
		config: conf,
	}

	return h, nil
}

func (h *handler) Auth(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	payload := &AuthenticationResponse{
		UserID: "Test",
		Token:  "TestToken",
	}

	server.JSON(ctx, res, http.StatusOK, payload)
}
