// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Route - used to route REST requests received by the service.
type Route struct {
	Name        string           // Name of the route
	Method      string           // REST method
	Path        string           // Resource path
	HandlerFunc http.HandlerFunc // Request handler function.
}

type routes []Route

// List of Routes and corresponding handler functions registered
// with the router.
var registeredRoutes = routes{
	// Health method.
	Route{
		Name:        "GetHealth",
		Method:      http.MethodGet,
		Path:        "/health",
		HandlerFunc: getHealthHandler,
	},

	// Metrics method.
	Route{
		Name:        "GetMetrics",
		Method:      http.MethodGet,
		Path:        "/metrics",
		HandlerFunc: promhttp.Handler().(http.HandlerFunc),
	},

	///////////////////////////////////////////////////////////////////////////
	//                   External API routes (device facing)                 //
	///////////////////////////////////////////////////////////////////////////
	Route{
		Name:        "EnrollDevice",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/enroll", apiUrlPrefix),
		HandlerFunc: esHandlerFunc(Enroll),
	},

	Route{
		Name:        "GetEnrollmentStatus",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/enroll/{enroll_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(EnrollStatus),
	},

	Route{
		Name:        "UnenrollDevice",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/enroll/{device_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(Unenroll),
	},

	Route{
		Name:        "ReEnrollDevice",
		Method:      http.MethodPatch,
		Path:        fmt.Sprintf("%s/enroll/{device_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(RenewEnroll),
	},

	///////////////////////////////////////////////////////////////////////////
	//                   Admin apis
	///////////////////////////////////////////////////////////////////////////
	Route{
		Name:        "CreateEnrollToken",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/enroll_token", apiUrlPrefix),
		HandlerFunc: esHandlerFunc(CreateEnrollToken),
	},

	Route{
		Name:        "DeleteEnrollToken",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/enroll_token", apiUrlPrefix),
		HandlerFunc: esHandlerFunc(DeleteEnrollToken),
	},

	Route{
		Name:        "CreatePolicy",
		Method:      http.MethodPost,
		Path:        fmt.Sprintf("%s/policy", apiUrlPrefix),
		HandlerFunc: esHandlerFunc(CreatePolicy),
	},

	Route{
		Name:        "GetPolicyInfo",
		Method:      http.MethodHead,
		Path:        fmt.Sprintf("%s/policy", apiUrlPrefix),
		HandlerFunc: esHandlerFunc(GetPolicyInfo),
	},

	Route{
		Name:        "GetPolicy",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/policy/{policy_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(GetPolicy),
	},

	Route{
		Name:        "UpdatePolicy",
		Method:      http.MethodPatch,
		Path:        fmt.Sprintf("%s/policy/{policy_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(UpdatePolicy),
	},

	Route{
		Name:        "DeletePolicy",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/policy/{policy_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(DeletePolicy),
	},

	///////////////////////////////////////////////////////////////////////////
	//                   App token API routes                                //
	///////////////////////////////////////////////////////////////////////////
	Route{
		Name:        "GetEnrollToken",
		Method:      http.MethodGet,
		Path:        fmt.Sprintf("%s/enroll_token/{tenant_id:%s}", apiUrlPrefix, uuidRegex),
		HandlerFunc: esHandlerFunc(GetEnrollToken),
	},

	///////////////////////////////////////////////////////////////////////////
	//                   Internal maintenance API routes                     //
	///////////////////////////////////////////////////////////////////////////
	Route{
		Name:        "DeleteExpiredEnroll",
		Method:      http.MethodDelete,
		Path:        fmt.Sprintf("%s/enroll/expired", apiInternalPrefix),
		HandlerFunc: esHandlerFunc(DeleteExpiredEnrolls),
	},
}
