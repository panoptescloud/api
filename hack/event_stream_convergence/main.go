package main

import (
	"fmt"
	"os"
	"slices"
	"time"
)

type EventType string

const (
	GithubEventTypePullRequestOpened EventType = "github.pull_request_opened"
	GithubEventTypePullRequestMerged EventType = "github.pull_request_merged"
	GithubEventTypePullRequestIssueLinked EventType = "github.issue_linked_to_pr"
	GithubEventTypePullRequestIssueUnLinked EventType = "github.issue_unlinked_to_pr"
)

const (
	JiraEventTypeIssueCreated EventType = "jira.issue_created"
	JiraEventTypeIssueStatusChanged EventType = "jira.issue_status_changed"
	JiraEventTypeIssueClosed EventType = "jira.issue_completed"
)

type GithubPullRequest struct {
	Status string
	Issues []string
}

func (agg *GithubPullRequest) addIssueRef(id string) {
	if slices.Contains(agg.Issues, id) {
		return
	}

	agg.Issues = append(agg.Issues, id)
}

func (agg *GithubPullRequest) removeIssueRef(id string) {
	idx := slices.Index(agg.Issues, id)
	if idx == -1 {
		return
	}

	// Blah, make this cleaner
	first := agg.Issues[0:idx]
	second := agg.Issues[idx:]
	joined := append(first, second...)

	agg.Issues = joined
}

func (agg *GithubPullRequest) Project(s EventStream)  {
	for _, e := range s {
		agg.Apply(e)
	}
}

func (agg *GithubPullRequest) Apply(e AggregateEvent[EventPayload])  {
	switch e.Type() {
	case GithubEventTypePullRequestOpened:
		agg.Status = "open"
	case GithubEventTypePullRequestMerged:
		agg.Status = "merged"
	case GithubEventTypePullRequestIssueLinked:
		pl, ok := e.Payload().(GithubPullRequestIssueLinkedPayload)
		if ! ok {
			panic("incorrect payload type for GithubEventTypePullRequestIssueLinked")
		}

		agg.addIssueRef(pl.IssueId)
	case GithubEventTypePullRequestIssueUnLinked:
		pl, ok := e.Payload().(GithubPullRequestIssueLinkedPayload)
		if ! ok {
			panic("incorrect payload type for GithubEventTypePullRequestIssueLinked")
		}

		agg.removeIssueRef(pl.IssueId)
	default:
	    panic("unsupported event type in github event apply")
	}
}

// Slightly awkward way to make generics work, we just need to implement this method
type EventPayload interface {
	_eventPayload()
}

type AggregateEvent[T EventPayload] struct {
	aggregateId string
	eventType EventType
	occurredAt time.Time
	payload T
}

func (e AggregateEvent[T]) AggregateId() string {
	return e.aggregateId
}

func (e AggregateEvent[T]) Type() EventType {
	return e.eventType
}

func (e AggregateEvent[T]) OccurredAt() time.Time {
	return e.occurredAt
}

func (e AggregateEvent[T]) Payload() T {
	return e.payload
}



var startTime = time.Date(2025, 01, 01, 0, 0 , 0, 0, time.UTC)

type GithubPullRequestIssueLinkedPayload struct {
	IssueId string
}

func (p GithubPullRequestIssueLinkedPayload) _eventPayload() {}

type DefaultPayload struct {
}

func (p DefaultPayload) _eventPayload() {}

type GithubPullRequestOpenedEvent struct {
	AggregateEvent[DefaultPayload]
}

type EventStream []AggregateEvent[EventPayload]


func (s EventStream) AggregateIds() []string {
	ids := []string{}
	for _, e := range s {
		ids = append(ids, e.AggregateId())
	}

	return ids
}

