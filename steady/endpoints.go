package steady

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"

	"github.com/maxmcd/steady/daemon/daemonrpc"
	db "github.com/maxmcd/steady/db"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/twitchtv/twirp"
)

var _ steadyrpc.Steady = new(Server)

func (s *Server) sendLoginEmail(ctx context.Context, user db.User) (err error) {
	token := steadyutil.RandomString(64)
	if token == "" {
		return twirp.InternalError("error generating random token")
	}
	resp, err := s.db.CreateLoginToken(ctx, db.CreateLoginTokenParams{
		UserID: user.ID,
		Token:  token,
	})
	if err != nil {
		return err
	}
	// Note: This endpoint sends urls that log into the webUI, not the API,
	// until we have API authentication that is not email based we rely on the
	// web-ui to serve these requests. There is no way to log into the api
	// without the web ui for the moment.

	// TODO: Send email to user.Email
	fmt.Println("LOGIN: /login/token/" + resp.Token)
	return nil
}

func (s *Server) getUserSession(ctx context.Context) (_ *db.UserSession, err error) {
	header, _ := twirp.HTTPRequestHeaders(ctx)
	token := header.Get("X-Steady-Token")
	if token == "" {
		return nil, twirp.NewError(twirp.Unauthenticated, "User session token not found in headers")
	}
	userSession, err := s.db.GetUserSession(ctx, token)
	if err == sql.ErrNoRows {
		return nil, twirp.NewError(twirp.Unauthenticated, "No active user session matches this token")
	}
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}
	return &userSession, nil
}

func (s *Server) Login(ctx context.Context, req *steadyrpc.LoginRequest) (_ *steadyrpc.LoginResponse, err error) {
	user, err := s.db.GetUserByEmailOrUsername(ctx, db.GetUserByEmailOrUsernameParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, twirp.NewError(twirp.NotFound, "not found")
		}
		return nil, err
	}

	if err := s.sendLoginEmail(ctx, user); err != nil {
		return nil, err
	}

	return &steadyrpc.LoginResponse{
		User: &steadyrpc.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *Server) Signup(ctx context.Context, req *steadyrpc.SignupRequest) (_ *steadyrpc.SignupResponse, err error) {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, err.Error())
	}

	user, err := s.db.GetUserByEmailOrUsername(ctx, db.GetUserByEmailOrUsernameParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		if user.Username == req.Username {
			return nil, twirp.NewError(twirp.AlreadyExists, "a user with this username already exists")
		}
		return nil, twirp.NewError(twirp.AlreadyExists, "a user with this email addr already exists")
	}

	user, err = s.db.CreateUser(ctx, db.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}

	if err := s.sendLoginEmail(ctx, user); err != nil {
		return nil, err
	}

	return &steadyrpc.SignupResponse{
		User: &steadyrpc.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *steadyrpc.ValidateTokenRequest) (
	_ *steadyrpc.ValidateTokenResponse, err error) {
	token, err := s.db.GetLoginToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	user, err := s.db.GetUser(ctx, token.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.db.DeleteLoginToken(ctx, token.Token); err != nil {
		return nil, err
	}
	userSession, err := s.db.CreateUserSession(ctx, db.CreateUserSessionParams{
		UserID: token.UserID,
	})
	if err != nil {
		return nil, err
	}
	return &steadyrpc.ValidateTokenResponse{
		User: &steadyrpc.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		UserSessionToken: userSession.Token,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *steadyrpc.LogoutRequest) (_ *steadyrpc.LogoutResponse, err error) {
	userSession, err := s.getUserSession(ctx)
	if err != nil {
		return nil, err
	}
	return &steadyrpc.LogoutResponse{}, s.db.DeleteUserSession(ctx, userSession.Token)
}

func (s *Server) CreateServiceVersion(ctx context.Context, req *steadyrpc.CreateServiceVersionRequest) (
	_ *steadyrpc.CreateServiceVersionResponse, err error) {
	if _, err := s.db.GetService(ctx, db.GetServiceParams{
		UserID: 1,
		ID:     req.ServiceId,
	}); err != nil {
		return nil, err
	}

	serviceVersion, err := s.db.CreateServiceVersion(ctx, db.CreateServiceVersionParams{
		ServiceID: req.ServiceId,
		Version:   req.Version,
		Source:    req.Source,
	})
	if err != nil {
		return nil, err
	}

	return &steadyrpc.CreateServiceVersionResponse{
		ServiceVersion: &steadyrpc.ServiceVersion{
			Id:        serviceVersion.ID,
			ServiceId: serviceVersion.ServiceID,
			Version:   serviceVersion.Version,
			Source:    serviceVersion.Source,
		},
	}, nil
}

func (s *Server) DeployApplication(ctx context.Context, req *steadyrpc.DeployApplicationRequeast) (
	_ *steadyrpc.DeployApplicationResponse, err error) {
	serviceVersion, err := s.db.GetServiceVersion(ctx, req.ServiceVersionId)
	if err != nil {
		return nil, err
	}
	app, err := s.daemonClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   req.Name,
		Script: serviceVersion.Source,
	})
	if err != nil {
		return nil, err
	}

	// TODO: deploy application to host, confirm that it works
	return &steadyrpc.DeployApplicationResponse{
		Application: &steadyrpc.Application{Name: app.Name},
		Url:         s.publicLoadBalancerURL,
	}, nil
}

func (s *Server) DeploySource(ctx context.Context, req *steadyrpc.DeploySourceRequest) (
	_ *steadyrpc.DeploySourceResponse, err error) {
	app, err := s.daemonClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   "faketemporaryname",
		Script: req.Source,
	})
	if err != nil {
		return nil, err
	}
	_ = app
	// TODO: deploy application to host, confirm that it works
	return &steadyrpc.DeploySourceResponse{
		Url: app.Name,
	}, nil
}