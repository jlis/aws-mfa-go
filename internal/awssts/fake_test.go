package awssts

import (
	"context"
	"testing"
	"time"
)

// fakeClient is a small in-package fake used by unit tests in other packages.
type fakeClient struct {
	gotInputs []GetSessionTokenInput
	out       GetSessionTokenOutput
	err       error
}

func (f *fakeClient) GetSessionToken(ctx context.Context, in GetSessionTokenInput) (GetSessionTokenOutput, error) {
	f.gotInputs = append(f.gotInputs, in)
	return f.out, f.err
}

func TestFakeClientRecordsInputs(t *testing.T) {
	f := &fakeClient{
		out: GetSessionTokenOutput{
			AccessKeyID:     "ASIA_TEST",
			SecretAccessKey: "SECRET",
			SessionToken:    "TOKEN",
			Expiration:      time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC),
		},
	}

	in := GetSessionTokenInput{
		SerialNumber:    "arn:aws:iam::123456789012:mfa/me",
		TokenCode:       "123456",
		DurationSeconds: 3600,
	}

	out, err := f.GetSessionToken(context.Background(), in)
	if err != nil {
		t.Fatalf("GetSessionToken: %v", err)
	}
	if out.AccessKeyID != "ASIA_TEST" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if len(f.gotInputs) != 1 {
		t.Fatalf("expected 1 input recorded, got %d", len(f.gotInputs))
	}
	if f.gotInputs[0] != in {
		t.Fatalf("expected recorded input %+v, got %+v", in, f.gotInputs[0])
	}
}