func (s EventStream) FindByType(t EventType) EventStream {
	filtered := EventStream{}
	for _, e := range s {
		if e.Type() == t {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

func (s EventStream) FindByAggregateId(id string) EventStream {
	filtered := EventStream{}
	for _, e := range s {
		if e.AggregateId() == id {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

func (s EventStream) FindByAggregateIds(ids... string) EventStream {
	filtered := EventStream{}
	for _, e := range s {
		matches := slices.ContainsFunc(ids, func(id string) bool {
			return e.AggregateId() == id
		})
		if matches {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

func getGithubEventStream() EventStream {
	return EventStream{
		{
			aggregateId: "1234",
			eventType: GithubEventTypePullRequestOpened,
			occurredAt: startTime.Add(12 * time.Hour),
			payload: DefaultPayload{},
		},
		{
			aggregateId: "1234",
			eventType: GithubEventTypePullRequestIssueLinked,
			occurredAt: startTime.Add(24 * time.Hour),
			payload: GithubPullRequestIssueLinkedPayload{
				IssueId: "TEST-123",
			},
		},
		{
			aggregateId: "1234",
			eventType: GithubEventTypePullRequestMerged,
			occurredAt: startTime.Add(36 * time.Hour),
			payload: DefaultPayload{},
		},
	}
}

func getJiraEventStream() EventStream {
	return EventStream{
		{
			aggregateId: "TEST-123",
			eventType: JiraEventTypeIssueCreated,
			occurredAt: startTime.Add(18 * time.Hour),
			payload: DefaultPayload{},
		},
		{
			aggregateId: "TEST-123",
			eventType: JiraEventTypeIssueStatusChanged,
			occurredAt: startTime.Add(19 * time.Hour),
			payload: DefaultPayload{},
		},
		{
			aggregateId: "TEST-123",
			eventType: JiraEventTypeIssueStatusChanged,
			occurredAt: startTime.Add(20 * time.Hour),
			payload: DefaultPayload{},
		},
		{
			aggregateId: "TEST-123",
			eventType: JiraEventTypeIssueStatusChanged,
			occurredAt: startTime.Add(30 * time.Hour),
			payload: DefaultPayload{},
		},
		{
			aggregateId: "TEST-123",
			eventType: JiraEventTypeIssueClosed,
			occurredAt: startTime.Add(48 * time.Hour),
			payload: DefaultPayload{},
		},
	}
}

func extractJiraReferences(s EventStream) []string {
	refs := []string{}

	for _, e := range s {
		unknownPayload := e.Payload()
		issuePayload, ok := unknownPayload.(GithubPullRequestIssueLinkedPayload)

		if ! ok {
			continue
		}

		refs = append(refs, issuePayload.IssueId)
	}

	return refs
}

func buildPullRequest(s EventStream) *GithubPullRequest {
	pr := &GithubPullRequest{}

	pr.Project(s)

	return pr
}

func main() {
	args := os.Args

	if len(args) != 2 {
		panic("single arg required, pull request id")
	}

	ghStream := getGithubEventStream()
	jiraStream := getJiraEventStream()

	forPR := ghStream.FindByAggregateId(args[1])
	pr := buildPullRequest(forPR)

	fmt.Printf("PR:\n\n\t%#v\n\n", pr)

	relatedJiraEvents := jiraStream.FindByAggregateIds(pr.Issues...)


	// ghEvents := githubEventsById(args[1])

	// jiraEvents := findJiraEvents(ghEvents.FindReferenceIssueIds())

	allEvents := EventStream{}

	for _, e := range forPR {
		allEvents = append(allEvents, e)
	}

	for _, e := range relatedJiraEvents {
		allEvents = append(allEvents, e)
	}

	slices.SortFunc(allEvents, func (a AggregateEvent[EventPayload], b AggregateEvent[EventPayload]) int {
		return a.occurredAt.Compare(b.occurredAt)
	})
	
	for _,e := range allEvents {
		fmt.Printf("[%s] %s@%s\n\n", e.AggregateId(), e.Type(), e.occurredAt.Format(time.RFC3339))
	}
}