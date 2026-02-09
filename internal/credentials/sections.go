package credentials

import (
	"errors"
	"fmt"
	"strings"
)

type SectionNames struct {
	LongTerm  string
	ShortTerm string
}

// ComputeSectionNames implements the upstream aws-mfa naming rules.
//
// - Long-term default: <profile>-long-term
// - Long-term suffix "none": <profile>
// - Short-term default / "none": <profile>
// - Short-term suffix set: <profile>-<suffix>
func ComputeSectionNames(profile, longTermSuffix, shortTermSuffix string) (SectionNames, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return SectionNames{}, errors.New("profile is empty")
	}

	longTermSuffix = strings.TrimSpace(longTermSuffix)
	shortTermSuffix = strings.TrimSpace(shortTermSuffix)

	longName := ""
	switch {
	case longTermSuffix == "":
		longName = fmt.Sprintf("%s-long-term", profile)
	case strings.EqualFold(longTermSuffix, "none"):
		longName = profile
	default:
		longName = fmt.Sprintf("%s-%s", profile, longTermSuffix)
	}

	shortName := ""
	switch {
	case shortTermSuffix == "":
		shortName = profile
	case strings.EqualFold(shortTermSuffix, "none"):
		shortName = profile
	default:
		shortName = fmt.Sprintf("%s-%s", profile, shortTermSuffix)
	}

	if longName == shortName {
		return SectionNames{}, fmt.Errorf("long-term section name %q equals short-term section name %q", longName, shortName)
	}

	return SectionNames{LongTerm: longName, ShortTerm: shortName}, nil
}
