package entity

type SegmentIdentifier string

type Group string

const (
	Control   Group = "Control"
	Treatment Group = "Treatment"
)

type Segment struct {
	SegmentID SegmentIdentifier
	Priority  int
	// A very simple constraint
	Constraints Constraints
	Result      SegmentResult
}

type Constraints struct {
	// A very simple constraint which check for userId equality.
	TreatmentUsers []string
}

type SegmentResult struct {
	Group     Group
	ContentId string
}

type Experiment struct {
	ExperimentName string
	Segments       []Segment
	DefaultValue   SegmentResult
}

type ExperimentAPIResult struct {
	SegmentId SegmentIdentifier
	Result    SegmentResult
}
