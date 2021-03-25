package json

import (
	"fmt"

	"github.com/anchore/grype/internal/config"

	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/vulnerability"
	"github.com/anchore/grype/internal"
	"github.com/anchore/grype/internal/version"
)

// Document represents the JSON document to be presented
type Document struct {
	Matches    []Match      `json:"matches"`
	Source     *source      `json:"source"`
	Distro     distribution `json:"distro"`
	Descriptor descriptor   `json:"descriptor"`
}

// Match is a single item for the JSON array reported
type Match struct {
	Vulnerability Vulnerability `json:"vulnerability"`
	MatchDetails  MatchDetails  `json:"matchDetails"`
	Artifact      Package       `json:"artifact"`
}

// MatchDetails contains all data that indicates how the result match was found
type MatchDetails struct {
	Matcher   string                 `json:"matcher"`
	SearchKey map[string]interface{} `json:"searchKey"`
	MatchInfo map[string]interface{} `json:"matchedOn"`
}

// NewDocument creates and populates a new Document struct, representing the populated JSON document.
func NewDocument(packages []pkg.Package, context pkg.Context, matches match.Matches, metadataProvider vulnerability.MetadataProvider, appConfig config.Application) (Document, error) {
	// we must preallocate the findings to ensure the JSON document does not show "null" when no matches are found
	var findings = make([]Match, 0)
	for m := range matches.Enumerate() {
		p := pkg.ByID(m.Package.ID(), packages)
		if p == nil {
			return Document{}, fmt.Errorf("unable to find package in collection: %+v", p)
		}

		metadata, err := metadataProvider.GetMetadata(m.Vulnerability.ID, m.Vulnerability.RecordSource)
		if err != nil {
			return Document{}, fmt.Errorf("unable to fetch vuln=%q metadata: %+v", m.Vulnerability.ID, err)
		}

		findings = append(
			findings,
			Match{
				Vulnerability: NewVulnerability(m, metadata),
				Artifact:      newPackage(*p),
				MatchDetails: MatchDetails{
					Matcher:   m.Matcher.String(),
					SearchKey: m.SearchKey,
					MatchInfo: m.SearchMatches,
				},
			},
		)
	}

	var src *source
	if context.Source != nil {
		theSrc, err := newSource(*context.Source)
		if err != nil {
			return Document{}, err
		}
		src = &theSrc
	}

	return Document{
		Matches: findings,
		Source:  src,
		Distro:  newDistribution(context.Distro),
		Descriptor: descriptor{
			Name:          internal.ApplicationName,
			Version:       version.FromBuild().Version,
			Configuration: appConfig,
		},
	}, nil
}
