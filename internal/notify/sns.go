package notify

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// snsPublisher abstracts the SNS Publish call for testing.
type snsPublisher interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// SNSNotifier sends alerts to an AWS SNS topic.
type SNSNotifier struct {
	client   snsPublisher
	topicARN string
	subject  string
}

// NewSNSNotifier creates an SNSNotifier using the default AWS credential chain.
// topicARN is the full ARN of the SNS topic, subject is the message subject.
func NewSNSNotifier(topicARN, subject, region string) (*SNSNotifier, error) {
	if topicARN == "" {
		return nil, fmt.Errorf("sns: topic ARN must not be empty")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("sns: failed to load AWS config: %w", err)
	}

	return &SNSNotifier{
		client:   sns.NewFromConfig(cfg),
		topicARN: topicARN,
		subject:  subject,
	}, nil
}

// newSNSNotifierWithClient creates an SNSNotifier with a custom publisher (for testing).
func newSNSNotifierWithClient(client snsPublisher, topicARN, subject string) *SNSNotifier {
	return &SNSNotifier{
		client:   client,
		topicARN: topicARN,
		subject:  subject,
	}
}

// Send publishes the alert message to the configured SNS topic.
func (n *SNSNotifier) Send(jobName, message string) error {
	subject := n.subject
	if subject == "" {
		subject = fmt.Sprintf("cronwatch alert: %s", jobName)
	}

	body := fmt.Sprintf("Job: %s\n\n%s", jobName, message)

	_, err := n.client.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(n.topicARN),
		Subject:  aws.String(subject),
		Message:  aws.String(body),
	})
	if err != nil {
		return fmt.Errorf("sns: publish failed: %w", err)
	}
	return nil
}
