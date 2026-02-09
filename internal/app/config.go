package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jlis/aws-mfa-go/internal/credentials"
)

// Inputs are the raw values coming from the CLI layer.
// Changed flags indicate whether a user explicitly set a flag (used for precedence).
type Inputs struct {
	Profile        string
	ProfileChanged bool

	Device        string
	DeviceChanged bool

	DurationSeconds        int
	DurationSecondsChanged bool

	Token        string
	TokenChanged bool

	Force bool

	LongTermSuffix  string
	ShortTermSuffix string

	CredentialsFile string
}

type Env interface {
	Get(key string) string
}

type Resolved struct {
	Profile          string
	LongTermSection  string
	ShortTermSection string

	Device          string
	DurationSeconds int32
	Token           string
	Force           bool

	CredentialsFile string
}

func Resolve(ctx context.Context, in Inputs, env Env, store *credentials.Store) (Resolved, error) {
	_ = ctx // reserved for future (e.g. tracing); keep signature stable for tests.

	profile := ""
	if in.ProfileChanged && strings.TrimSpace(in.Profile) != "" {
		profile = strings.TrimSpace(in.Profile)
	} else if v := strings.TrimSpace(env.Get("AWS_PROFILE")); v != "" {
		profile = v
	} else {
		profile = "default"
	}

	names, err := credentials.ComputeSectionNames(profile, in.LongTermSuffix, in.ShortTermSuffix)
	if err != nil {
		return Resolved{}, err
	}

	device := ""
	if in.DeviceChanged && strings.TrimSpace(in.Device) != "" {
		device = strings.TrimSpace(in.Device)
	} else if v := strings.TrimSpace(env.Get("MFA_DEVICE")); v != "" {
		device = v
	} else if v, ok := store.Get(names.LongTerm, "aws_mfa_device"); ok && v != "" {
		device = v
	} else {
		return Resolved{}, errors.New("missing MFA device: set --device, MFA_DEVICE, or aws_mfa_device in long-term credentials section")
	}

	duration := int32(0)
	if in.DurationSecondsChanged && in.DurationSeconds > 0 {
		duration = int32(in.DurationSeconds)
	} else if v := strings.TrimSpace(env.Get("MFA_STS_DURATION")); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed <= 0 {
			return Resolved{}, fmt.Errorf("invalid MFA_STS_DURATION %q", v)
		}
		duration = int32(parsed)
	} else {
		duration = 43200 // 12 hours (upstream default without assume-role)
	}

	token := ""
	if in.TokenChanged && strings.TrimSpace(in.Token) != "" {
		token = strings.TrimSpace(in.Token)
	}

	return Resolved{
		Profile:          profile,
		LongTermSection:  names.LongTerm,
		ShortTermSection: names.ShortTerm,
		Device:           device,
		DurationSeconds:  duration,
		Token:            token,
		Force:            in.Force,
		CredentialsFile:  in.CredentialsFile,
	}, nil
}
