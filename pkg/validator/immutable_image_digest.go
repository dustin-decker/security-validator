package validator

import (
	"strings"

	"github.com/dustin-decker/security-validator/pkg/violation"
	digest "github.com/opencontainers/go-digest"
	corev1 "k8s.io/api/core/v1"
)

func ValidateImageImmutableReference(p corev1.PodSpec) []violation.PodViolation {

	podViolations := []violation.PodViolation{}

	violationText := "image name does not include a valid digest"

	for _, container := range p.Containers {

		// validate that the image name ends with a digest
		refSplit := strings.Split(container.Image, "@")
		if len(refSplit) == 2 {
			d, _ := digest.Parse(refSplit[len(refSplit)-1])
			err := d.Validate()
			if err != nil {
				podViolations = append(podViolations, violation.PodViolation{
					PodName:   container.Name,
					Violation: violationText,
					Error:     err,
				})
			}
		} else {
			podViolations = append(podViolations, violation.PodViolation{
				PodName:   container.Name,
				Violation: violationText,
				Error:     nil,
			})
		}
	}

	return podViolations
}
