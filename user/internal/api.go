package api

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
	ifollowers "yatc/user/internal/followers"
	"yatc/user/pkg/followers"
	"yatc/user/pkg/users"
)

type UserApi struct {
	userService     users.Service
	followerService followers.Service
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

func NewUserApi(userService users.Service, followerService followers.Service) *UserApi {
	return &UserApi{userService, followerService}
}

func (api *UserApi) ConfigureRouter(router chi.Router) {
	handler := HandlerWithOptions(api,
		ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		}})

	router.Mount("/", handler)
}

func (api *UserApi) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest CreateUserRequest
	err := render.Decode(r, &createUserRequest)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		return
	}

	user := UserFromCreateUserRequest(createUserRequest)
	user, err = api.userService.CreateUser(user)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	internal.ReplyWithStatusWithJSON(w, r, http.StatusCreated, user)
}

func (api *UserApi) DeleteUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	user, err := api.userService.DeleteUser(userId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) GetUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	user, err := api.userService.GetUser(userId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, UserResponseFromUser(user))
}

func (api *UserApi) GetFollowees(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	allUsers, err := api.followerService.GetFollowees(userId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(userId)) {
			internal.ReplyWithError(w, r, err, http.StatusNotFound)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	userResponses := make([]UserResponse, len(allUsers))
	for i, user := range allUsers {
		userResponses[i] = UserResponseFromUser(user)
	}

	internal.ReplyWithStatusOkWithJSON(w, r, userResponses)
}

func (api *UserApi) GetFollowers(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	allUsers, err := api.followerService.GetFollowers(userId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(userId)) {
			internal.ReplyWithError(w, r, err, http.StatusNotFound)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	userResponses := make([]UserResponse, len(allUsers))
	for i, user := range allUsers {
		userResponses[i] = UserResponseFromUser(user)
	}

	internal.ReplyWithStatusOkWithJSON(w, r, userResponses)
}

func (api *UserApi) FollowUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	var createFollowerRequest CreateFollowerRequest
	err := render.Decode(r, &createFollowerRequest)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	user, err := api.followerService.FollowUser(userId, createFollowerRequest.Id)
	if err != nil {
		if errors.Is(err, ifollowers.SelfFollowError) {
			internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	internal.ReplyWithStatusWithJSON(w, r, http.StatusOK, UserResponseFromUser(user))
}

func (api *UserApi) UnfollowUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID, followerUserId uuid.UUID) {
	err := api.followerService.UnfollowUser(userId, followerUserId)
	if err != nil {
		if errors.Is(err, ifollowers.SelfFollowError) {
			internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	internal.ReplyWithStatusOk(r)
}
