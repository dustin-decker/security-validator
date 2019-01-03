package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"io/ioutil"

	"encoding/json"

	"github.com/dustin-decker/security-validator/pkg/validator"
	"github.com/dustin-decker/security-validator/pkg/violation"
	"github.com/golang/glog"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

func writeAdmissionError(w http.ResponseWriter, ar v1beta1.AdmissionReview, e error) {
	w.WriteHeader(http.StatusBadRequest)
	ar.Response.Result.Message = e.Error()
	payload, _ := json.Marshal(ar)
	w.Write(payload)
}

// ValidatingWebhook is a ValidatingWebhook endpoint that accepts K8s resources to process
func ValidatingWebhook(w http.ResponseWriter, r *http.Request) {

	ar := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: false,
		},
	}

	// require application/json content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		writeAdmissionError(w, ar, errors.New("incorrect content type"))
	}

	// set the response content-type
	w.Header().Set("Content-Type", "application/json")

	// safely read the body into memory
	body, err := ioutil.ReadAll(r.Body)
	// body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1*10^6))
	if err != nil {
		glog.Error("error reading body: ", err)
		writeAdmissionError(w, ar, err)
	}

	fmt.Println(string(body))

	// unmarshall review request
	deserializer := codecs.UniversalDeserializer()
	if _, _, err = deserializer.Decode(body, nil, &ar); err != nil {
		glog.Error("error unmarshall review request: ", err)
		writeAdmissionError(w, ar, err)
	}
	ar.Response.Allowed = false
	ar.Response.UID = ar.Request.UID

	// validate the resources
	// this slips any results into the review structure
	ar = validateResources(ar)

	// write the review and results JSON response
	payload, err := json.Marshal(ar)
	if err != nil {
		fmt.Println(err)
		writeAdmissionError(w, ar, err)
	}
	w.Write(payload)
}

// validateResources accepts K8s resources to process
func validateResources(ar v1beta1.AdmissionReview) v1beta1.AdmissionReview {
	admitResponse := &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
		},
	}

	// handle Pods
	if ar.Request.Kind.Kind == "Pod" {
		pod := v1.Pod{}
		err := json.Unmarshal(ar.Request.Object.Raw, &pod)
		if err != nil {
			fmt.Println(err)
			admitResponse.Response.Allowed = false
			admitResponse.Response.Result = &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("Error unmarshalling Pod request: %s", err.Error()),
			}
		}

		podViolations := []violation.PodViolation{}

		podViolations = append(podViolations, validator.ValidateImageImmutableReference(pod.Spec)...)

		for _, v := range podViolations {
			fmt.Println(fmt.Sprintf("%s: %s", v.PodName, v.Violation))
		}

		if len(podViolations) > 0 {
			fmt.Println("do something beause pod violations")
			ar.Response.Allowed = false
			ar.Response.Result = &metav1.Status{
				Message: fmt.Sprintf("rejected pod %s because: %s", pod.Name, podViolations[0])
			}
		}
	}

	fmt.Println(ar)
	return ar
}
