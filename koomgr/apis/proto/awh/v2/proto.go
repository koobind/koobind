package v2

// -------------------------- Kubernetes Authentication webkook protocol

var TokenReviewUrlPath = "/awh/v2/tokenReview"

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
