package prchecker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/github"
)

// pullRequestChecker checks pull request event and manage pull request.
type pullRequestChecker struct {
	config Config
	client *github.Client
	event  *github.PullRequestEvent

	eventAction  string
	repoOwner    string
	repoName     string
	issueCreator string
	issueNumber  int
	commitBefore string
	commitAfter  string
	changedFiles int
}

// newPullRequestChecker creates initialized pullRequestChecker.
func newPullRequestChecker(conf Config, client *github.Client, e *github.PullRequestEvent, payload []byte) *pullRequestChecker {
	pushData := github.PushEvent{}
	json.Unmarshal(payload, &pushData)

	repo := e.PullRequest.Base.Repo
	return &pullRequestChecker{
		config:       conf,
		client:       client,
		event:        e,
		eventAction:  e.GetAction(),
		repoOwner:    repo.Owner.GetLogin(),
		repoName:     repo.GetName(),
		issueCreator: e.PullRequest.User.GetLogin(),
		issueNumber:  e.GetNumber(),
		commitBefore: pushData.GetBefore(),
		commitAfter:  pushData.GetAfter(),
		changedFiles: e.PullRequest.GetChangedFiles(),
	}
}

// Do checks pull request event and append comments/assignees/reviwers.
func (c *pullRequestChecker) Do() error {
	switch {
	case c.issueNumber < 1,
		c.changedFiles < 1,
		!c.IsValidAction():
		return nil
	}

	conf := c.config
	base := c.event.PullRequest.Base
	head := c.event.PullRequest.Head
	repoConfig := conf.GetRepository(base.Repo.GetFullName())
	if repoConfig == nil {
		return nil
	}

	c.loggingInfo("[commitHash:%s]", c.commitAfter)

	switch {
	case !c.config.AddComment,
		!c.HasCompareCommitHash():
		c.commitBefore = base.GetSHA()
		c.commitAfter = head.GetSHA()
	}

	changedFiles, err := c.apiCompareCommit()
	switch {
	case err != nil:
		return err
	case len(changedFiles) == 0:
		return nil
	}

	comment := Comment{}
	assignees := &Assignees{}
	reviewers := &Assignees{}
	for i := range repoConfig.Files {
		rf := repoConfig.Files[i]
		for _, f := range changedFiles {
			if rf.Match(f) {
				comment.Add(rf.GetComment())
				assignees.Add(rf.Assignees...)
				reviewers.Add(rf.Reviewers...)
				break
			}
		}
	}

	c.apiCreateOrEditComment(comment)

	// ignore already removed assignee and reviewers.
	removedUsers, err := c.apiGetRemovedEvent()
	if err == nil {
		removed := removedUsers.List()
		fmt.Printf("[removed] %+v\n", removed)
		assignees.RemoveFromList(removed...)
		reviewers.RemoveFromList(removed...)
	}
	c.apiAddAssignee(assignees)
	c.apiAddReviewers(reviewers)
	return nil
}

// HasCompareCommitHash checks if pull request event has commet hash of "before" and "after".
func (c *pullRequestChecker) HasCompareCommitHash() bool {
	return c.commitBefore != "" && c.commitAfter != ""
}

// IsValidAction checks if the pull request event should be process or not.
func (c *pullRequestChecker) IsValidAction() bool {
	switch c.eventAction {
	case "opened",
		"synchronize":
		return true
	}
	return false
}

// apiCompareCommit compares past commit and current comment and return changed files list.
func (c *pullRequestChecker) apiCompareCommit() (changedFiles []string, err error) {
	ctx := context.Background()
	comp, _, err := c.client.Repositories.CompareCommits(ctx, c.repoOwner, c.repoName, c.commitBefore, c.commitAfter)
	switch {
	case err != nil:
		c.loggingError("`CompareCommits` operation: %s", err)
		return nil, err
	case len(comp.Files) == 0:
		return nil, nil
	}

	changedFiles = make([]string, len(comp.Files))
	for i, f := range comp.Files {
		changedFiles[i] = f.GetFilename()
	}
	return changedFiles, nil
}

// apiCreateOrEditComment creates or edits comment to a pull request.
func (c *pullRequestChecker) apiCreateOrEditComment(comment Comment) error {
	if !comment.HasComment() {
		return nil
	}

	newCommentBody := comment.Show()
	if c.config.AddComment {
		return c.apiCreateComment(newCommentBody)
	}

	// create or edit comment.
	pastComment, err := c.apiGetPastComment()
	if err != nil {
		return err
	}

	var pastCommentID int64
	if pastComment != nil {
		pastCommentID = pastComment.GetID()
	}

	if pastComment != nil && pastComment.GetBody() == newCommentBody {
		c.loggingInfo("[commentID:%d] same comment is found, skip update.", pastCommentID)
		return nil
	}

	switch {
	case pastCommentID > 0:
		return c.apiEditComment(newCommentBody, pastCommentID)
	default:
		return c.apiCreateComment(newCommentBody)
	}
}

