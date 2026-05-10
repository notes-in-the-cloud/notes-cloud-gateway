package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
)

const (
	publicAPIPrefix = "/api/v1"

	authPublicPrefix   = "/api/v1/auth"
	authInternalPrefix = "/authService/api/v1"
	authPublicMePath   = "/api/v1/me"
	authInternalMePath = "/authService/api/v1/me"

	notesPublicPrefix   = "/api/v1/notes"
	notesInternalPrefix = "/api/users/%s/notes"

	todoListsPublicPrefix   = "/api/v1/todo-lists"
	todoListsInternalPrefix = "/api/v1/users/%s/todo-lists"

	todosPublicPrefix   = "/api/v1/todos"
	todosInternalPrefix = "/api/v1/users/%s/todo-tasks"

	remindersPublicPrefix   = "/api/v1/reminders"
	remindersInternalPrefix = "/api/users/%s/reminders"

	notificationsPublicPrefix   = "/api/v1/notifications"
	notificationsInternalPrefix = "/api/users/%s/notifications"

	shareLinksPublicPrefix   = "/api/v1/share-links"
	shareLinksInternalPrefix = "/api/v1/share-links"

	noteShareLinksSuffix = "/share-links"
)

type ServiceProxy struct {
	authProxy     *httputil.ReverseProxy
	notesProxy    *httputil.ReverseProxy
	todoProxy     *httputil.ReverseProxy
	sharingProxy  *httputil.ReverseProxy
	reminderProxy *httputil.ReverseProxy
}

type ServiceURLs struct {
	AuthURL     string
	NotesURL    string
	TodoURL     string
	SharingURL  string
	ReminderURL string
}

func NewServiceProxy(urls ServiceURLs) (*ServiceProxy, error) {
	authProxy, err := newProxy(urls.AuthURL)
	if err != nil {
		return nil, err
	}

	notesProxy, err := newProxy(urls.NotesURL)
	if err != nil {
		return nil, err
	}

	todoProxy, err := newProxy(urls.TodoURL)
	if err != nil {
		return nil, err
	}

	sharingProxy, err := newProxy(urls.SharingURL)
	if err != nil {
		return nil, err
	}

	reminderProxy, err := newProxy(urls.ReminderURL)
	if err != nil {
		return nil, err
	}

	return &ServiceProxy{
		authProxy:     authProxy,
		notesProxy:    notesProxy,
		todoProxy:     todoProxy,
		sharingProxy:  sharingProxy,
		reminderProxy: reminderProxy,
	}, nil
}

func newProxy(targetURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		writeProxyErrorResponse(
			w,
			http.StatusBadGateway,
			"SERVICE_UNAVAILABLE",
			"target service is currently unavailable",
		)
	}

	return proxy, nil
}

func (p *ServiceProxy) Auth(w http.ResponseWriter, r *http.Request) {
	rewrittenPath := rewriteAuthPath(r.URL.Path)
	p.authProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))
}

func (p *ServiceProxy) Notes(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}

	rewrittenPath := rewriteUserScopedPath(
		r.URL.Path,
		notesPublicPrefix,
		notesInternalPrefix,
		userID,
	)

	p.notesProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))
}

func (p *ServiceProxy) Todo(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}

	rewrittenPath := rewriteTodoPath(r.URL.Path, userID)

	p.todoProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))
}

func (p *ServiceProxy) Sharing(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, shareLinksPublicPrefix) {
		rewrittenPath := rewritePublicShareLinkPath(r.URL.Path)
		p.sharingProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))

		return
	}

	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}

	rewrittenPath := rewriteNoteShareLinkPath(r.URL.Path, userID)

	p.sharingProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))
}

func (p *ServiceProxy) Reminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}

	rewrittenPath := rewriteReminderPath(r.URL.Path, userID)

	p.reminderProxy.ServeHTTP(w, cloneRequestWithPath(r, rewrittenPath))
}

func userIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, err := accesstoken.UserIDFromContext(r.Context())
	if err != nil {
		accesstoken.WriteErrorResponse(
			w,
			http.StatusUnauthorized,
			accesstoken.ErrInvalidToken,
			"user is not authenticated",
		)

		return "", false
	}

	return userID, true
}

func cloneRequestWithPath(r *http.Request, newPath string) *http.Request {
	clonedRequest := r.Clone(r.Context())

	clonedURL := *r.URL
	clonedURL.Path = newPath
	clonedURL.RawPath = ""

	clonedRequest.URL = &clonedURL

	return clonedRequest
}

func rewriteAuthPath(originalPath string) string {
	if originalPath == authPublicMePath {
		return authInternalMePath
	}

	if strings.HasPrefix(originalPath, "/api/v1/auth/google") {
		return strings.Replace(
			originalPath,
			"/api/v1/auth/google",
			"/authService/api/v1/auth/google",
			1,
		)
	}

	if strings.HasPrefix(originalPath, "/api/v1/auth/gitlab") {
		return strings.Replace(
			originalPath,
			"/api/v1/auth/gitlab",
			"/authService/api/v1/auth/gitlab",
			1,
		)
	}

	if strings.HasPrefix(originalPath, "/api/v1/auth/email") {
		return strings.Replace(
			originalPath,
			"/api/v1/auth/email",
			"/authService/api/v1/email",
			1,
		)
	}

	if strings.HasPrefix(originalPath, authPublicPrefix) {
		return strings.Replace(originalPath, authPublicPrefix, authInternalPrefix, 1)
	}

	return originalPath
}

func rewriteTodoPath(originalPath string, userID string) string {
	if strings.HasPrefix(originalPath, todoListsPublicPrefix) {
		return rewriteUserScopedPath(
			originalPath,
			todoListsPublicPrefix,
			todoListsInternalPrefix,
			userID,
		)
	}

	if strings.HasPrefix(originalPath, todosPublicPrefix) {
		return rewriteUserScopedPath(
			originalPath,
			todosPublicPrefix,
			todosInternalPrefix,
			userID,
		)
	}

	return originalPath
}

func rewriteReminderPath(originalPath string, userID string) string {
	if strings.HasPrefix(originalPath, remindersPublicPrefix) {
		return rewriteUserScopedPath(
			originalPath,
			remindersPublicPrefix,
			remindersInternalPrefix,
			userID,
		)
	}

	if strings.HasPrefix(originalPath, notificationsPublicPrefix) {
		return rewriteUserScopedPath(
			originalPath,
			notificationsPublicPrefix,
			notificationsInternalPrefix,
			userID,
		)
	}

	return originalPath
}

func rewriteNoteShareLinkPath(originalPath string, userID string) string {
	escapedUserID := url.PathEscape(userID)

	if strings.HasPrefix(originalPath, notesPublicPrefix) &&
		strings.HasSuffix(originalPath, noteShareLinksSuffix) {
		return strings.Replace(
			originalPath,
			notesPublicPrefix,
			"/api/v1/users/"+escapedUserID+"/notes",
			1,
		)
	}

	return originalPath
}

func rewritePublicShareLinkPath(originalPath string) string {
	if strings.HasPrefix(originalPath, shareLinksPublicPrefix) {
		return strings.Replace(
			originalPath,
			shareLinksPublicPrefix,
			shareLinksInternalPrefix,
			1,
		)
	}

	return originalPath
}

func rewriteUserScopedPath(
	originalPath string,
	publicPrefix string,
	internalPrefixFormat string,
	userID string,
) string {
	escapedUserID := url.PathEscape(userID)
	internalPrefix := strings.Replace(internalPrefixFormat, "%s", escapedUserID, 1)

	return strings.Replace(originalPath, publicPrefix, internalPrefix, 1)
}

type proxyErrorResponse struct {
	Error proxyErrorDetails `json:"error"`
}

type proxyErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeProxyErrorResponse(
	w http.ResponseWriter,
	statusCode int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := proxyErrorResponse{
		Error: proxyErrorDetails{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}