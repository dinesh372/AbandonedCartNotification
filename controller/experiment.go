package controller

import (
	"cmp"
	"quince/entity"
	"slices"
)

type ExperimentController interface {
	GetExperimentGroup(userID string) entity.ExperimentAPIResult
}

type experimentController struct {
	Experiment entity.Experiment
}

func NewExperimentController() ExperimentController {
	experiment := entity.Experiment{
		ExperimentName: "test_experiment",
		Segments: []entity.Segment{
			{
				SegmentID: "segment-1",
				Priority:  1,
				Constraints: entity.Constraints{
					TreatmentUsers: []string{
						"user-1",
					},
				},
				Result: entity.SegmentResult{
					Group:     entity.Treatment,
					ContentId: "content-id-1",
				},
			},
			{
				SegmentID: "segment-2",
				Priority:  2,
				Constraints: entity.Constraints{
					TreatmentUsers: []string{
						"user-2",
					},
				},
				Result: entity.SegmentResult{
					Group:     entity.Control,
					ContentId: "invalid",
				},
			},
		},
		DefaultValue: entity.SegmentResult{
			Group:     entity.Control,
			ContentId: "invalid",
		},
	}

	// sort by priority
	slices.SortFunc(experiment.Segments, func(a, b entity.Segment) int {
		return cmp.Compare(a.Priority, b.Priority)
	})

	return &experimentController{
		Experiment: experiment,
	}
}

func (e *experimentController) GetExperimentGroup(userID string) entity.ExperimentAPIResult {

	var matchingSegment *entity.Segment
	for _, segment := range e.Experiment.Segments {
		treatmentUsers := segment.Constraints.TreatmentUsers
		if slices.Contains(treatmentUsers, userID) {
			matchingSegment = &segment
			break
		}
	}

	if matchingSegment == nil {
		matchingSegment = &entity.Segment{
			SegmentID: "default_value",
			Result:    e.Experiment.DefaultValue,
		}
	}

	return entity.ExperimentAPIResult{
		SegmentId: matchingSegment.SegmentID,
		Result:    matchingSegment.Result,
	}
}