// apiCreateComment creates comment to a pull request.
func (c *pullRequestChecker) apiCreateComment(comment string) error {
	issueComment := github.IssueComment{
		Body: &comment,
	}

	ctx := context.Background()
	_, _, err := c.client.Issues.CreateComment(ctx, c.repoOwner, c.repoName, c.issueNumber, &issueComment)
	c.loggingInfo("Executed `CreateComment` operation.")
	c.loggingError("`CreateComment` operation: %s", err)
	return err
}

// apiEditComment edits the comment to a pull request.
func (c *pullRequestChecker) apiEditComment(comment string, commentID int64) error {
	issueComment := github.IssueComment{
		Body: &comment,
	}

	ctx := context.Background()
	_, _, err := c.client.Issues.EditComment(ctx, c.repoOwner, c.repoName, commentID, &issueComment)
	c.loggingInfo("[commentID:%d] Executed `EditComment` operation.", commentID)
	c.loggingError("`EditComment` operation: %s", err)
	return err
}

// apiGetPastComment gets the past comment from the pull request.
func (c *pullRequestChecker) apiGetPastComment() (*github.IssueComment, error) {
	ctx := context.Background()
	list, _, err := c.client.Issues.ListComments(ctx, c.repoOwner, c.repoName, c.issueNumber, nil)
	switch {
	case err != nil:
		c.loggingError("`ListComments` operation: %s", err)
		return nil, err
	case len(list) == 0:
		return nil, nil
	}

	// check bot's past comment
	botID := c.config.BotID
	for _, comment := range list {
		if comment.User.GetID() == botID {
			return comment, nil
		}
	}
	return nil, nil
}

// apiGetRemovedEvent gets the removing of assignees and reviwers events from the pull request.
func (c *pullRequestChecker) apiGetRemovedEvent() (*Assignees, error) {
	ctx := context.Background()
	unassigned := Assignees{}

	list, _, err := c.client.Issues.ListIssueEvents(ctx, c.repoOwner, c.repoName, c.issueNumber, nil)
	if err != nil {
		c.loggingError("`ListIssueEvents` operation: %s", err)
		return nil, err
	}

	// check removed assignee and reviewr event.
	for _, ev := range list {
		var username string
		switch ev.GetEvent() {
		case "unassigned":
			username = ev.GetAssignee().GetName()
		case "review_request_removed":
			// username = events[i].GetRequestedReviewer().GetName()
			continue
		default:
			continue
		}
		unassigned.Add(username)
	}
	return &unassigned, nil
}

// apiAddAssignee adds assignees to a pull request.
func (c *pullRequestChecker) apiAddAssignee(assignees *Assignees) error {
	if !assignees.HasAssignees() {
		return nil
	}

	ctx := context.Background()
	_, _, err := c.client.Issues.AddAssignees(ctx, c.repoOwner, c.repoName, c.issueNumber, assignees.List())
	c.loggingInfo("Executed `AddAssignees` operation: %v", assignees.List())
	c.loggingError("`AddAssignees` operation: %s", err)
	return err
}

// apiAddReviewers adds reviwers to a pull request.
func (c *pullRequestChecker) apiAddReviewers(reviewers *Assignees) error {
	reviewers.RemoveFromList(c.issueCreator)
	if !reviewers.HasAssignees() {
		return nil
	}

	ctx := context.Background()
	_, _, err := c.client.PullRequests.RequestReviewers(ctx, c.repoOwner, c.repoName, c.issueNumber, github.ReviewersRequest{
		Reviewers: reviewers.List(),
	})
	c.loggingInfo("Executed `RequestReviewers` operation: %v", reviewers.List())
	c.loggingError("`RequestReviewers` operation: %s", err)
	return err
}

func (c *pullRequestChecker) loggingError(template string, err error) {
	if err == nil {
		return
	}

	fmt.Printf("[Checker] [ERROR] [repo:%s/%s] [issue:%d] %s\n", c.repoOwner, c.repoName, c.issueNumber, fmt.Sprintf(template, err.Error()))
}

func (c *pullRequestChecker) loggingInfo(template string, params ...interface{}) {
	fmt.Printf("[Checker] [INFO] [repo:%s/%s] [issue:%d] %s\n", c.repoOwner, c.repoName, c.issueNumber, fmt.Sprintf(template, params...))
}
