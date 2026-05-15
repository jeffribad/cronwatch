package notify

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// mockSNSPublisher implements snsPublisher for testing.
type mockSNSPublisher struct {
	publishFn func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	lastInput *sns.PublishInput
}

func (m *mockSNSPublisher) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	m.lastInput = params
	if m.publishFn != nil {
		return m.publishFn(ctx, params, optFns...)
	}
	return &sns.PublishOutput{}, nil
}

func TestSNSNotifier_Send_Success(t *testing.T) {
	mock := &mockSNSPublisher{}
	n := newSNSNotifierWithClient(mock, "arn:aws:sns:us-east-1:123456789012:alerts", "")

	if err := n.Send("backup-job", "job failed after 3 retries"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.lastInput == nil {
		t.Fatal("expected Publish to be called")
	}
	if *mock.lastInput.TopicArn != "arn:aws:sns:us-east-1:123456789012:alerts" {
		t.Errorf("unexpected topic ARN: %s", *mock.lastInput.TopicArn)
	}
}

func TestSNSNotifier_Send_PayloadContents(t *testing.T) {
	mock := &mockSNSPublisher{}
	n := newSNSNotifierWithClient(mock, "arn:aws:sns:us-east-1:123456789012:alerts", "custom subject")

	if err := n.Send("db-backup", "exit code 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := *mock.lastInput.Message
	if !strings.Contains(body, "db-backup") {
		t.Errorf("expected body to contain job name, got: %s", body)
	}
	if !strings.Contains(body, "exit code 1") {
		t.Errorf("expected body to contain message, got: %s", body)
	}
	if *mock.lastInput.Subject != "custom subject" {
		t.Errorf("expected custom subject, got: %s", *mock.lastInput.Subject)
	}
}

func TestSNSNotifier_Send_DefaultSubject(t *testing.T) {
	mock := &mockSNSPublisher{}
	n := newSNSNotifierWithClient(mock, "arn:aws:sns:us-east-1:123456789012:alerts", "")

	if err := n.Send("nightly-report", "missed run"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(*mock.lastInput.Subject, "nightly-report") {
		t.Errorf("expected default subject to contain job name, got: %s", *mock.lastInput.Subject)
	}
}

func TestSNSNotifier_Send_PublishError(t *testing.T) {
	mock := &mockSNSPublisher{
		publishFn: func(_ context.Context, _ *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			return nil, errors.New("network timeout")
		},
	}
	n := newSNSNotifierWithClient(mock, "arn:aws:sns:us-east-1:123456789012:alerts", "")

	err := n.Send("cleanup", "disk full")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "sns: publish failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewSNSNotifier_EmptyARN(t *testing.T) {
	_, err := NewSNSNotifier("", "subject", "us-east-1")
	if err == nil {
		t.Fatal("expected error for empty ARN, got nil")
	}
}
