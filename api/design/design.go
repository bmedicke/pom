package design

import . "goa.design/goa/v3/dsl"

var _ = API("pom", func() {
	Title("pom")
	Description("pom's web interface")
	Server("pom", func() {
		Host("localhost", func() { URI("http://localhost:8421") })
	})
})

var _ = Service("pom", func() {
	Method("state", func() {
		Result(String)
		HTTP(func() {
			GET("/state")
			Response(StatusOK)
		})
	})
})
