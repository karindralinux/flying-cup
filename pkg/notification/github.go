package notification

import (
	"context"

	"github.com/google/go-github/v55/github"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

type GithubNotifier struct {
	client *github.Client
}

func NewGithubNotifier(token string) *GithubNotifier {
	return &GithubNotifier{
		client: github.NewClient(nil).WithAuthToken(token),
	}
}

func (g *GithubNotifier) CreateCommentPR(ctx context.Context, webhook *webhook.GithubPRWebhook, comment string) error {
	
	prComment := &github.IssueComment{
		Body: &comment,
	}

	_, _, err := g.client.Issues.CreateComment(ctx, webhook.Sender.Username, webhook.Repository.Name, webhook.Number, prComment)

	if err != nil {
		return err
	}

	return nil
}
