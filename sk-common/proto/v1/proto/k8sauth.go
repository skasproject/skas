package proto

// -------------------------- Kubernetes Authentication webkook protocol

// Request is issued by Kubernetes API Server authentication webhook to validate a token
// Protocol is defined by Kubernetes

var TokenReviewMeta = RequestMeta{
	Method:  "POST",
	UrlPath: "/v1/tokenReview",
}

type TokenReviewRequest struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Token string `json:"token"`
	} `json:"spec"`
}

type TokenReviewUser struct {
	Username string   `json:"username"`
	Uid      string   `json:"uid"`
	Groups   []string `json:"groups"`
}

type TokenReviewResponse struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Authenticated bool             `json:"authenticated"`
		User          *TokenReviewUser `json:"user,omitempty"`
	} `json:"status"`
}
