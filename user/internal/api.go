package api

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
	"yatc/internal"
	ifollowers "yatc/user/internal/followers"
	"yatc/user/pkg/followers"
	"yatc/user/pkg/users"
)

type UserApi struct {
	userService     users.UserService
	followerService followers.FollowerService
}

type ErrorResponse struct {
	Method    string
	Path      string
	Timestamp time.Time
	Message   string
}

func Error(err error, user int, w http.ResponseWriter, r *http.Request) {
	log.Println(err.Error())
	render.Status(r, user)
	errorRes := ErrorResponse{
		Method:    r.Method,
		Path:      r.RequestURI,
		Timestamp: time.Now().UTC(),
		Message:   "Error",
	}
	render.JSON(w, r, errorRes)
}

func UserResponseFromUser(user users.User) UserResponse {
	return UserResponse{
		Username: user.Name,
		Id:       user.Id,
	}
}

func UserFromCreateUserRequest(request CreateUserRequest) users.User {
	return users.User{
		Id:   uuid.New(),
		Name: request.Username,
	}
}

func NewUserApi(userService users.UserService, followerService followers.FollowerService) *UserApi {
	return &UserApi{userService, followerService}
}

func (api *UserApi) ConfigureRouter(router chi.Router) {
	handler := HandlerWithOptions(api,
		ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			Error(err, http.StatusBadRequest, w, r)
		}})

	router.Mount("/", handler)
}

func (api *UserApi) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest CreateUserRequest
	err := render.Decode(r, &createUserRequest)
	if err != nil {
		Error(err, http.StatusBadRequest, w, r)
		return
	}

	user := UserFromCreateUserRequest(createUserRequest)
	user, err = api.userService.CreateUser(user)
	if err != nil {
		Error(err, http.StatusInternalServerError, w, r)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) DeleteUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	user, err := api.userService.DeleteUser(userId)
	if err != nil {
		Error(err, http.StatusNotFound, w, r)
		return
	}

	render.JSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) GetUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	user, err := api.userService.GetUser(userId)
	if err != nil {
		Error(err, http.StatusNotFound, w, r)
		return
	}

	render.JSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) GetFollowees(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	allUsers, err := api.followerService.GetFollowees(userId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(userId)) {
			Error(err, http.StatusNotFound, w, r)
		} else {
			Error(err, http.StatusInternalServerError, w, r)
		}
		return
	}

	userResponses := make([]UserResponse, len(allUsers))
	for i, user := range allUsers {
		userResponses[i] = UserResponseFromUser(user)
	}

	render.JSON(w, r, userResponses)
}

func (api *UserApi) GetFollowers(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	allUsers, err := api.followerService.GetFollowers(userId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(userId)) {
			Error(err, http.StatusNotFound, w, r)
		} else {
			Error(err, http.StatusInternalServerError, w, r)
		}
		return
	}

	userResponses := make([]UserResponse, len(allUsers))
	for i, user := range allUsers {
		userResponses[i] = UserResponseFromUser(user)
	}

	render.JSON(w, r, userResponses)
}

func (api *UserApi) FollowUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	var createFollowerRequest CreateFollowerRequest
	err := render.Decode(r, &createFollowerRequest)
	if err != nil {
		Error(err, http.StatusBadRequest, w, r)
		return
	}

	user, err := api.followerService.FollowUser(userId, createFollowerRequest.Id)
	if err != nil {
		if errors.Is(err, ifollowers.SelfFollowError) {
			Error(err, http.StatusBadRequest, w, r)
		} else {
			Error(err, http.StatusNotFound, w, r)
		}
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) UnfollowUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID, followerUserId uuid.UUID) {
	err := api.followerService.UnfollowUser(userId, followerUserId)
	if err != nil {
		if errors.Is(err, ifollowers.SelfFollowError) {
			Error(err, http.StatusBadRequest, w, r)
		} else {
			Error(err, http.StatusNotFound, w, r)
		}
		return
	}

	render.Status(r, http.StatusOK)
}
