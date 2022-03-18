package internal_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/dotnet-execute/cmd/port-chooser/internal"
	"github.com/paketo-buildpacks/occam"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPortChooser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it.Before(func() {
		Expect(os.Unsetenv("PORT")).NotTo(HaveOccurred())
		Expect(os.Unsetenv("ASPNETCORE_URLS")).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.Unsetenv("PORT")).NotTo(HaveOccurred())
		Expect(os.Unsetenv("ASPNETCORE_URLS")).NotTo(HaveOccurred())
	})

	context(`when ASPNETCORE_URLS is not set`, func() {
		context(`when PORT is not set`, func() {
			it(`will set ASPNETCORE_URLS to http://0.0.0.0:8080`, func() {
				envVars, err := internal.ChoosePort()
				Expect(err).NotTo(HaveOccurred())
				Expect(envVars).To(Equal(map[string]string{
					"ASPNETCORE_URLS": "http://0.0.0.0:8080",
				}))
			})
		})

		context(`when PORT is set to a valid number`, func() {
			it.Before(func() {
				Expect(os.Setenv("PORT", "9876")).NotTo(HaveOccurred())
			})

			it(`will set ASPNETCORE_URLS to http://0.0.0.0:9876`, func() {
				envVars, err := internal.ChoosePort()
				Expect(err).NotTo(HaveOccurred())
				Expect(envVars).To(Equal(map[string]string{
					"ASPNETCORE_URLS": "http://0.0.0.0:9876",
				}))
			})
		})

		context(`when PORT is set to a invalid string`, func() {
			it.Before(func() {
				Expect(os.Setenv("PORT", "hi")).NotTo(HaveOccurred())
			})

			it(`will set ASPNETCORE_URLS to http://0.0.0.0:8080`, func() {
				envVars, err := internal.ChoosePort()
				Expect(err).NotTo(HaveOccurred())
				Expect(envVars).To(Equal(map[string]string{
					"ASPNETCORE_URLS": "http://0.0.0.0:8080",
				}))
			})
		})
	})

	context(`when ASPNETCORE_URLS is set`, func() {
		var (
			aspNetCoreUrl string
		)

		it.Before(func() {
			var err error
			aspNetCoreUrl, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			Expect(os.Setenv("ASPNETCORE_URLS", aspNetCoreUrl)).NotTo(HaveOccurred())
		})

		it(`will leave ASPNETCORE_URLS in place`, func() {
			envVars, err := internal.ChoosePort()
			Expect(err).NotTo(HaveOccurred())
			Expect(envVars).To(Equal(map[string]string{}))
			Expect(os.Getenv("ASPNETCORE_URLS")).To(Equal(aspNetCoreUrl))
		})
	})
}
