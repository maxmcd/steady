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

	link := "/login/token/" + resp.Token
	// TODO: Send email to user.Email
	if s.emailSink != nil {
		s.emailSink(link)
	}
	fmt.Println("LOGIN: " + link)
	return nil
}

func (s *Server) getUserSession(ctx context.Context) (_ *db.UserSession, err error) {
	header, found := twirp.HTTPRequestHeaders(ctx)
	if !found {
		return nil, twirp.NewError(twirp.Unauthenticated, "User session token not found in headers")
	}
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
	if req.Email == "" && req.Username == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "username or email cannot be blank")
	}
	user, err := s.db.GetUserByEmailOrUsername(ctx, db.GetUserByEmailOrUsernameParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, twirp.NewError(twirp.NotFound, "A user with this username or email could not be found")
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
	if req.Email == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "email address cannot be blank")
	}
	if req.Username == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "username cannot be blank")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, "email address is invalid")
	}

	user, err := s.db.GetUserByEmailOrUsername(ctx, db.GetUserByEmailOrUsernameParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && user.Username == req.Username {
		return nil, twirp.NewError(twirp.AlreadyExists, "a user with this username already exists")
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

func (s *Server) GetUser(ctx context.Context, req *steadyrpc.GetUserRequest) (_ *steadyrpc.GetUserResponse, err error) {
	userSession, err := s.getUserSession(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.db.GetUser(ctx, userSession.UserID)
	if err != nil {
		return nil, err
	}
	return &steadyrpc.GetUserResponse{
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
	sessionToken := steadyutil.RandomString(64)
	if sessionToken == "" {
		return nil, twirp.InternalError("error generating random token")
	}

	userSession, err := s.db.CreateUserSession(ctx, db.CreateUserSessionParams{
		UserID: token.UserID,
		Token:  sessionToken,
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

func (s *Server) RunApplication(ctx context.Context, req *steadyrpc.RunApplicationRequest) (
	_ *steadyrpc.RunApplicationResponse, err error,
) {
	if req.Source == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "Application source cannot be empty")
	}

	if req.Name == "" {
		req.Name = steadyutil.RandomString(8)
	}
	app, err := s.db.CreateApplication(ctx, db.CreateApplicationParams{
		Name:   req.Name,
		UserID: sql.NullInt64{},
		Source: req.Source,
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.daemonClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   app.Name,
		Script: req.Source,
	}); err != nil {
		return nil, err
	}

	// TODO: deploy application to host, confirm that it works
	return &steadyrpc.RunApplicationResponse{
		Application: &steadyrpc.Application{Name: app.Name},
		Url:         s.appURL(app.Name),
	}, nil
}

func (s *Server) appURL(name string) string {
	copy := *s.parsedPublicLB
	copy.Host = name + "." + copy.Host
	return copy.String()
}

func (s *Server) GetApplication(ctx context.Context, req *steadyrpc.GetApplicationRequest) (
	_ *steadyrpc.GetApplicationResponse, err error) {
	resp, err := s.db.GetApplication(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &steadyrpc.GetApplicationResponse{
		Application: &steadyrpc.Application{
			Name:   resp.Name,
			UserId: resp.UserID.Int64,
			Id:     resp.ID,
		},
		Url: s.appURL(resp.Name),
	}, nil
}

func (s *Server) UpdateApplication(ctx context.Context, req *steadyrpc.UpdateApplicationRequest) (
	_ *steadyrpc.UpdateApplicationResponse, err error,
) {
	app, err := s.db.GetApplication(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if !app.UserID.Valid {
		userSession, err := s.getUserSession(ctx)
		if err != nil {
			return nil, err
		}
		if app.UserID.Int64 != userSession.UserID {
			return nil, twirp.NewError(twirp.PermissionDenied, "this application is owned by a different user")
		}
	}
	if _, err := s.daemonClient.UpdateApplication(ctx, &daemonrpc.UpdateApplicationRequest{
		Name:   app.Name,
		Script: req.Source,
	}); err != nil {
		return nil, err
	}
	if app, err = s.db.UpdateApplication(ctx, db.UpdateApplicationParams{
		Source: req.Source,
		ID:     app.ID,
	}); err != nil {
		return nil, err
	}
	return &steadyrpc.UpdateApplicationResponse{Application: &steadyrpc.Application{
		Name:   app.Name,
		UserId: app.UserID.Int64,
		Id:     app.ID,
	}}, nil
}
