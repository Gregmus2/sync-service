package main

import (
	core "github.com/GregmusCo/poll-play-golang-core"
	"github.com/GregmusCo/poll-play-template/internal/common"
	"github.com/GregmusCo/poll-play-template/internal/presenters"
	"go.uber.org/fx"
)

func main() {
	core.Serve(
		[]core.Server{
			{
				Services: []core.Service{
					{ServiceDesc: private.TemplateService_ServiceDesc, Constructor: presenters.NewAPI},
				},
				Interceptors: []interceptors.Interceptor{
					&interceptors.ErrorHandlingInterceptor{},
					&interceptors.RequestValidationInterceptor{},
				},
			},
			/*{
				Services: []core.Service{
					{ServiceDesc: public.UserService_ServiceDesc, Constructor: presenters.NewPublic},
				},
				Interceptors: []interceptors.Interceptor{
					&interceptors.AuthorizationInterceptor{},
					&interceptors.ErrorHandlingInterceptor{},
					&interceptors.RequestValidationInterceptor{},
				},
				Port: ":9000",
				Stream: true,
			},*/
		},
		fx.Provide(
			common.NewConfig,
		),
	)
}
